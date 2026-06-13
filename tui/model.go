package tui

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/peterjohnbishop/symmetrical-void/p2p"
	"github.com/peterjohnbishop/symmetrical-void/wsclient"
)

type Model struct {
	wsConnectionManager *wsclient.ConnectionManager
	webRTCManager       *p2p.WebRTCManager
	wsServerConnected   bool
	webRTCConnected     bool
	availablePeers      []string
	cursor              int
	err                 error
	logs                []string
	activePeer          string
	localFrame          string
	remoteFrame         string
}

// InitialModel creates and returns the initial TUI model with the provided WebSocket connection manager and WebRTC manager.
func InitialModel(wsConnManager *wsclient.ConnectionManager, webRTCManager *p2p.WebRTCManager) Model {
	return Model{
		wsConnectionManager: wsConnManager,
		webRTCManager:       webRTCManager,
		wsServerConnected:   false,
		webRTCConnected:     false,
		availablePeers:      []string{},
		cursor:              0,
		logs:                []string{"App started. Waiting for connections..."},
		activePeer:          "",
	}
}

// Init is the initial command that runs when the TUI starts. It attempts to connect to the WebSocket server and starts listening for messages.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.connectToServer(),
		m.listenForMessages(),
	)
}

// connectToServer attempts to establish a WebSocket connection to the signaling server, initializes the WebRTC manager, and sends a "connect" event message with the device name as the sender.
func (m Model) connectToServer() tea.Cmd {
	return func() tea.Msg {
		ws, err := m.wsConnectionManager.Connect()
		if err != nil {
			return errMsg{fmt.Errorf("could not connect to server: %w", err)}
		}
		m.wsConnectionManager.Conn = ws

		m.webRTCManager.StartWebRTC()
		go m.wsConnectionManager.StartListening()

		deviceName, err := os.Hostname()
		if err == nil {
			m.wsConnectionManager.ID = deviceName
			msg := fmt.Sprintf("[CONN] %s connected to the Websocket Server", deviceName)
			go m.wsConnectionManager.SendEventMessage("connect", msg, nil, nil)
		}

		return connectedMsg{}
	}
}

// listenForMessages is a command that continuously listens for incoming messages from the WebSocket connection and errors, and sends them as TUI messages to be processed in the Update function.
func (m Model) listenForMessages() tea.Cmd {
	return func() tea.Msg {
		select {
		case msg := <-m.wsConnectionManager.MessageChan:
			return msg
		case err := <-m.wsConnectionManager.ErrorChan:
			return errMsg{err}
		case status := <-m.webRTCManager.StatusChan:
			return logMsg(status)
		case frame := <-m.webRTCManager.LocalFrameChan:
			return localFrameMsg(frame)
		case frame := <-m.webRTCManager.RemoteFrameChan:
			return remoteFrameMsg(frame)

		}
	}
}

// helper to check if a peer id already exists
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
