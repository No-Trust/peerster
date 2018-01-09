// Procedure for incoming client messages (client API)
package main

import (
	"crypto/rsa"
	"crypto/sha256"
	"crypto"
	"crypto/rand"
	"github.com/No-Trust/peerster/common"
	"io/ioutil"
	"net"
	"os"
	"log"
	"path/filepath"
)

// New Node : a peer has been added by the user
func processNewNode(nnode *common.NewNode, g *Gossiper) {
	// new node

	newPeer := nnode.NewPeer
	// add the new peer the the peerset
	g.peerSet.Add(newPeer)
	g.standardOutputQueue <- g.peerSet.PeersListString()
}

// New Message : a message has been sent by the user
func processNewMessage(msg *common.NewMessage, g *Gossiper, remoteaddr *net.UDPAddr) {
	// new message
	g.standardOutputQueue <- msg.ClientNewMessageString()
	g.standardOutputQueue <- g.peerSet.PeersListString()

	nextSeq := g.vectorClock.Get(g.Parameters.Identifier)

	//g.Parameters.Name = msg.SenderName
	//g.Parameters.Identifier = msg.SenderName

	// create rumor from message
	rumor := RumorMessage{
		Origin: g.Parameters.Identifier,
		ID:     nextSeq,
		Text:   msg.Text,
	}

	// update status vector
	g.vectorClock.Update(g.Parameters.Identifier)

	// update messages
	g.messages.Add(&rumor)

	// send to ClientPacket
	g.clientOutputQueue <- &common.Packet{
		ClientPacket: common.ClientPacket{
			NewMessage: &common.NewMessage{
				SenderName: rumor.Origin,
				Text:       rumor.Text,
			},
		},
		Destination: *remoteaddr,
	}

	// and send the rumor
	destPeer := g.peerSet.RandomPeer()
	if destPeer != nil {
		go g.rumormonger(&rumor, destPeer)
	}
}

// New Private Message : a private message has been sent by the user
func processNewPrivateMessage(pcm *common.NewPrivateMessage, g *Gossiper) {
	// new private message
	g.standardOutputQueue <- pcm.ClientNewPrivateMessageString()

	// encrypt text with receiver's public key
	rpub, pres := g.keyRing.GetKey(pcm.Dest)
	if !pres {
		// no public key to the destination
		return
	}
	secret := []byte(pcm.Text)
	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, &rpub, secret, nil)
	if err != nil {
		log.Println(err)
		return
	}
	pm := PrivateMessage{
		Origin:   g.Parameters.Identifier,
		ID:       0,
		Text:     string(ciphertext),
		Dest:     pcm.Dest,
		HopLimit: g.Parameters.Hoplimit,
	}

	// check if this peer is the destination
	if pm.Dest == g.Parameters.Identifier {
		// this node is the destination
		g.standardOutputQueue <- pm.PrivateMessageString(&g.Parameters.GossipAddr)
		// send the message to the client
		g.clientOutputQueue <- &common.Packet{
			ClientPacket: common.ClientPacket{
				NewPrivateMessage: &common.NewPrivateMessage{
					Origin: g.Parameters.Identifier,
					Dest:   pm.Dest,
					Text:   pm.Text,
				},
			},
			Destination: *g.ClientAddress,
		}
		// the destination has been reached
		return
	}

	// decrement TTL, drop if less than 0
	pm.HopLimit -= 1
	if pm.HopLimit <= 0 {
		return
	}

	nextHop := g.routingTable.Get(pm.Dest)
	if nextHop == "" {
		return
	}
	nextHopAddress := stringToUDPAddr(nextHop)

	// sending
	g.gossipOutputQueue <- &Packet{
		GossipPacket: GossipPacket{
			Private: &pm,
		},
		Destination: nextHopAddress,
	}

	return
}

