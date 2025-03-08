package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"
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
		v, err := strconv.ParseInt(rm, 10, 64)
		if err != nil || v == 0 {
			rm = ""
		}

		if len(pr) > 1 {
			ch = pr[1]
		}
	}

	wcon := newWsConn(ws, rm, ch)

	clients.Add(wcon)

	log.Printf("connect    | %s", wcon.toString())
}
