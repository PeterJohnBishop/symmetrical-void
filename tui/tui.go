package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/peterjohnbishop/symmetrical-void/rtcomm"
)

type rtcLogMsg string

type Model struct {
	Logs       []string
	LogCh      chan string
	RTCManager *rtcomm.RTCManager
}

func waitForRTCLog(ch chan string) tea.Cmd {
	return func() tea.Msg {
		msg := <-ch
		return rtcLogMsg(msg)
	}
}

func (m Model) Init() tea.Cmd {
	go m.RTCManager.StartSignalingRouter()

	return waitForRTCLog(m.LogCh)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.RTCManager.Conn.Close()
			if m.RTCManager.PC != nil {
				m.RTCManager.PC.Close()
			}
			return m, tea.Quit

		case "s":
			m.Logs = append(m.Logs, "[SYSTEM] Initiating WebRTC Handshake...")
			go m.RTCManager.InitiateCall()
			return m, nil
		}

	case rtcLogMsg:
		m.Logs = append(m.Logs, string(msg))

		if len(m.Logs) > 15 {
			m.Logs = m.Logs[1:]
		}

		return m, waitForRTCLog(m.LogCh)
	}

	return m, nil
}

func (m Model) View() string {
	s := strings.Builder{}
	s.WriteString("WebRTC\n")

	for _, logLine := range m.Logs {
		s.WriteString(logLine)
		s.WriteString("\n")
	}

	s.WriteString("\n[s] Start WebRTC Handshake   [q] Quit\n")
	return s.String()
}
