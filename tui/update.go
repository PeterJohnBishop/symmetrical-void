package tui

import (
	"encoding/json"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/peterjohnbishop/symmetrical-void/wsclient"
	"github.com/pion/webrtc/v4"
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
				go m.webRTCManager.Disconnect()
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

		case "enter":
			if len(m.availablePeers) == 0 {
				m.logs = append(m.logs, "Enter pressed, but no peers are in the list!")
			} else {
				if m.webRTCConnected {
					go m.webRTCManager.Disconnect()
					m.activePeer = ""
					m.logs = append(m.logs, "Disconnected from peer.")
					return m, nil
				} else {
					target := m.availablePeers[m.cursor]
					m.activePeer = target
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
			m.activePeer = ""
		}
		return m, m.listenForMessages()

	case frameMsg:
		m.currentFrame = string(msg)
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
			if err := json.Unmarshal(msg.Data, &offer); err != nil {
				m.logs = append(m.logs, "Failed to parse incoming offer: "+err.Error())
			} else {
				m.logs = append(m.logs, fmt.Sprintf("Received offer from %s! Generating answer...", msg.Sender))
				m.activePeer = msg.Sender
				go func() {
					err := m.webRTCManager.HandleOffer(msg.Sender, offer.SDP)
					if err != nil {
						m.webRTCManager.StatusChan <- "Handle Offer Error: " + err.Error()
					}
				}()
			}

		case "answer":
			var answer webrtc.SessionDescription
			if err := json.Unmarshal(msg.Data, &answer); err != nil {
				m.logs = append(m.logs, "Failed to parse incoming answer: "+err.Error())
			} else {
				m.logs = append(m.logs, "Received answer! Completing handshake...")

				go func() {
					err := m.webRTCManager.HandleAnswer(answer.SDP)
					if err != nil {
						m.webRTCManager.StatusChan <- "Handle Answer Error: " + err.Error()
					}
				}()
			}

		case "candidate":
			var candidate webrtc.ICECandidateInit
			if err := json.Unmarshal(msg.Data, &candidate); err != nil {
				m.logs = append(m.logs, "Failed to parse ICE candidate: "+err.Error())
			} else {
				m.logs = append(m.logs, "Received ICE candidate from peer.")
				if m.webRTCManager.PC != nil {
					m.webRTCManager.PC.AddICECandidate(candidate)
				}
			}
		}

		return m, m.listenForMessages()
	}

	return m, nil
}
