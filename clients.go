package main

import (
	"sync"
)

type clientsSlice struct {
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
	var mu sync.Mutex
	mu.Lock()
	defer mu.Unlock()
	c.clients = append(c.clients, w)
}

func (c *clientsSlice) Remove(uuid string) {
	var mu sync.Mutex

	for n, cl := range c.clients {
		if cl.uid.String() == uuid {
			mu.Lock()
			c.clients = append(c.clients[:n], c.clients[n+1:]...)
			mu.Unlock()
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
