package chat

import (
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

type Message struct {
	Id       int64     `json:"id"`
	Username string    `json:"username"`
	Message  string    `json:"message"`
	Time     time.Time `json:"time" db:"time" sql:"type:datetime"`
	Color    string    `json:"color" db:"color" sql:"type:string"`
}

// ClientManager manages connected clients with thread safety
type ClientManager struct {
	clients map[*websocket.Conn]bool
	mutex   sync.RWMutex
}

func (cm *ClientManager) Add(conn *websocket.Conn) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.clients[conn] = true
}

func (cm *ClientManager) Remove(conn *websocket.Conn) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	delete(cm.clients, conn)
}

func (cm *ClientManager) GetAll() map[*websocket.Conn]bool {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	clients := make(map[*websocket.Conn]bool)
	for conn, active := range cm.clients {
		clients[conn] = active
	}
	return clients
}

// Global client manager
var Clients = &ClientManager{
	clients: make(map[*websocket.Conn]bool),
}
