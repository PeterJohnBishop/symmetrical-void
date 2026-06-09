package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/peterjohnbishop/symmetrical-void/p2p"
	"github.com/peterjohnbishop/symmetrical-void/wsclient"
	"github.com/pion/webrtc/v3"
)

// This is the main entry point for the Symmetrical Void application.
// It initializes the WebSocket connection,
func main() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found, relying on system environment variables")
	}

	c := wsclient.ConnectionManager{
		MessageChan: make(chan wsclient.EventMessage),
		ErrorChan:   make(chan error),
	}

	ws, err := c.Connect()
	if err != nil {
		log.Fatalf("Fatal: Could not connect to server: %v", err)
	}
	c.Conn = ws
	defer c.Conn.Close()

	p := p2p.WebRTCManager{
		WC: &c,
	}

	p.StartWebRTC()

	go c.StartListening()

	var availablePeers []string

	deviceName, err := os.Hostname()
	if err != nil {
		log.Fatalf("Failed to read device name: %v", err)
	}
	msg := fmt.Sprintf("[CONN] %s connected to the Websocket Server", deviceName)
	c.ID = deviceName

	go c.SendEventMessage("connect", msg, nil, nil)

	for {
		select {
		case msg := <-c.MessageChan:
			fmt.Printf("\n[RECV] %s from %s: %s\n> ", msg.Type, msg.Sender, msg.Message)
			switch msg.Type {
			case "connect":
				if msg.Sender != c.ID && msg.Sender != "" {
					availablePeers = append(availablePeers, msg.Sender)
					log.Printf("%s\n", msg.Message)
					log.Printf("Device added: %s. Total peers: %d", msg.Sender, len(availablePeers))
				}
			case "offer":
				var offer webrtc.SessionDescription
				json.Unmarshal(msg.Data, &offer)
				if p.WC == nil {
					p.StartWebRTC()
				}
				p.HandleOffer(msg.Sender, offer.SDP)

			case "answer":
				var answer webrtc.SessionDescription
				json.Unmarshal(msg.Data, &answer)
				p.HandleAnswer(answer.SDP)

			case "candidate":
				var candidate webrtc.ICECandidateInit
				json.Unmarshal(msg.Data, &candidate)
				p.PC.AddICECandidate(candidate)
			}

		case err := <-c.ErrorChan:
			log.Fatalf("\nFatal Network Error: %v", err)
		}
	}
}
