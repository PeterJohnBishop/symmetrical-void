package wsclient

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/gorilla/websocket"
)

type ConnectionManager struct {
	Conn        *websocket.Conn
	Err         error
	MessageChan chan EventMessage
	ErrorChan   chan error
	ID          string
}

type EventMessage struct {
	Type    string          `json:"type"`
	Sender  string          `json:"sender"`
	Target  string          `json:"target,omitempty"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// NewConnectionManager creates a new ConnectionManager instance and
// attempts to connect to the WebSocket server.
func (c *ConnectionManager) Connect() (*websocket.Conn, error) {
	host := os.Getenv("HOST")
	if host == "" {
		host = "localhost:8080"
	}
	scheme := "wss"
	if strings.HasPrefix(host, "localhost") || strings.HasPrefix(host, "127.0.0.1") {
		scheme = "ws"
	}
	u := url.URL{Scheme: scheme, Host: host, Path: "/ws"}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}

	log.Println("Connected to the websocket server")

	return conn, nil
}

// StartListening continuously reads messages from the WebSocket
// connection and sends them to the MessageChan.
func (c *ConnectionManager) StartListening() {
	defer close(c.MessageChan)

	for {
		_, rawMsg, err := c.Conn.ReadMessage()
		if err != nil {
			c.ErrorChan <- fmt.Errorf("connection closed or read error: %w", err)
			return
		}

		var msg EventMessage
		if err := json.Unmarshal(rawMsg, &msg); err != nil {
			log.Printf("[WARN] Received non-JSON or invalid message: %s", string(rawMsg))
			continue
		}
		c.MessageChan <- msg
	}
}

// SendEventMessage sends a structured event message to the WebSocket server.
func (c *ConnectionManager) SendEventMessage(eventType string, msgContent string, target *string, rawData ...json.RawMessage) {
	var targetVal string
	if target != nil {
		targetVal = *target
	}

	event := EventMessage{
		Type:    eventType,
		Message: msgContent,
		Sender:  c.ID,
		Target:  targetVal,
	}
	if len(rawData) > 0 {
		event.Data = rawData[0]
	}

	err := c.Conn.WriteJSON(event)
	if err != nil {
		select {
		case c.ErrorChan <- err:
		default:
			log.Printf("[ERROR] Failed to send error: %v", err)
		}
	}
}
