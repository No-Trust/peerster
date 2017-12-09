package main

import (
	"io/ioutil"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/No-Trust/peerster/common"
	"github.com/dedis/protobuf"
	"github.com/gorilla/mux"
	"net"
	"net/http"
	"time"
)

var UIPort *uint
var LocalAddr *net.UDPAddr
var ServerAddr *net.UDPAddr
var ServerConn *net.UDPConn
var outputQueue chan *common.ClientPacket
var peers common.PeerSlice
var messages []common.NewMessage
var privateMessages []common.NewPrivateMessage
var ids []string

type WebMessage struct {
	Message     string
	Destination string
	Filename    string
	Hexhash     string
	Node        string
}

func parse(req *http.Request) *WebMessage {
	bytes, err := ioutil.ReadAll(req.Body)
	if common.CheckRead(err) {
		return nil
	}
	err = req.Body.Close()
	if common.CheckRead(err) {
		return nil
	}

	webm := WebMessage{}
	err = json.Unmarshal(bytes, &webm)
	if common.CheckRead(err) {
		return nil
	}

	return &webm
}

func main() {

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
	r.HandleFunc("/file", newFileHandler).Methods("POST")                      // client adds a file
	r.HandleFunc("/download", downloadFileHandler).Methods("POST")             // client request to download a file

	r.HandleFunc("/message", getMessagesHandler).Methods("GET")                // request new messages
	r.HandleFunc("/private-message", getPrivateMessagesHandler).Methods("GET") // request new private messages
	r.HandleFunc("/node", getNodesHandler).Methods("GET")                      // request update on nodes
	r.HandleFunc("/reachable-node", getReachableNodesHandler).Methods("GET")   // request update on reachable nodes

	http.Handle("/", r)

	http.ListenAndServe(":"+fmt.Sprintf("%d", *port), r)
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
	buf, err := json.Marshal(peers.Peers)
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

func newFileHandler(w http.ResponseWriter, r *http.Request) {
	// Assumption : the file is present in the folder of the running webserver
	// Because js security prevent giving absolute filepath

	webm := parse(r)
	if webm == nil {
		return
	}
	filename := webm.Filename
	fmt.Println("*** file : ", filename)

	outputQueue <- &common.ClientPacket{
		NewFile: &common.NewFile{filename},
	}
}

func downloadFileHandler(w http.ResponseWriter, r *http.Request) {
	webm := parse(r)
	if webm == nil {
		return
	}

	hexhash := webm.Hexhash
	destination := webm.Destination
	filename := webm.Filename
	fmt.Println("*** Requesting file :", filename, hexhash)
	// hex -> []byte
	metahash, err := hex.DecodeString(hexhash)
	if err != nil {
		// TODO send notification
		return
	}
	// sending
	outputQueue <- &common.ClientPacket{
		FileRequest: &common.FileRequest{
			MetaHash:    metahash,
			Destination: destination,
			FileName:    filename,
		},
	}
}

func addNodeHandler(w http.ResponseWriter, r *http.Request) {
	webm := parse(r)
	if webm == nil {
		return
	}

	node := webm.Node

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
	fmt.Println(r.Body)
	webm := parse(r)
	fmt.Println(webm)
	if webm == nil {
		return
	}

	msgText := webm.Message
	dest := webm.Destination

	if dest != "" {
		// private message

		fmt.Printf("Sending to %s, %s\n", dest, msgText)

		// sending

		outputQueue <- &common.ClientPacket{
			NewPrivateMessage: &common.NewPrivateMessage{
				Origin: "", // for compatibility
				Dest:   dest,
				Text:   msgText,
			},
		}
	} else {

		fmt.Printf("Sending to all, %s\n", msgText)

		// sending
		outputQueue <- &common.ClientPacket{
			NewMessage: &common.NewMessage{
				SenderName: "", // for compatibility
				Text:       msgText,
			},
		}
	}
}
