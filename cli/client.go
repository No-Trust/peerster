// Command Line Client for Peerster
package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/No-Trust/peerster/common"
	"github.com/dedis/protobuf"
	"net"
)

func main() {

	UIPort := flag.Uint("UIPort", 10000, "port for the UI client")
	msg := flag.String("msg", "", "message to be sent")
	dest := flag.String("Dest", "", "destination for a private message")
	filename := flag.String("file", "", "file to be indexed")
	request := flag.String("request", "", "metahash of the file to download")
	origin := flag.String("origin", "", "origin of the file to download")
	flag.Parse()

	pkt := common.ClientPacket{}

	if *origin == "" {
		origin = nil
	}

	if *filename != "" {
		if *request != "" {
			// this is a file request

			if *dest == "" {
				fmt.Println("Dest must be given")
				return
			}

			// request
			metahash, err := hex.DecodeString(*request)
			common.CheckError(err)

			fmt.Println("Sending file request")

			pkt.FileRequest = &common.FileRequest{
				MetaHash:    metahash,
				Destination: *dest,
				FileName:    *filename,
				Origin:      origin,
			}

		} else {
			// index a file

			// get absolute path
			// path, err := filepath.Abs(*filename)
			// common.CheckError(err)
			fmt.Println("Sending file indexing request")

			// add file to the message
			pkt.NewFile = &common.NewFile{Path: *filename}
		}
	}

	if *dest == "" && *msg != "" {
		// normal message

		newMessage := common.NewMessage{
			SenderName: "",
			Text:       *msg,
		}

		pkt.NewMessage = &newMessage

	} else if *msg != "" {
		// private message

		newPrivateMessage := common.NewPrivateMessage{
			Origin: "", // putting client name
			Dest:   *dest,
			Text:   *msg,
		}

		pkt.NewPrivateMessage = &newPrivateMessage
	}

	// send message to peer at port peerPort
	ServerAddr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:"+fmt.Sprint(*UIPort))
	common.CheckError(err)

	LocalAddr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:0") // using 0 as port : random unassigned by os
	common.CheckError(err)

	Conn, err := net.DialUDP("udp", LocalAddr, ServerAddr)
	defer Conn.Close()
	common.CheckError(err)

	// sending
	buf, err := protobuf.Encode(&pkt)
	common.CheckError(err)

	_, err = Conn.Write(buf)
	common.CheckError(err)

	// wait if needed

	/*
		if pkt.FileRequest != nil {
			goon := true

			for goon {
				// wait for the notifications
				buf := make([]byte, 65535) // receiving byte array
				_, _, err := Conn.ReadFromUDP(buf)
				if err != nil {
					continue
				}

				var pkt common.ClientPacket
				err = protobuf.Decode(buf, &pkt)
				if err != nil {
					continue
				}

				if pkt.Notification != nil {
					fmt.Println(*pkt.Notification)

					// check if last one
					if (*pkt.Notification)[:13] == "RECONSTRUCTED" {
						goon = false
					}
				}

			}
		}
	*/

}
