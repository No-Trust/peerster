// Procedure for incoming data request from other gossipers
package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"github.com/No-Trust/peerster/common"
	"io/ioutil"
	"net"
)

// Handler for inbound data request
// Forward the packet except if it attained the correct destination
// Checks in order if the required file (chunk or metadata) is (if yes, sends the retrieved file)
// 	- is a in-memory metadata
//	- is a in-disk chunk
//	- is a in-memory chunk (corresponding to a current download)
func (g *Gossiper) processDataRequest(req *DataRequest, remoteaddr *net.UDPAddr) {
	// check if this peer is the destination

	if req.Destination == g.Parameters.Identifier {
		// this node is the destination

		// get nexthop
		nexthop := g.routingTable.Get(req.Origin)
		// if no next hop entry : send back to remoteaddr
		nextHopAddress := remoteaddr
		if nexthop != "" {
			hop := stringToUDPAddr(nexthop)
			nextHopAddress = &hop
		}

		hash := req.HashValue

		// check if this is a metafile request
		fm := g.metadataSet.Get(hash)

		if fm != nil {
			// this is a metafile request

			// signing the metahash
			SigUploader, err := rsa.SignPSS(rand.Reader, &(g.key), crypto.SHA256, fm.Metahash, nil)
			common.CheckError(err)

			var SigMetaUploaderP *[]byte = nil
			if fm.SigOrigin != nil {
				metac := append(fm.Metafile, append(*fm.SigOrigin, SigUploader...)...)
				newhash := sha256.New()
				newhash.Write(metac)
				metachashed := newhash.Sum(nil)
				SigMetaUploader, err := rsa.SignPSS(rand.Reader, &(g.key), crypto.SHA256, metachashed, nil)
				common.CheckError(err)
				SigMetaUploaderP = &SigMetaUploader
			}

			// send the metafile

			g.gossipOutputQueue <- &Packet{
				GossipPacket: GossipPacket{
					DataReply: &DataReply{
						Origin:          g.Parameters.Identifier,
						Destination:     req.Origin,
						HopLimit:        g.Parameters.Hoplimit,
						FileName:        req.FileName,
						HashValue:       hash,
						Data:            fm.Metafile,
						SigOrigin:       fm.SigOrigin,
						SigUploader:     &SigUploader,
						SigMetaUploader: SigMetaUploaderP,
					},
				},
				Destination: *nextHopAddress,
			}
			return
		}

		// get the corresponding FileMetadata

		// Check in downloaded chunks
		chunkfilename := GetChunkFilenameFromHash(req.HashValue, g.Parameters.HashLength)
		filepath := g.Parameters.ChunksDirectory + chunkfilename
		chunk, err := ioutil.ReadFile(filepath)

		if err == nil && chunk != nil {
			// got it
			// send it

			g.gossipOutputQueue <- &Packet{
				GossipPacket: GossipPacket{
					DataReply: &DataReply{
						Origin:      g.Parameters.Identifier,
						Destination: req.Origin,
						HopLimit:    g.Parameters.Hoplimit,
						FileName:    req.FileName,
						HashValue:   req.HashValue,
						Data:        chunk,
					},
				},
				Destination: *nextHopAddress,
			}

			return
		}

		// Check in downloading files
		chunkPointer := g.FileDownloads.GetChunk(hash, g.Parameters.HashLength)
		if chunkPointer != nil {
			// got it
			// Send it

			g.gossipOutputQueue <- &Packet{
				GossipPacket: GossipPacket{
					DataReply: &DataReply{
						Origin:      g.Parameters.Identifier,
						Destination: req.Origin,
						HopLimit:    g.Parameters.Hoplimit,
						FileName:    req.FileName,
						HashValue:   hash,
						Data:        *chunkPointer,
					},
				},
				Destination: *nextHopAddress,
			}
			return
		}

		return
	}

	// this is not the destination
	// forward the packet

	if g.Parameters.NoForward {
		return
	}

	// decrement TTL, drop if less than 0
	req.HopLimit -= 1
	if req.HopLimit <= 0 {
		return
	}

	// get nexthop
	nexthop := g.routingTable.Get(req.Destination)
	if nexthop != "" {
		// only forward if we have a route
		nextHopAddress := stringToUDPAddr(nexthop)

		g.gossipOutputQueue <- &Packet{
			GossipPacket: GossipPacket{
				DataRequest: req,
			},
			Destination: nextHopAddress,
		}
	}
	return
}
