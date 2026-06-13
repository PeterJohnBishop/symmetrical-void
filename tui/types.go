package tui

type connectedMsg struct{}

type errMsg struct{ err error }

type logMsg string
type localFrameMsg string
type remoteFrameMsg string
