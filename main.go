package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/peterjohnbishop/symmetrical-void/p2p"
	"github.com/peterjohnbishop/symmetrical-void/wsclient"
	"github.com/pion/webrtc/v3"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found, relying on system environment variables")
	}

	c := wsclient.ConnectionManager{
		MessageChan: make(chan wsclient.EventMessage),
		ErrorChan:   make(chan error),
	}

	p := p2p.WebRTCManager{}

	ws, err := c.Connect()
	if err != nil {
		log.Fatalf("Fatal: Could not connect to server: %v", err)
	}
	c.Conn = ws
	defer c.Conn.Close()

	log.Println("Connected to the websocket server")
	log.Println("Type a message and press Enter to broadcast. (Type 'quit' to exit)")

	go c.StartListening()

	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			text, _ := reader.ReadString('\n')
			text = strings.TrimSpace(text)

			if text == "quit" {
				os.Exit(0)
			}

			if text != "" {
				c.SendEventMessage("broadcast", text)
			}
		}
	}()

	for {
		select {
		case msg := <-c.MessageChan:
			fmt.Printf("\n[RECV] %s: %s\n> ", msg.Type, msg.Message)
			switch msg.Type {
			case "offer":
				var offer webrtc.SessionDescription
				json.Unmarshal(msg.Data, &offer)

				p.StartWebRTC(&c) // Make sure engine is running
				p.HandleOffer(&c, offer.SDP)

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
