package tui

import tea "charm.land/bubbletea/v2"

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case connectedMsg:
		switch msg.Type {
		case "ws_connected":
			m.wsServerConnected = true
		case "ws_disconnected":
			m.wsServerConnected = false
			m.webRTCConnected = false
			m.availablePeers = []string{}
		case "webrtc_connected":
			m.webRTCConnected = true
		case "webrtc_disconnected":
			m.webRTCConnected = false
		}
	case peersMsg:
		m.availablePeers = msg
	}
	return m, nil
}
