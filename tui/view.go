package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
)

// View renders the TUI view based on the current model state, displaying the connection status of the WebSocket server and WebRTC connection, as well as a list of available peers.
func (m Model) View() tea.View {
	if m.cursor >= len(m.availablePeers) && len(m.availablePeers) > 0 {
		m.cursor = len(m.availablePeers) - 1
	}

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

	if len(m.availablePeers) == 0 {
		view += "  (No peers available)\n"
	} else {
		for i, peer := range m.availablePeers {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}
			if m.webRTCConnected && peer == m.activePeer {
				view += fmt.Sprintf("%s %s ✔\n", cursor, peer)
			} else {
				view += fmt.Sprintf("%s %s\n", cursor, peer)
			}
		}
	}

	if len(m.availablePeers) > 0 {
		view += "\n  [ space: connect/disconnect • up/down: select • q/ctrl+c: quit ]\n"
	} else {
		view += "\n  [ q/ctrl+c: quit ]\n"
	}

	view += "\n--- Activity Log ---\n"
	for _, logItem := range m.logs {
		view += fmt.Sprintf(" > %s\n", logItem)
	}

	return tea.NewView(view)
}
