package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/peterjohnbishop/symmetrical-void/wsclient"
)

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

		case err := <-c.ErrorChan:
			log.Fatalf("\nFatal Network Error: %v", err)
		}
	}
}
