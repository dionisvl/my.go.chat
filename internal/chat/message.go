package chat

import (
	"github.com/gorilla/websocket"
	"time"
)

type Message struct {
	Id       int64     `json:"id"`
	Username string    `json:"username"`
	Message  string    `json:"message"`
	Time     time.Time `json:"time" db:"time" sql:"type:datetime"`
	Color    string    `json:"color" db:"color" sql:"type:string"`
}

// Global clients map
var Clients = make(map[*websocket.Conn]bool)
