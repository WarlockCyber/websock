package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

func startServer() error {
	strPort := strconv.FormatInt(cfg.Port, 10)

	http.HandleFunc("/", handleConnections)
	log.Println("http server started on " + strPort)

	var err error

	if cfg.Crt != "" && cfg.Key != "" {
		err = http.ListenAndServeTLS(":"+strPort, cfg.Crt, cfg.Key, nil)
	} else {
		err = http.ListenAndServe(":"+strPort, nil)
	}

	if err != nil {
		return (err)
	}

	return nil
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	const methodName = "handleConnections"

	// обновление соединения до WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("%s: %s", methodName, err.Error())
		return
	}

	pr := strings.Split(strings.Trim(r.URL.Path, "/"), "|")

	var ch, rm string
	if len(pr) > 0 {
		rm = pr[0]
		if len(pr) > 1 {
			ch = pr[1]
		}
	}
	var mutex sync.Mutex

	mutex.Lock()
	wcon := newWsConn(ws, rm, ch)
	clients = append(clients, wcon)
	mutex.Unlock()

	log.Printf("%s: new connection %s opened", methodName, wcon.toString())
}

func subscribeUnsubscribeMessage(char string) error {
	sb := make([]*wsConn, 0)
	if char != "" {
		for _, c := range clients {
			if c.char == char {
				sb = append(sb, c)
			}
		}

		sys := unsubMessage
		if len(sb) > 1 {
			sys = subMessage
		}

		mess := &sysMessage{
			System: sys,
			On:     char,
		}

		ms, err := json.Marshal(mess)
		if err != nil {
			return err
		}

		for _, c := range sb {
			err := c.send(ms)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
