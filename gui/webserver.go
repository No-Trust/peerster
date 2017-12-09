package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/dedis/protobuf"
	"github.com/gorilla/mux"
	"github.com/No-Trust/peerster/common"
	"math/big"
	"net"
	"net/http"
	"time"
)

var UIPort *uint
var LocalAddr *net.UDPAddr
var ServerAddr *net.UDPAddr
var ServerConn *net.UDPConn
var outputQueue chan *common.ClientPacket
var name string
var peers common.PeerSlice
var messages []common.NewMessage
var privateMessages []common.NewPrivateMessage
var ids []string

func main() {
	maxIdentifier := big.NewInt(1000000000000)

	nBig, _ := rand.Int(rand.Reader, maxIdentifier)
	n := nBig.Int64()

	name = fmt.Sprintf("%d", n)

	// flags
	UIPort := flag.Uint("UIPort", 10000, "port for the UI client")
	port := flag.Uint("port", 8080, "port for the web server")
	flag.Parse()

	var err error

	ServerAddr, err = net.ResolveUDPAddr("udp4", "127.0.0.1:"+fmt.Sprint(*UIPort))
	common.CheckError(err)

	LocalAddr, err = net.ResolveUDPAddr("udp4", "127.0.0.1:0") // using 0 as port : random unassigned by os
	common.CheckError(err)

	ServerConn, err = net.ListenUDP("udp", LocalAddr)
	//defer ServerConn.Close()
	common.CheckError(err)

	outputQueue = make(chan *common.ClientPacket)

	// listener from server
	go listener(*ServerConn, handleServerMessage)
	// writer to server
	go writer(*ServerConn, outputQueue, ServerAddr)
	// requester
	go requester()

	r := mux.NewRouter()

	// handlers
	r.HandleFunc("/", mainHandler)
	r.HandleFunc("/script.js", jsHandler)
	r.HandleFunc("/style.css", cssHandler)
	r.HandleFunc("/message", sendMessageHandler).Methods("POST")               // client send message
	r.HandleFunc("/node", addNodeHandler).Methods("POST")                      // client add node
	r.HandleFunc("/id", changeNameHandler).Methods("POST")                     // client change id
	r.HandleFunc("/file", newFileHandler).Methods("POST")                      // client adds a file
	r.HandleFunc("/download", downloadFileHandler).Methods("POST")            // client request to download a file
	r.HandleFunc("/message", getMessagesHandler).Methods("GET")                // request new messages
	r.HandleFunc("/private-message", getPrivateMessagesHandler).Methods("GET") // request new private messages
	r.HandleFunc("/node", getNodesHandler).Methods("GET")                      // request update on nodes
	r.HandleFunc("/reachable-node", getReachableNodesHandler).Methods("GET")   // request update on reachable nodes

	http.Handle("/", r)

	http.ListenAndServe(":"+fmt.Sprintf("%d", *port), nil)
}

func requester() {
	// Send a request every second for update
	ticker := time.NewTicker(time.Millisecond * 1000)
	defer ticker.Stop()

	for _ = range ticker.C {
		// send request
		var t bool = true
		outputQueue <- &common.ClientPacket{
			NewMessage:    nil,
			NewNode:       nil,
			RequestUpdate: &t,
		}
	}
}

func handleServerMessage(buf []byte, remoteaddr *net.UDPAddr) {
	var pkt common.ClientPacket
	err := protobuf.Decode(buf, &pkt)
	if common.CheckRead(err) {
		return
	}

	if pkt.NewMessage != nil {
		// Update messages
		messages = append(messages, *pkt.NewMessage)
	}
	if pkt.ReachableNodes != nil {
		// update reachable nodes
		ids = *pkt.ReachableNodes
	}
	if pkt.PeerSlice != nil {
		// update peers
		peers = *pkt.PeerSlice
	}
	if pkt.NewPrivateMessage != nil {
		// update private messages
		privateMessages = append(privateMessages, *pkt.NewPrivateMessage)
	}
	if pkt.Notification != nil {
		// send notification to client
	}

}

// HTTP Handlers

func getNodesHandler(w http.ResponseWriter, r *http.Request) {
	buf, err := json.Marshal(peers)
	common.CheckError(err)
	w.Header().Set("Content-Type", "application/json")
	w.Write(buf)
}

func getMessagesHandler(w http.ResponseWriter, r *http.Request) {
	buf, err := json.Marshal(messages)
	common.CheckError(err)
	w.Header().Set("Content-Type", "application/json")
	w.Write(buf)
}

func getPrivateMessagesHandler(w http.ResponseWriter, r *http.Request) {
	buf, err := json.Marshal(privateMessages)
	common.CheckError(err)
	w.Header().Set("Content-Type", "application/json")
	w.Write(buf)
}

func getReachableNodesHandler(w http.ResponseWriter, r *http.Request) {
	buf, err := json.Marshal(ids)
	common.CheckError(err)
	w.Header().Set("Content-Type", "application/json")
	w.Write(buf)
}

func cssHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "public/style.css")
}
func mainHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "public/")
}

func jsHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "public/script.js")
}

func changeNameHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	newName := r.Form.Get("name")
	name = newName
	fmt.Println("Received name change request : ", name)
	// sending

	outputQueue <- &common.ClientPacket{
		NewName: &name,
	}
}

func newFileHandler(w http.ResponseWriter, r *http.Request) {
	// Assumption : the file is present in the folder of the running webserver
	// Because js security prevent giving absolute filepath
	r.ParseForm()
	filename := r.Form.Get("filename")
	fmt.Println("*** file : ", filename)

	outputQueue <- &common.ClientPacket{
		NewFile: &common.NewFile{filename},
	}
}

func downloadFileHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	hexhash := r.Form.Get("MetaHash")
	destination := r.Form.Get("Destination")
	filename := r.Form.Get("FileName")
	fmt.Println("*** Requesting file :", filename, hexhash)
	// hex -> []byte
	metahash, err := hex.DecodeString(hexhash)
	if err != nil {
		// TODO send notification
		return
	}
	// sending
	outputQueue <- &common.ClientPacket{
		FileRequest: &common.FileRequest {
			MetaHash: metahash,
			Destination: destination,
			FileName: filename,
		},
	}
}

func addNodeHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	node := r.Form.Get("node")

	fmt.Println("Received add node request : ", node)

	nodeAddr, err := net.ResolveUDPAddr("udp4", node)
	if err != nil {
		return
	}

	// sending
	outputQueue <- &common.ClientPacket{
		NewNode: &common.NewNode{
			NewPeer: common.Peer{
				Address:    *nodeAddr,
				Identifier: "",
			},
		},
	}
}

func sendMessageHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	msgText := r.Form.Get("msg")
	dest := r.Form.Get("dest")

	if dest != "" {
		// private message

		fmt.Printf("Sending to %s, %s\n", dest, msgText)

		// sending

		outputQueue <- &common.ClientPacket{
			NewPrivateMessage: &common.NewPrivateMessage{
				Origin: name, // putting client name
				Dest:   dest,
				Text:   msgText,
			},
		}
	} else {

		fmt.Printf("Sending to all, %s\n", msgText)

		// sending
		outputQueue <- &common.ClientPacket{
			NewMessage: &common.NewMessage{
				SenderName: name, // putting client name
				Text:       msgText,
			},
		}
	}
}
