package main

import (
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

func (c *clientsSlice) Send(char string, msg string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	chr := c.byChar(char)
	for _, ch := range chr {
		if err := ch.send([]byte(msg)); err != nil {
			return err
		}
	}

	return nil
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
