package p2p

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/peterjohnbishop/symmetrical-void/wsclient"
	"github.com/pion/webrtc/v3"
)

type WebRTCManager struct {
	PC *webrtc.PeerConnection
	DC *webrtc.DataChannel
}

func (m *WebRTCManager) StartWebRTC(c *wsclient.ConnectionManager) error {
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{{URLs: []string{"stun:stun.l.google.com:19302"}}},
	}

	pc, err := webrtc.NewPeerConnection(config)
	if err != nil {
		return fmt.Errorf("failed to create peer connection: %w", err)
	}
	m.PC = pc

	dc, err := m.PC.CreateDataChannel("dataTransfer", nil)
	if err != nil {
		return fmt.Errorf("failed to create data channel: %w", err)
	}

	m.DC = dc

	dc.OnOpen(func() {
		log.Println("Data channel is open!")
	})

	dc.OnMessage(func(msg webrtc.DataChannelMessage) {
		log.Printf("Received: %s", string(msg.Data))
	})

	m.PC.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate != nil {
			candidateJSON := candidate.ToJSON()

			candidateBytes, err := json.Marshal(candidateJSON)
			if err != nil {
				log.Println("Failed to marshal ICE candidate:", err)
				return
			}

			c.SendEventMessage("candidate", "ICE Candidate", candidateBytes)
		}
	})

	return nil
}

func (m *WebRTCManager) SendOffer(c *wsclient.ConnectionManager) error {
	if m.PC == nil {
		return fmt.Errorf("peer connection is nil. Call StartWebRTC first")
	}

	offer, err := m.PC.CreateOffer(nil)
	if err != nil {
		return fmt.Errorf("failed to create offer: %w", err)
	}

	if err := m.PC.SetLocalDescription(offer); err != nil {
		return fmt.Errorf("failed to set local description: %w", err)
	}

	offerBytes, err := json.Marshal(offer)
	if err != nil {
		return fmt.Errorf("failed to marshal offer: %w", err)
	}

	c.SendEventMessage("offer", "WebRTC Offer", offerBytes)

	log.Println("Outbound offer generated and sent to signaling server")
	return nil
}

func (m *WebRTCManager) HandleOffer(c *wsclient.ConnectionManager, remoteSDP string) error {
	if m.PC == nil {
		return fmt.Errorf("peer connection not initialized")
	}

	// apply offer
	offer := webrtc.SessionDescription{Type: webrtc.SDPTypeOffer, SDP: remoteSDP}
	if err := m.PC.SetRemoteDescription(offer); err != nil {
		return fmt.Errorf("failed to set remote description: %w", err)
	}

	// generate answer
	answer, err := m.PC.CreateAnswer(nil)
	if err != nil {
		return fmt.Errorf("failed to create answer: %w", err)
	}

	// apply answer
	if err := m.PC.SetLocalDescription(answer); err != nil {
		return fmt.Errorf("failed to set local description: %w", err)
	}

	// send answer
	answerBytes, _ := json.Marshal(answer)
	c.SendEventMessage("answer", "WebRTC Answer", answerBytes)

	log.Println("Offer accepted. Outbound answer sent.")
	return nil
}

func (m *WebRTCManager) HandleAnswer(remoteSDP string) error {
	if m.PC == nil {
		return fmt.Errorf("peer connection not initialized")
	}

	answer := webrtc.SessionDescription{Type: webrtc.SDPTypeAnswer, SDP: remoteSDP}

	if err := m.PC.SetRemoteDescription(answer); err != nil {
		return fmt.Errorf("failed to apply remote answer: %w", err)
	}

	log.Println("Handshake complete for P2P tunnel.")
	return nil
}

// send text message
func (m *WebRTCManager) SendTextMessage(text string) error {
	if m.DC == nil || m.DC.ReadyState() != webrtc.DataChannelStateOpen {
		return fmt.Errorf("data channel is not open")
	}

	log.Printf("[TEXT] -> %s", text)
	return m.DC.SendText(text)
}

// send []byte data
func (m *WebRTCManager) SendBinaryData(data []byte) error {
	if m.DC == nil || m.DC.ReadyState() != webrtc.DataChannelStateOpen {
		return fmt.Errorf("data channel is not open")
	}

	log.Printf("[BINARY] -> Sending %d bytes", len(data))
	return m.DC.Send(data)
}
