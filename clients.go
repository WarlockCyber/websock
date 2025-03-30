package main

import (
	"encoding/json"
	"log"
	"sync"
	"unsafe"
)

type clientsSlice struct {
	mu      sync.Mutex
	clients []*wsConn
}

var clients *clientsSlice

func init() {
	clients = newClientsSlice()
}

func newClientsSlice() *clientsSlice {
	return &clientsSlice{
		clients: make([]*wsConn, 0),
	}
}

func (c *clientsSlice) Add(w *wsConn) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.clients = append(c.clients, w)
}

func (c *clientsSlice) AddRoomSubscribe(uid string,subscribeOnRoom string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, cl := range clients.clients {
		if cl.uid.String()==uid {cl.subscribeOnRoom=subscribeOnRoom}
	}
}


func (c *clientsSlice) SendMessage(room, uid string, message []byte) {
	const methodName = "send message"

	ms := new(clientMessage)

	err := json.Unmarshal([]byte(message), ms)
	if err != nil {
		log.Printf("%s: unmarshal message getting error %s", methodName, err.Error())
	}

	if ms.SubscribeOnRoom!="" {clients.AddRoomSubscribe(uid,ms.SubscribeOnRoom)}


	if ms.ForServerOnly {return} // no need to send this message

	if ms.Char != "" { // Char sen messages
		for _, cl := range clients.clients {
			if uid != cl.uid.String() && ms.Char == cl.char && (cl.char != "" && cl.isValidChar || ms.SelfSend || ms.SelfOnly) {
				log.Printf("sending to char %s| %s | %s", cl.char, substr(string(message), 0, logLen), cl.toString())
				err := cl.send(message)
				if err != nil {
					log.Printf("%s: send message getting error %s", methodName, err.Error())
					continue
				}
			}
		}
	} else { // Room messages
		for _, cl := range clients.clients {
			if (cl.room == room || (cl.subscribeOnRoom == room &&  !ms.SelfOnly)) && (uid != cl.uid.String() || ms.SelfSend || ms.SelfOnly) && cl.room != "" {
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

func (c *clientsSlice) Remove(uuid string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for n, cl := range c.clients {
		if cl.uid.String() == uuid {
			c.clients = append(c.clients[:n], c.clients[n+1:]...)
			break
		}
	}
}

func (c *clientsSlice) byChar(char string) []*wsConn {
	sb := make([]*wsConn, 0)
	for _, c := range c.clients {
		if c.char == char {
			sb = append(sb, c)
		}
	}

	return sb
}

func (c *clientsSlice) byRoom(room string) []*wsConn {
	sb := make([]*wsConn, 0)
	for _, c := range c.clients {
		if c.room == room {
			sb = append(sb, c)
		}
	}

	return sb
}

func (c *clientsSlice) len() int {
	var mu sync.Mutex
	mu.Lock()
	defer mu.Unlock()
	return len(c.clients)
}

func (c *clientsSlice) memory() int {
	var mu sync.Mutex
	mu.Lock()
	defer mu.Unlock()

	ln := unsafe.Sizeof(c.clients)

	return int(ln)
}

func (c *clientsSlice) uniqueRoom() int {
	var mu sync.Mutex
	mu.Lock()
	defer mu.Unlock()

	r := make(map[string]bool)

	for _, v := range c.clients {
		r[v.room] = true
	}

	return len(r)
}
