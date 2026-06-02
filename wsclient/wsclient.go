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
}

type EventMessage struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

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
	return conn, nil
}

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

func (c *ConnectionManager) SendEventMessage(eventType string, msgContent string) {
	event := EventMessage{
		Type:    eventType,
		Message: msgContent,
	}

	err := c.Conn.WriteJSON(event)
	if err != nil {
		c.ErrorChan <- err
	}

}
