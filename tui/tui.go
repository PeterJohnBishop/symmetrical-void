package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/gorilla/websocket"
)

type socketMsg string
type errMsg error

type Model struct {
	Conn *websocket.Conn
	Logs []string
	err  error
}

// listens for messages to send to the Update loop
func waitForMessage(conn *websocket.Conn) tea.Cmd {
	return func() tea.Msg {
		_, payload, err := conn.ReadMessage()
		if err != nil {
			return errMsg(err)
		}
		// Return the successful payload as a custom message type
		return socketMsg(string(payload))
	}
}

func (m Model) Init() tea.Cmd {
	return waitForMessage(m.Conn)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			// close the socket before exiting
			m.Conn.Close()
			return m, tea.Quit

		case "s":
			// Send test payload
			testPayload := `{"type": "test", "data": "Hello from the Omarchy terminal"}`
			err := m.Conn.WriteMessage(websocket.TextMessage, []byte(testPayload))
			if err != nil {
				m.err = err
			}
			m.Logs = append(m.Logs, "[SENT] -> "+testPayload)
			return m, nil
		}

	case socketMsg:
		m.Logs = append(m.Logs, "[RECV] <- "+string(msg))

		if len(m.Logs) > 15 {
			m.Logs = m.Logs[1:]
		}

		return m, waitForMessage(m.Conn)

	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, nil
}

func (m Model) View() tea.View {
	s := strings.Builder{}
	s.WriteString("WebRTC Signaling Monitor\n")
	s.WriteString("===========================\n\n")

	if m.err != nil {
		s.WriteString(fmt.Sprintf("Error: %v\n\n", m.err))
	}

	for _, logLine := range m.Logs {
		s.WriteString(logLine + "\n")
	}

	s.WriteString("\n[s] Send Test Ping   [q] Quit\n")
	return tea.NewView(s.String())
}
