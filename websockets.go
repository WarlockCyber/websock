package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader websocket.Upgrader

const logLen = 50

type wsConn struct {
	mu          sync.Mutex
	ws          *websocket.Conn
	char        string
	isValidChar bool
	room        string
	uid         uuid.UUID
	pingTicker  *time.Ticker
	tickerDone  chan bool
}

var md5test = regexp.MustCompile("^[a-f0-9]{32}$")

func initWS() {
	upgrader = websocket.Upgrader{
		ReadBufferSize:  int(cfg.ReadBuferSize),
		WriteBufferSize: int(cfg.WriteBuferSize),
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
}

func newWsConn(c *websocket.Conn, room, char string) *wsConn {
	conn := &wsConn{
		ws:          c,
		char:        char,
		isValidChar: md5test.Match([]byte(char)),
		room:        room,
		uid:         uuid.New(),
		pingTicker:  time.NewTicker(time.Duration(cfg.PingTimeout) * time.Second),
		tickerDone:  make(chan bool),
	}

	c.SetCloseHandler(conn.onClose)

	go conn.handle()

	return conn
}

func (c *wsConn) handle() {
	const methodName = "handle connection"

	defer func() {
		c.close()
	}()

	err := c.initPing()
	if err != nil {
		log.Printf("%s: init ping getting error %s", methodName, err.Error())
		return
	}

	if err := c.subscribeUnsubscribeMessage(); err != nil {
		log.Printf("%s: subscribe getting error %s", methodName, err.Error())
		return
	}

	// цикл обработки сообщений
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			log.Printf("%s: read message getting error %s", methodName, err.Error())
			break
		}

		log.Printf("recived | %s | %s", substr(string(message), 0, logLen), c.toString())

		ms := new(clientMessage)

		err = json.Unmarshal([]byte(message), ms)
		if err != nil {
			log.Printf("%s: unmarshal message getting error %s", methodName, err.Error())
		}

		if ms.Char != "" {
			for _, cl := range clients.clients {
				if c.uid != cl.uid && ms.Char == cl.char && cl.char != "" && cl.isValidChar {
					log.Printf("sending to char %s| %s | %s", cl.char, substr(string(message), 0, logLen), cl.toString())
					err := cl.send(message)
					if err != nil {
						log.Printf("%s: send message getting error %s", methodName, err.Error())
						continue
					}
				}
			}
		} else {
			for _, cl := range clients.clients {
				if cl.room == c.room && c.uid != cl.uid && cl.room != "" {
					log.Printf("sending to room %s| %s | %s", cl.room, substr(string(message), 0, logLen), cl.toString())
					err := cl.send(message)
					if err != nil {
						log.Printf("%s: send message getting error %s", methodName, err.Error())
						continue
					}
				}
			}
		}
	}
}

func (c *wsConn) initPing() error {
	pingMess := &sysMessage{
		System: pingMessage,
	}
	pm, err := json.Marshal(pingMess)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-c.tickerDone:
				return
			case _ = <-c.pingTicker.C:
				c.send(pm)
			}
		}
	}()

	return nil
}

func (c *wsConn) stopPing() {
	c.tickerDone <- true
}

func (c *wsConn) close() {
	const methodName = "connection close"

	c.stopPing()
	clients.Remove(c.uid.String())

	if err := c.subscribeUnsubscribeMessage(); err != nil {
		log.Printf("%s: subscribe ping getting error %s", methodName, err.Error())
	}

	message := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")

	c.mu.Lock()
	defer c.mu.Unlock()

	c.ws.WriteControl(websocket.CloseMessage, message, time.Now().Add(writeWait))

	c.ws.Close()

	log.Printf("disconnect | %s", c.toString())

	c = nil
}

func (c *wsConn) onClose(code int, text string) error {
	c.close()

	return nil
}

func (c *wsConn) subscribeUnsubscribeMessage() error {
	if c.char != "" {
		sb := clients.byChar(c.char)

		sys := unsubMessage
		if len(sb) > 1 {
			sys = subMessage
		}

		mess := &sysMessage{
			System: sys,
			On:     c.char,
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

func (c *wsConn) send(m []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.ws.WriteMessage(websocket.TextMessage, m)
}

func (c *wsConn) toString() string {
	return fmt.Sprintf("%d session | %d users | room %s | char %s", clients.memory(), clients.len(), c.room, c.char)
}

func substr(input string, start int, length int) string {
	asRunes := []rune(input)

	if start >= len(asRunes) {
		return ""
	}

	if start+length > len(asRunes) {
		length = len(asRunes) - start
	}

	return string(asRunes[start : start+length])
}
