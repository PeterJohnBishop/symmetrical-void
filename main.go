package main

import (
	"log"

	tea "charm.land/bubbletea/v2"
	"github.com/joho/godotenv"
	"github.com/peterjohnbishop/symmetrical-void/p2p"
	"github.com/peterjohnbishop/symmetrical-void/tui"
	"github.com/peterjohnbishop/symmetrical-void/wsclient"
)

// main is the entry point of the application. It initializes the WebSocket connection manager and WebRTC manager,
// creates the TUI model, and starts the Bubble Tea program.
func main() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found, relying on system environment variables")
	}

	c := wsclient.ConnectionManager{
		MessageChan: make(chan wsclient.EventMessage),
		ErrorChan:   make(chan error),
	}

	p := p2p.WebRTCManager{
		WC:              &c,
		StatusChan:      make(chan string, 100),
		LocalFrameChan:  make(chan string, 5),
		RemoteFrameChan: make(chan string, 5),
	}

	m := tui.InitialModel(&c, &p)
	t := tea.NewProgram(m)

	if _, err := t.Run(); err != nil {
		log.Fatal(err)
	}
}
