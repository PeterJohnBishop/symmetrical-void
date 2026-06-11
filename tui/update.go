package tui

import (
	"encoding/json"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/peterjohnbishop/symmetrical-void/wsclient"
	"github.com/pion/webrtc/v3"
)

// Update is the main update function for the TUI model. It handles various message types, including key presses, WebSocket connection events, and incoming WebRTC signaling messages, updating the model state accordingly.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			if m.wsConnectionManager.Conn != nil {
				m.wsConnectionManager.Conn.Close()
			}
			if m.webRTCConnected {
				m.webRTCManager.Disconnect()
			}
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.availablePeers)-1 {
				m.cursor++
			}

		case " ", "enter":
			if len(m.availablePeers) == 0 {
				m.logs = append(m.logs, "Spacebar pressed, but no peers are in the list!")
			} else {
				if m.webRTCConnected {
					m.webRTCManager.Disconnect()
				} else {
					target := m.availablePeers[m.cursor]

					m.logs = append(m.logs, fmt.Sprintf("Initiating connection to %s...", target))

					go func() {
						err := m.webRTCManager.SendOffer(target)
						if err != nil {
							m.webRTCManager.StatusChan <- "Offer Error: " + err.Error()
						}
					}()
				}
			}
		}

	case connectedMsg:
		m.wsServerConnected = true
		return m, nil

	case errMsg:
		m.err = msg.err
		return m, tea.Quit
	case logMsg:
		msgStr := string(msg)

		m.logs = append(m.logs, msgStr)
		if len(m.logs) > 5 {
			m.logs = m.logs[1:]
		}

		if msgStr == "Data channel is open!" {
			m.webRTCConnected = true
		}
		if msgStr == "WebRTC connection closed" {
			m.webRTCConnected = false
		}
		return m, m.listenForMessages()

	case wsclient.EventMessage:
		switch msg.Type {
		case "connect":
			if msg.Sender != m.wsConnectionManager.ID && msg.Sender != "" {
				if !contains(m.availablePeers, msg.Sender) {
					m.availablePeers = append(m.availablePeers, msg.Sender)
					target := msg.Sender
					go m.wsConnectionManager.SendEventMessage("presence", "I'm availible for connection", &target, nil)
				}
			}

		case "presence":
			if msg.Sender != m.wsConnectionManager.ID && msg.Sender != "" {
				if !contains(m.availablePeers, msg.Sender) {
					m.availablePeers = append(m.availablePeers, msg.Sender)
				}
			}
		case "offer":
			var offer webrtc.SessionDescription
			json.Unmarshal(msg.Data, &offer)
			if m.webRTCManager.WC == nil {
				m.webRTCManager.StartWebRTC()
			}
			m.webRTCManager.HandleOffer(msg.Sender, offer.SDP)

		case "answer":
			var answer webrtc.SessionDescription
			json.Unmarshal(msg.Data, &answer)
			m.webRTCManager.HandleAnswer(answer.SDP)

		case "candidate":
			var candidate webrtc.ICECandidateInit
			json.Unmarshal(msg.Data, &candidate)
			m.webRTCManager.PC.AddICECandidate(candidate)
		}

		return m, m.listenForMessages()
	}

	return m, nil
}
