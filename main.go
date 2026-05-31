package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/peterjohnbishop/symmetrical-void/tui"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found, relying on system environment variables")
	}

	host := os.Getenv("SIGNAL_SERVER")
	if host == "" {
		host = "localhost:8080"
	}
	scheme := "wss"
	if strings.HasPrefix(host, "localhost") || strings.HasPrefix(host, "127.0.0.1") {
		scheme = "ws"
	}
	wsUrl := fmt.Sprintf("%s://%s/ws", scheme, host)

	conn, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)
	if err != nil {
		log.Fatalf("Could not connect to signaling server: %v", err)
	}

	initialModel := tui.Model{
		Conn: conn,
		Logs: []string{"[SYSTEM] Connected to " + wsUrl},
	}

	p := tea.NewProgram(initialModel)
	if _, err := p.Run(); err != nil {
		log.Fatalf("Alas, there's been an error: %v", err)
	}
}
