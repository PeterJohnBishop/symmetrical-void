package tui

import "fmt"

func (m Model) View() string {
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
	return view
}
