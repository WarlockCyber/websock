package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"net/http/pprof"
	"strconv"
	"strings"
)

func startServer() error {
	strPort := strconv.FormatInt(cfg.Port, 10)
	strPPPort := strconv.FormatInt(cfg.PprofPort, 10)

	sockR := http.NewServeMux()
	sockR.HandleFunc("/", handleConnections)

	socServer := &http.Server{
		Addr:    ":" + strPort,
		Handler: sockR,
	}

	r := http.NewServeMux()
	// Регистрация pprof-обработчиков
	r.HandleFunc("/debug/pprof/", pprof.Index)
	r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("/debug/pprof/profile", pprof.Profile)
	r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	r.HandleFunc("/debug/pprof/trace", pprof.Trace)

	pprofServer := &http.Server{
		Addr:    ":" + strPPPort,
		Handler: r,
	}

	defer func() {
		socServer.Close()
		pprofServer.Close()
	}()

	errChan := make(chan error)

	go func() {
		if cfg.Crt != "" && cfg.Key != "" {
			if err := pprofServer.ListenAndServeTLS(cfg.Crt, cfg.Key); err != nil {
				errChan <- err
			}
		} else {
			if err := pprofServer.ListenAndServe(); err != nil {
				errChan <- err
			}
		}
	}()

	go func() {
		if cfg.Crt != "" && cfg.Key != "" {
			if err := socServer.ListenAndServeTLS(cfg.Crt, cfg.Key); err != nil {
				errChan <- err
			}
		} else {
			if err := socServer.ListenAndServe(); err != nil {
				errChan <- err
			}
		}
	}()

	log.Println("http server started on " + strPort)
	log.Println("pprof http server started on " + strPPPort)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigChan:
		return nil
	case e := <-errChan:
		return e
	}
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
