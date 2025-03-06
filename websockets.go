package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader websocket.Upgrader

type wsConn struct {
	ws         *websocket.Conn
	char       string
	room       string
	uid        uuid.UUID
	pingTicker *time.Ticker
	tickerDone chan bool
}

func initWS() {
	upgrader = websocket.Upgrader{
		ReadBufferSize:  int(cfg.ReadBuferSize),
		WriteBufferSize: int(cfg.WriteBuferSize),
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	clients = make([]*wsConn, 0)
}

func newWsConn(c *websocket.Conn, room, char string) *wsConn {
	conn := &wsConn{
		ws:   c,
		char: char,
		room: room,
		uid:  uuid.New(),
	}

	pingTicker := time.NewTicker(1 * time.Minute)
	done := make(chan bool)

	conn.pingTicker = pingTicker
	conn.tickerDone = done

	pingMess := new(sysMessage)
	pingMess.System = pingMessage
	pm, _ := json.Marshal(pingMess)

	go func() {
		for {
			select {
			case <-conn.tickerDone:
				return
			case t := <-conn.pingTicker.C:
				log.Printf("Ping at %s for connection %s", t.String(), conn.uid.String())
				conn.send(pm)
			}
		}
	}()

	c.SetCloseHandler(conn.onClose)

	go conn.handle()

	return conn
}

func (c *wsConn) handle() {
	const methodName = "handle connection"

	defer c.ws.Close()

	// цикл обработки сообщений
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			log.Printf("%s: read message getting error %s", methodName, err.Error())
			break
		}

		ms := new(clientMessage)

		err = json.Unmarshal([]byte(message), ms)
		if err != nil {
			log.Printf("%s: unmarshal message getting error %s", methodName, err.Error())
		}

		if ms.Char != "" {
			for _, cl := range clients {
				if c.uid != cl.uid {
					err := cl.send(message)
					if err != nil {
						log.Printf("%s: send message getting error %s", methodName, err.Error())
						continue
					}
				}
			}
		} else {
			for _, cl := range clients {
				if cl.room == c.room {
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

func (c *wsConn) onClose(code int, text string) error {
	const methodName = "connection close"

	log.Printf("%s: connection %s closed", methodName, c.toString())

	var mutex sync.Mutex

	for n, cl := range clients {
		if cl.uid == c.uid {
			mutex.Lock()
			clients = append(clients[:n], clients[n+1:]...)
			mutex.Unlock()
		}
	}

	message := websocket.FormatCloseMessage(code, "")
	c.ws.WriteControl(websocket.CloseMessage, message, time.Now().Add(writeWait))

	c.tickerDone <- true

	return nil
}

func (c *wsConn) send(m []byte) error {
	return c.ws.WriteMessage(websocket.TextMessage, m)
}

func (c *wsConn) toString() string {
	return "uuid: " + c.uid.String() + ", room " + c.room + ", char " + c.char
}
