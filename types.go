package main

import "time"

const (
	portEnvName           = "WS_PORT"
	pprofPortEnvName      = "WS_PPROF_PORT"
	readBuferSizeEnvName  = "WS_READ_BUFER_SIZE"
	writeBuferSizeEnvName = "WS_WRITE_BUFER_SIZE"
	crtEnvName            = "WS_CRT_FILE"
	keyEnvName            = "WS_KEY_FILE"
	pingTimeoutEnvName    = "WS_PING_TIMEOUT"

	writeWait = time.Second

	subMessage   = "subscribe"
	unsubMessage = "unsubscribe"
	pingMessage  = "ping"
)

const (
	defReadSize    = 1024
	defWriteSize   = 1024
	defPingTimeout = 60
)

type sysMessage struct {
	System string `json:"system"`
	On     string `json:"on"`
}

type clientMessage struct {
	Char string `json:"char"`
}
