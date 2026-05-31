package rtcomm

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

type SignalingMessage struct {
	Type      string                   `json:"type"`
	SDP       string                   `json:"sdp,omitempty"`
	Candidate *webrtc.ICECandidateInit `json:"candidate,omitempty"`
}

type RTCManager struct {
	PC    *webrtc.PeerConnection
	Conn  *websocket.Conn
	LogCh chan string
}

func NewRTCManager(conn *websocket.Conn, logCh chan string) *RTCManager {
	return &RTCManager{
		Conn:  conn,
		LogCh: logCh,
	}
}

func logIt(ch chan string, format string, v ...any) {
	ch <- fmt.Sprintf(format, v...)
}

// initPC initializes a new PeerConnection and its event handlers
func (m *RTCManager) initPC() error {
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{{URLs: []string{"stun:stun.l.google.com:19302"}}},
	}

	pc, err := webrtc.NewPeerConnection(config)
	if err != nil {
		return err
	}
	m.PC = pc

	// forward discovered ICE candidates to the signaling server
	m.PC.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate == nil {
			return
		}
		candidateInit := candidate.ToJSON()
		payload := SignalingMessage{Type: "candidate", Candidate: &candidateInit}
		_ = m.Conn.WriteJSON(payload)
	})

	// notify tui on connection state change
	m.PC.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		logIt(m.LogCh, "WebRTC State: %s", state.String())
	})

	// reciever data channel
	m.PC.OnDataChannel(func(d *webrtc.DataChannel) {
		logIt(m.LogCh, "Remote Data channel '%s' opened", d.Label())
		d.OnMessage(func(msg webrtc.DataChannelMessage) {
			if msg.IsString {
				logIt(m.LogCh, "[RECV] <- %s", string(msg.Data))
			} else {
				logIt(m.LogCh, "[RECV] <- Binary chunk: %d bytes", len(msg.Data))
			}
		})
	})

	return nil
}

// constant websocket listerner
func (m *RTCManager) StartSignalingRouter() {
	for {
		_, rawMsg, err := m.Conn.ReadMessage()
		if err != nil {
			logIt(m.LogCh, "WebSocket closed or read error")
			return
		}

		var msg SignalingMessage
		if err := json.Unmarshal(rawMsg, &msg); err != nil {
			continue
		}

		switch msg.Type {
		case "offer":
			logIt(m.LogCh, "Received Remote Offer. Processing...")
			if m.PC == nil {
				if err := m.initPC(); err != nil {
					logIt(m.LogCh, "Failed to init PC: %v", err)
					continue
				}
			}

			remoteOffer := webrtc.SessionDescription{Type: webrtc.SDPTypeOffer, SDP: msg.SDP}
			if err := m.PC.SetRemoteDescription(remoteOffer); err != nil {
				logIt(m.LogCh, "Failed to set remote description: %v", err)
				continue
			}

			answer, err := m.PC.CreateAnswer(nil)
			if err != nil {
				continue
			}

			if err := m.PC.SetLocalDescription(answer); err != nil {
				continue
			}

			answerPayload := SignalingMessage{Type: "answer", SDP: answer.SDP}
			if err := m.Conn.WriteJSON(answerPayload); err != nil {
				logIt(m.LogCh, "Failed to transmit answer: %v", err)
			} else {
				logIt(m.LogCh, "Local Answer transmitted back to initiator.")
			}

		case "answer":
			if m.PC == nil {
				continue
			}
			remoteAnswer := webrtc.SessionDescription{Type: webrtc.SDPTypeAnswer, SDP: msg.SDP}
			if err := m.PC.SetRemoteDescription(remoteAnswer); err != nil {
				logIt(m.LogCh, "Failed to set remote answer: %v", err)
			} else {
				logIt(m.LogCh, "Remote Answer applied. Tunnel securing...")
			}

		case "candidate":
			if m.PC != nil && msg.Candidate != nil {
				if err := m.PC.AddICECandidate(*msg.Candidate); err != nil {
					logIt(m.LogCh, "Failed to add ICE candidate: %v", err)
				} else {
					logIt(m.LogCh, "Remote ICE candidate added.")
				}
			}
		}
	}
}

// creates the local offer and transmits it.
func (m *RTCManager) InitiateCall() {
	logIt(m.LogCh, "Generating Outbound Offer...")

	if m.PC == nil {
		if err := m.initPC(); err != nil {
			logIt(m.LogCh, "Failed to init PC: %v", err)
			return
		}
	}

	ordered := true
	maxRetransmits := uint16(30)
	dataChannel, err := m.PC.CreateDataChannel("fileTransfer", &webrtc.DataChannelInit{
		Ordered:        &ordered,
		MaxRetransmits: &maxRetransmits,
	})
	if err != nil {
		logIt(m.LogCh, "Failed to create data channel: %v", err)
		return
	}

	dataChannel.OnOpen(func() {
		logIt(m.LogCh, "Data channel '%s' is OPEN!", dataChannel.Label())
		go func() {
			for i := 0; i < 5; i++ {
				time.Sleep(1 * time.Second)
				message := fmt.Sprintf("Chunk packet batch #%d", i)
				_ = dataChannel.SendText(message)
				logIt(m.LogCh, "[SENT] -> %s", message)
			}
		}()
	})

	dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
		if msg.IsString {
			logIt(m.LogCh, "[RECV] <- %s", string(msg.Data))
		} else {
			logIt(m.LogCh, "[RECV] <- Binary chunk: %d bytes", len(msg.Data))
		}
	})

	offer, err := m.PC.CreateOffer(nil)
	if err != nil {
		logIt(m.LogCh, "Failed to create offer: %v", err)
		return
	}

	if err := m.PC.SetLocalDescription(offer); err != nil {
		logIt(m.LogCh, "Failed to set local description: %v", err)
		return
	}

	payload := SignalingMessage{Type: "offer", SDP: offer.SDP}
	if err := m.Conn.WriteJSON(payload); err != nil {
		logIt(m.LogCh, "Failed to send offer: %v", err)
	}
}
