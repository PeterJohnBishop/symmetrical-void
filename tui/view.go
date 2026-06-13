package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var (
	leftPanelStyle = lipgloss.NewStyle().
			Width(45).
			PaddingRight(2).
			MarginRight(2).
			BorderStyle(lipgloss.NormalBorder()).
			BorderRight(true).
			BorderForeground(lipgloss.Color("238"))

	localVideoStyle = lipgloss.NewStyle().
			PaddingLeft(2).
			PaddingRight(2).
			MarginRight(2).
			BorderStyle(lipgloss.NormalBorder()).
			BorderRight(true).
			BorderForeground(lipgloss.Color("238"))

	remoteVideoStyle = lipgloss.NewStyle().
				PaddingLeft(2)
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

	leftContent := "Symmetrical Void\n\n"
	leftContent += fmt.Sprintf("WebSocket Server: %s\n", wsStatus)
	leftContent += fmt.Sprintf("WebRTC Connection: %s\n", webrtcStatus)
	leftContent += "\nAvailable Peers:\n"

	if len(m.availablePeers) == 0 {
		leftContent += "  (No peers available)\n"
	} else {
		for i, peer := range m.availablePeers {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}
			if m.webRTCConnected && peer == m.activePeer {
				leftContent += fmt.Sprintf("%s %s ✔\n", cursor, peer)
			} else {
				leftContent += fmt.Sprintf("%s %s\n", cursor, peer)
			}
		}
	}

	if len(m.availablePeers) > 0 {
		leftContent += "\n  [ enter: connect/disconnect • up/down: select • q/ctrl+c: quit ]\n"
	} else {
		leftContent += "\n  [ q/ctrl+c: quit ]\n"
	}

	leftContent += "\n--- Activity Log ---\n"
	for _, logItem := range m.logs {
		leftContent += fmt.Sprintf(" > %s\n", logItem)
	}

	localContent := "--- Local Preview ---\n\n"
	if m.webRTCConnected && m.localFrame != "" {
		localContent += m.localFrame
	} else {
		localContent += "[ Waiting for camera... ]\n"
		localContent += strings.Repeat("\n", 39)
	}

	remoteContent := "--- Remote Peer ---\n\n"
	if m.webRTCConnected && m.remoteFrame != "" {
		remoteContent += m.remoteFrame
	} else {
		remoteContent += "[ Waiting for peer video... ]\n"
		remoteContent += strings.Repeat("\n", 39)
	}

	finalLayout := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPanelStyle.Render(leftContent),
		localVideoStyle.Render(localContent),
		remoteVideoStyle.Render(remoteContent),
	)

	return tea.NewView(finalLayout)
}
