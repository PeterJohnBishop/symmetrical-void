package tui

import (
	tea "charm.land/bubbletea/v2"
)

type Model struct {
	wsServerConnected bool
	webRTCConnected   bool
	availablePeers    []string
}

func InitialModel() Model {
	return Model{
		wsServerConnected: false,
		webRTCConnected:   false,
		availablePeers:    []string{},
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}
