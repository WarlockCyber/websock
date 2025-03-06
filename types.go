package main

import "time"

var clients []*wsConn

const (
	portEnvName           = "WS_PORT"
	readBuferSizeEnvName  = "WS_READ_BUFER_SIZE"
	writeBuferSizeEnvName = "WS_WRITE_BUFER_SIZE"
	crtEnvName            = "WS_CRT_FILE"
	keyEnvName            = "WS_KEY_FILE"
	pingTimeoutEnvName    = "WS_PING_TIMEOUT"

	writeWait = time.Second

	subMessage   = "subscrbe"
	unsubMessage = "unsubscrbe"
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
