package main

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
	"github.com/No-Trust/peerster/common"
	"os"
	"path/filepath"
	"sync"
	"time"
)

/***** File Download *****/

// A FileDownload is a data structure containing all required information about a file being downloaded
type FileDownload struct {
	FileMetadata FileMetadata
	Chunks       [][]byte
	NextChunk    uint
	LastChunk    uint
	SigUploader  *[]byte
}

/***** File Downloads *****/

// List of in progress Downloads
type FileDownloads struct {
	downloads map[string]*FileDownload // metahash -> filedownloads
	mutex     *sync.Mutex
}

func NewFileDownloads() *FileDownloads {
	FileDownloads := FileDownloads{
		downloads: make(map[string]*FileDownload),
		mutex:     &sync.Mutex{},
	}
	return &FileDownloads
}

// Get the FileDownload in the list, by its metahash, nil if it does not exist
func (fds *FileDownloads) Get(metahash []byte) *FileDownload {
	fds.mutex.Lock()
	r := fds.downloads[string(metahash)]
	fds.mutex.Unlock()
	return r
}

// Return the chunk coreesponding the the given hash in the FileDownloads list, and nil if it does not exist
func (fds *FileDownloads) GetChunk(hash []byte, hashlen uint) *[]byte {
	fds.mutex.Lock()
	for _, v := range fds.downloads {
		// check for each current FileDownload
		// check if the chunk is inside this download
		pos := v.FileMetadata.GetPositionOfChunk(hash, hashlen)
		if pos != nil {
			// found it
			// return the chunk
			chunk := v.Chunks[*pos]
			fds.mutex.Unlock()
			return &chunk
		}
	}
	// did not find it
	fds.mutex.Unlock()
	return nil
}

func (fds *FileDownloads) Add(f *FileDownload) bool {
	fds.mutex.Lock()
	if fds.downloads[string(f.FileMetadata.Metahash)] != nil {
		// exists
		fmt.Println("Collision of downloads in FileDownloads")
		return false
	}
	fds.downloads[string(f.FileMetadata.Metahash)] = f
	fds.mutex.Unlock()
	return true
}

func (fds *FileDownloads) Remove(f *FileDownload) {
	fds.mutex.Lock()
	fds.downloads[string(f.FileMetadata.Metahash)] = nil
	fds.mutex.Unlock()
}

/***** Download Function *****/

