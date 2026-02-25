package main

import (
	"encoding/json"
	"fmt"
	"log"
	"slices"
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

func (c *clientsSlice) Find(uid string) *wsConn {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, cl := range clients.clients {
		if cl.uid.String() == uid {
			return cl
		}
	}

	return nil
}

func (c *clientsSlice) Add(w *wsConn) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.clients = append(c.clients, w)
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

func (c *clientsSlice) updateRoomSubscribe(uid string, subscribeOnRoom string) {
	cl := c.Find(uid)
	if cl == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	room := subscribeOnRoom
	if room == "null" {
		room = ""
	}

	cl.subscribeOnRoom = room
}

func (c *clientsSlice) setViewer(uid string, isViewer bool) {
	cl := c.Find(uid)
	if cl == nil {
		return
	}

	cl.isCharViewer = isViewer
	clients.SendSubscribersCount(cl.char)
}

func (c *clientsSlice) addFileSubscription(uid string, char string) {
	cl := c.Find(uid)
	if cl == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if !slices.Contains(cl.charSaveSubscriptions, char) {
		cl.charSaveSubscriptions = append(cl.charSaveSubscriptions, char)
	}
}

func (c *clientsSlice) notifyFileSubscribers(char string, message string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, cl := range clients.clients {
		if slices.Contains(cl.charSaveSubscriptions, char) {
			cl.send([]byte(message))
		}
	}
}

func (c *clientsSlice) SendSubscribersCount(char string) {
	if char == "" || char == "0" {
		return
	}

	connectsCnt := clients.countSubscriptions(char)
	editorsCnt := clients.countEditors(char)
	usersCnt := clients.uniqueSubscribersOnChar(char)

	for _, cl := range clients.clients {
		if cl.char == char {
			cl.send([]byte(fmt.Sprintf(`{"system":"usersInRoom","connects":"%d","editors":"%d","users":"%d"}`, connectsCnt, editorsCnt, usersCnt)))
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

	if ms.SubscribeOnRoom != "" {
		clients.updateRoomSubscribe(uid, ms.SubscribeOnRoom)
	}

	if ms.SetViewerMode {
		clients.setViewer(uid, true)
	}

	if ms.SubscribeOnFile != "" {
		clients.addFileSubscription(uid, ms.SubscribeOnFile)
	}

	if ms.CharSaved != "" {
		clients.notifyFileSubscribers(ms.CharSaved, fmt.Sprintf("{\"system\":\"charSaved\",\"char\":\"%s\"}", ms.CharSaved))
	}

	if ms.ForServerOnly {
		return
	} // no need to send this message

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

func (c *clientsSlice) byChar(char string) []*wsConn {
	c.mu.Lock()
	defer c.mu.Unlock()

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
	c.mu.Lock()
	defer c.mu.Unlock()

	return len(c.clients)
}

func (c *clientsSlice) memory() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	ln := unsafe.Sizeof(c.clients)

	return int(ln)
}

func (c *clientsSlice) uniqueRoom() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	r := make(map[string]bool)

	for _, v := range c.clients {
		r[v.room] = true
	}

	return len(r)
}

func (c *clientsSlice) uniqueSubscribersOnChar(char string) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	r := make(map[string]bool)

	for _, v := range c.clients {
		if v.char == char {
			r[v.room] = true
		}
	}

	return len(r)
}

func (c *clientsSlice) uniqueSubscribersInRoom(room string) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	r := make(map[string]bool)

	for _, v := range c.clients {
		if v.room == room || v.subscribeOnRoom == room {
			r[v.room] = true
		}
	}

	return len(r)
}

func (c *clientsSlice) countSubscriptions(char string) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	var cnt int = 0

	for _, v := range c.clients {
		if v.char == char {
			cnt++
		}
	}

	return cnt
}

func (c *clientsSlice) countEditors(char string) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	var cnt int = 0

	for _, v := range c.clients {
		if v.char == char && v.isCharViewer == false {
			cnt++
		}
	}

	return cnt
}
