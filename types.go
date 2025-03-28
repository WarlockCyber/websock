package main

import "time"

const (
	portEnvName           = "WS_PORT"
	pprofPortEnvName      = "WS_PPROF_PORT"
	apiPortEnvName        = "WS_API_PORT"
	readBuferSizeEnvName  = "WS_READ_BUFER_SIZE"
	writeBuferSizeEnvName = "WS_WRITE_BUFER_SIZE"
	crtEnvName            = "WS_CRT_FILE"
	keyEnvName            = "WS_KEY_FILE"
	pingTimeoutEnvName    = "WS_PING_TIMEOUT"
	pprofEnabledEnvName   = "WS_PPROF_ENABLED"

	writeWait = time.Second

	subMessage   = "subscribe"
	unsubMessage = "unsubscribe"
	pingMessage  = "ping"

	roomParam = "room"
	dataParam = "data"
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
	SelfSend bool `json:"selfSend"`
	ForServerOnly bool `json:"forServerOnly"`
	SubscribeOnUser bool `json:"subscribeOnUser"`
}