// Perform the download from given information in the FileRequest
func startDownload(g *Gossiper, filereq *common.FileRequest) {

	verifiedUploader := true     // true if the uploader is verified to be the sender
	knownOrigin := true          // true if the origin is known (this peer has its public key)
	validOriginSignature := true // true if the received file is verified to be originated from supposed origin

	uploaderKey, present := g.keyRing.GetKey(filereq.Destination)
	if !present {
		verifiedUploader = false
	}

	// assuming we have the metahash
	metadata := g.metadataSet.Get(filereq.MetaHash)

	var sigUploaderP *[]byte = nil

	if metadata == nil {
		// download metafile

		// build request
		req := DataRequest{
			Origin:      g.Parameters.Identifier,
			Destination: filereq.Destination,
			HopLimit:    g.Parameters.Hoplimit,
			FileName:    filereq.FileName,
			HashValue:   filereq.MetaHash,
		}
		// decrement TTL, drop if less than 0
		req.HopLimit -= 1
		if req.HopLimit <= 0 {
			return
		}

		// get nexthop
		nextHop := g.routingTable.Get(req.Destination)
		if nextHop == "" {
			return
		}
		nextHopAddress := stringToUDPAddr(nextHop)

		// sending
		g.gossipOutputQueue <- &Packet{
			GossipPacket: GossipPacket{
				DataRequest: &req,
			},
			Destination: nextHopAddress,
		}

		// send notification to client
		notification := common.DownloadingMetafileNotification(req.FileName, req.Destination)
		g.clientOutputQueue <- &common.Packet{
			ClientPacket: common.ClientPacket{
				Notification: notification,
			},
			Destination: *g.ClientAddress,
		}
		// print same notification
		g.standardOutputQueue <- notification

		// and wait for data reply
		metaReplyChannel := make(chan *DataReply)

		metaReplyString := string(req.HashValue)

		g.fileWaitersMutex.Lock()
		_, present := g.fileWaiters[metaReplyString]
		g.fileWaitersMutex.Unlock()

		if present {
			// there is a goroutine already waiting for this data
			// too bad
			return
		}

		// Register
		g.fileWaitersMutex.Lock()
		g.fileWaiters[metaReplyString] = metaReplyChannel
		g.fileWaitersMutex.Unlock()

		received := false // not yet received

		for !received {
			timer := time.NewTimer(time.Millisecond * 5000) // timeout = 5sec

			select {
			case <-timer.C:
				// timer stops first
				timer.Stop()
				// send same request again
				g.gossipOutputQueue <- &Packet{
					GossipPacket: GossipPacket{
						DataRequest: &req,
					},
					Destination: nextHopAddress,
				}
				// received == false, therefore, a new timer will be created
			case metareply := <-metaReplyChannel:
				// received the data reply before timeout
				timer.Stop()

				// check integrity
				h := sha256.New()
				h.Write(metareply.Data)
				receivedHash := h.Sum(nil)

				if bytes.Equal(receivedHash, filereq.MetaHash) {
					// We received the correct chunk
					received = true

					metadata = &FileMetadata{
						Name:     filereq.FileName,
						Size:     GetNumberOfChunks(metareply.Data, g.Parameters.HashLength),
						Metahash: receivedHash,
					}

					metafile := metareply.Data
					metadata.Metafile = make([]byte, len(metafile))
					copy(metadata.Metafile, metafile)

					if verifiedUploader && metareply.SigOrigin != nil && metareply.SigUploader != nil && metareply.SigMetaUploader != nil {

						sigoriginI := *metareply.SigOrigin
						SigOrigin := make([]byte, len(sigoriginI))
						copy(SigOrigin, sigoriginI)
						metadata.SigOrigin = &SigOrigin

						siguploader := *metareply.SigUploader
						SigUploader := make([]byte, len(siguploader))
						copy(SigUploader, siguploader)
						sigUploaderP = &SigUploader

						sigmetauploader := *metareply.SigMetaUploader
						SigMetaUploader := make([]byte, len(sigmetauploader))
						copy(SigMetaUploader, sigmetauploader)

						// check that SigMetaUploader corresponds
						metac := append(metadata.Metafile, append(SigOrigin, SigUploader...)...)
						newhash := sha256.New()
						newhash.Write(metac)
						metachashed := newhash.Sum(nil)

						// verify against SigMetaUploader
						err := rsa.VerifyPSS(&uploaderKey, crypto.SHA256, metachashed, SigMetaUploader, nil)
						if err != nil {
							g.standardOutputQueue <- FileWrongSigMetaUploader(filereq.Destination)
							verifiedUploader = false
						} else {
							g.standardOutputQueue <- FileGoodSigMetaUploader(filereq.Destination)
						}
					} else {
						// no signatures, uploader cannot be verified
						verifiedUploader = false
					}

					if metareply.SigOrigin == nil {
						// if no origin signature, cannot verify origin
						validOriginSignature = false
					}

					g.metadataSet.Add(*metadata)

					close(metaReplyChannel)
					g.fileWaitersMutex.Lock()
					g.fileWaiters[metaReplyString] = nil
					g.fileWaitersMutex.Unlock()

				} else {
					// invalid metafile
					continue
				}
			}
		}
	}

	chunkNumber := GetNumberOfChunks(metadata.Metafile, g.Parameters.HashLength)

	download := FileDownload{
		FileMetadata: *metadata,
		Chunks:       make([][]byte, 0),
		NextChunk:    0,
		LastChunk:    chunkNumber,
		SigUploader:  sigUploaderP,
	}

	// add current download info to the database of current downloads
	newDownload := g.FileDownloads.Add(&download)
	if !newDownload {
		return
	}

	for chunkNb := 0; uint(chunkNb) < chunkNumber; chunkNb++ {

		// Downloading chunk # chunkNb

		// get wanted chunk hash
		chunkhash := metadata.GetChunkHash(chunkNb, g.Parameters.HashLength)

		// build the request
		req := DataRequest{
			Origin:      g.Parameters.Identifier,
			Destination: filereq.Destination,
			HopLimit:    g.Parameters.Hoplimit,
			FileName:    filereq.FileName,
			HashValue:   chunkhash,
		}

		// decrement TTL, drop if less than 0
		req.HopLimit -= 1
		if req.HopLimit <= 0 {
			return
		}

		// get nexthop
		nextHop := g.routingTable.Get(req.Destination)
		if nextHop == "" {
			return
		}
		nextHopAddress := stringToUDPAddr(nextHop)

		// sending
		g.gossipOutputQueue <- &Packet{
			GossipPacket: GossipPacket{
				DataRequest: &req,
			},
			Destination: nextHopAddress,
		}

		// send notification to client
		notification := common.DownloadingChunkNotification(req.FileName, req.Destination, chunkNb)
		g.clientOutputQueue <- &common.Packet{
			ClientPacket: common.ClientPacket{
				Notification: notification,
			},
			Destination: *g.ClientAddress,
		}
		// print same notification
		g.standardOutputQueue <- notification

		// and wait for data reply
		dataReplyChannel := make(chan *DataReply)

		dataReplyString := string(req.HashValue)

		g.fileWaitersMutex.Lock()
		_, present := g.fileWaiters[dataReplyString]
		g.fileWaitersMutex.Unlock()

		if present {
			// there is a goroutine already waiting for this data
			// too bad
			return
		}

		// Register
		g.fileWaitersMutex.Lock()
		g.fileWaiters[dataReplyString] = dataReplyChannel
		g.fileWaitersMutex.Unlock()

		received := false // not yet received

		for !received {

			timer := time.NewTimer(time.Millisecond * 5000) // timeout = 5sec

			select {
			case <-timer.C:
				// timer stops first
				timer.Stop()
				// send same request again
				g.gossipOutputQueue <- &Packet{
					GossipPacket: GossipPacket{
						DataRequest: &req,
					},
					Destination: nextHopAddress,
				}
				// received == false, therefore, a new timer will be created
			case reply := <-dataReplyChannel:

				// HERE THE Data replaces the hashvalues

				// received the data reply before timeout
				timer.Stop()

				// check integrity
				h := sha256.New()
				h.Write(reply.Data)
				receivedHash := h.Sum(nil)

				if bytes.Equal(receivedHash, chunkhash) {
					// We received the correct chunk
					received = true

					data := reply.Data
					ndata := make([]byte, len(data))
					copy(ndata, data)

					download.NextChunk += 1
					download.Chunks = append(download.Chunks, ndata)

					close(dataReplyChannel)

					g.fileWaitersMutex.Lock()
					g.fileWaiters[dataReplyString] = nil
					g.fileWaitersMutex.Unlock()
				} else {
					// invalid chunk
					continue
				}
			}

		}
	}

	// got all the chunks

	// reconstruct file
	fileBytes := reassembleChunks(&(download.Chunks))

	// check signatures against metahash
	sigOrigin := download.FileMetadata.SigOrigin
	sigUploader := download.SigUploader
	metahash := download.FileMetadata.Metahash

	// check sigUploader
	if verifiedUploader {
		err := rsa.VerifyPSS(&uploaderKey, crypto.SHA256, metahash, *sigUploader, nil)
		if err != nil {
			g.standardOutputQueue <- FileWrongSigUploader(filereq.Destination)
			verifiedUploader = false
		} else {
			g.standardOutputQueue <- FileGoodSigUploader(filereq.Destination)
		}
	}
	// check sigOrigin
	if sigOrigin != nil && validOriginSignature && filereq.Origin != nil {
		originKey, present := g.keyRing.GetKey(*filereq.Origin)

		if !present {
			// do not have the public key : impossible to verify
			validOriginSignature = false
			knownOrigin = false
		} else {
			err := rsa.VerifyPSS(&originKey, crypto.SHA256, metahash, *sigOrigin, nil)
			if err != nil {
				g.standardOutputQueue <- FileWrongSigOrigin(*filereq.Origin)
				validOriginSignature = false
			} else {
				g.standardOutputQueue <- FileGoodSigOrigin(*filereq.Origin)
			}
		}
	} else {
		validOriginSignature = false
	}

	if verifiedUploader && knownOrigin && !validOriginSignature && filereq.Origin != nil {
		// the file has been sent by Uploader, and its signature by origin is invalid
		// action : decrease reputation of Uploader

		if record, ok := g.keyRing.GetRecord(*filereq.Origin); ok {
			g.reputationTable.DecreaseSigRep(filereq.Destination, record.Confidence)
		}

	} else if verifiedUploader && knownOrigin && validOriginSignature && filereq.Origin != nil {
		// the file has been sent by Uploader, and its signature by origin is valid
		// action : increase reputation of Uploader

		if record, ok := g.keyRing.GetRecord(*filereq.Origin); ok {
			g.reputationTable.IncreaseSigRep(filereq.Destination, record.Confidence)
		}

	}

	if validOriginSignature {
		// the file itself is fine
		g.standardOutputQueue <- FileGoodOrigin(*filereq.Origin)
	} else if filereq.Origin == nil {
		// did not ask for origin check
		g.standardOutputQueue <- FileWarningUnverifiedOrigin()
	} else {
		// the origin of the file cannot be certified
		g.standardOutputQueue <- FileErrorUnverifiedOrigin(*filereq.Origin)
		// drop the file
		g.FileDownloads.Remove(&download)
		return
	}

	// store file in disk
	path, err := filepath.Abs("")
	common.CheckError(err)
	downloadDir := path + string(os.PathSeparator) + g.Parameters.FilesDirectory
	writeToDisk(*fileBytes, downloadDir, filereq.FileName)

	// send notification to client
	notification := common.ReconstructedNotification(filereq.FileName)
	g.clientOutputQueue <- &common.Packet{
		ClientPacket: common.ClientPacket{
			Notification: notification,
		},
		Destination: *g.ClientAddress,
	}
	// print same notification
	g.standardOutputQueue <- notification

	// store chunks in disk
	writeChunksToDisk(download.Chunks, g.Parameters.ChunksDirectory, filereq.FileName)

	// we are done with the download
	g.FileDownloads.Remove(&download)
}