// Update request : the client request an update on the peers, messages...
func processRequestUpdate(req *bool, g *Gossiper, remoteaddr *net.UDPAddr) {

	if *(req) == true {

		// Update Request
		cpy := g.peerSet.ToPeerSlice() // copy of the peerset
		ids := g.routingTable.GetIds() // ids of peer with known route
		keys := g.keyRing.GetPeerList() // ids of peers with known public key
		// nasty intersection TODO
		rNodes := make([]string, 0)
		for _, p1 := range ids {
			for _, p2 := range keys {
				if p1 == p2 {
					rNodes = append(rNodes, p2)
					break;
				}
			}
		}

		graph, err := g.keyRing.JSON()
		if err != nil {
			graph = nil
		}

    // Get reputation update
    repUpdate := g.reputationTable.GetUpdate()

    update := common.RepUpdate {
      SigReps     : common.ReputationMap(repUpdate.SigReps),
      ContribReps : common.ReputationMap(repUpdate.ContribReps),
    }

		// sending
		g.clientOutputQueue <- &common.Packet{
			ClientPacket: common.ClientPacket{
				ReachableNodes: &rNodes,
				PeerSlice:      &cpy,
				KeyRingJSON:    &graph,
				Reputations:    &update,
			},
			Destination: *remoteaddr,
		}
	}
}

// New file : the client sends a new file to be indexed
func processNewFile(newfile *common.NewFile, g *Gossiper) {

	g.standardOutputQueue <- newfile.ClientNewFileString()

	filename := filepath.Base(newfile.Path)

	// Read file
	data, err := ioutil.ReadFile(newfile.Path)
	common.CheckRead(err)

	if err != nil {
		// could not read file, stop processing
		return
	}

	filesize := uint(len(data))

	// divide into chunks
	chunks := splitInChunks(data, g.Parameters.ChunkSize)

	// compute hashes
	hashes := hashChunks(chunks)

	// build metafile
	var metafile []byte
	for _, hash := range hashes {
		metafile = append(metafile, hash...)
	}

	// compute metahash
	h := sha256.New()
	h.Write(metafile)
	metahash := h.Sum(nil)

	// signing the metahash
	SigOrigin, err := rsa.SignPSS(rand.Reader, &(g.key), crypto.SHA256, metahash, nil)
	common.CheckError(err)

	meta := FileMetadata{
		Name:      filename,
		Size:      filesize,
		Metafile:  metafile,
		Metahash:  metahash,
		SigOrigin: &SigOrigin,
	}

	g.metadataSet.Add(meta)

	// store file in disk
	path, err := filepath.Abs("")
	common.CheckError(err)

	downloadDir := path + string(os.PathSeparator) + g.Parameters.FilesDirectory
	// store whole file to disk
	writeToDisk(data, downloadDir, filename)

	// store chunks to disk
	writeChunksToDisk(*chunks, g.Parameters.ChunksDirectory, filename)

	g.standardOutputQueue <- FileSubmissionDone(metahash)
}

// File request : the client requests a file to be downloaded
func processFileRequest(filereq *common.FileRequest, g *Gossiper) {
	g.standardOutputQueue <- filereq.ClientNewFileRequestString()

	req := DataRequest{
		Origin:      g.Parameters.Identifier,
		Destination: filereq.Destination,
		HopLimit:    g.Parameters.Hoplimit,
		FileName:    filereq.FileName,
		HashValue:   filereq.MetaHash,
	}

	// check if this peer is the destination
	if filereq.Destination == g.Parameters.Identifier {
		// this node is the destination
		// The client is sending a request for a file on the gossiper it is attached to
		// If the gossiper has the file, then no need to send request to someone else
		// If the gossiper does not have the file, then TODO
		g.standardOutputQueue <- req.DataRequestString(&g.Parameters.GossipAddr)

		return
	}

	// check if already received
	/*
		metadata := g.metadataSet.Get(filereq.MetaHash)
		if metadata != nil {
			// having the metadata != have the file
			filepath := g.Parameters.FilesDirectory + metadata.Name

			if _, err := os.Stat(filepath); !os.IsNotExist(err) {
				// file exists
				fmt.Println("METADATA : ", *metadata)
				fmt.Println("name : ", metadata.Name)
				g.standardOutputQueue <- filereq.Gossi	perAlreadyHasFileString()
				return
			}

			// data, err := ioutil.ReadFile(filepath)
			// if err == nil && data != nil {
			// 	// this peer already has the file
			// 	g.standardOutputQueue <- filereq.GossiperAlreadyHasFileString()
			// 	return
			// }
		}
	*/

	// otherwise, start the download process
	go startDownload(g, filereq)
}
