package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
)

// View renders the TUI view based on the current model state, displaying the connection status of the WebSocket server and WebRTC connection, as well as a list of available peers.
func (m Model) View() tea.View {
	wsStatus := "Disconnected"
	if m.wsServerConnected {
		wsStatus = "Connected"
	}

	webrtcStatus := "Disconnected"
	if m.webRTCConnected {
		webrtcStatus = "Connected"
	}

	view := "Symmetrical Void\n\n"
	view += fmt.Sprintf("WebSocket Server: %s\n", wsStatus)
	view += fmt.Sprintf("WebRTC Connection: %s\n", webrtcStatus)
	view += "\nAvailable Peers:\n"
	for _, peer := range m.availablePeers {
		view += fmt.Sprintf("- %s\n", peer)
	}
	return tea.NewView(view)
}
