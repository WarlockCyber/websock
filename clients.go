package main

import (
	"encoding/json"
	"fmt"
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

func (c *clientsSlice) updateRoomSubscribe(uid string,subscribeOnRoom string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	room:=subscribeOnRoom;
	if room=="null" {room=""}
	for _, cl := range clients.clients {
		if cl.uid.String()==uid {cl.subscribeOnRoom=room;}
	}
}
func (c *clientsSlice) setViewer(uid string,isViewer bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, cl := range clients.clients {
		if cl.uid.String()==uid {
			cl.isCharViewer=isViewer;
			clients.SendSubscribersCount(cl.char)
			return
		}
	}
}

func (c *clientsSlice) SendSubscribersCount(char string) {
	if char=="" {return}
	connectsCnt:=clients.countSubscriptions(char)
	editorsCnt:=clients.countEditors(char)
	usersCnt:=clients.uniqueSubscribersOnChar(char);

	for _, cl :=  range clients.clients {
		if (cl.char == char) {
			cl.send([]byte(fmt.Sprintf(`{"system":"usersInRoom","connects":"%d","editors":"%d","users":"%d"}`,connectsCnt,editorsCnt,usersCnt)))
		}
	}
}


func (c *clientsSlice) processMessage(room, uid string, message []byte) {
	const methodName = "process message"

	ms := new(clientMessage)

	err := json.Unmarshal([]byte(message), ms)
	if err != nil {
		log.Printf("%s: unmarshal message getting error %s", methodName, err.Error())
	}



	if ms.SubscribeOnRoom!="" {
		clients.updateRoomSubscribe(uid,ms.SubscribeOnRoom)
	}

	if ms.SetViewerMode {
		clients.setViewer(uid,true)
	}



	if ms.ForServerOnly {return} // no need to send this message

	if ms.Char != "" { // Char send messages
		for _, cl := range clients.clients {
			if uid != cl.uid.String() && ms.Char == cl.char && (cl.char != "" && cl.isValidChar || ms.SelfSend) {
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
			if (cl.room == room || (ms.RoomMessage && !ms.ForRoomOwnerOnly && cl.subscribeOnRoom == room)) && (uid != cl.uid.String() || ms.SelfSend) && cl.room != "" {
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

func (c *clientsSlice) uniqueSubscribersOnChar(char string) int {
	var mu sync.Mutex
	mu.Lock()
	defer mu.Unlock()

	r := make(map[string]bool)

	for _, v := range c.clients {
		if v.char==char {r[v.room] = true}
	}

	return len(r)
}

func (c *clientsSlice) uniqueSubscribersInRoom(room string) int {
	var mu sync.Mutex
	mu.Lock()
	defer mu.Unlock()

	r := make(map[string]bool)

	for _, v := range c.clients {
		if v.room==room || v.subscribeOnRoom==room {r[v.room] = true}
	}

	return len(r)
}

func (c *clientsSlice) countSubscriptions(char string) int {
	var mu sync.Mutex
	mu.Lock()
	defer mu.Unlock()
	var cnt int = 0;

	for _, v := range c.clients {
		if v.char==char  {cnt++}
	}

	return cnt;
}

func (c *clientsSlice) countEditors(char string) int {
	var mu sync.Mutex
	mu.Lock()
	defer mu.Unlock()
	var cnt int = 0;

	for _, v := range c.clients {
		if v.char==char && v.isCharViewer==false  {cnt++}
	}

	return cnt;
}