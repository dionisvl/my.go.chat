package chat

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"mygochat/internal/model"
)

const (
	// sendBuffer is the per-client outbound queue depth. Slow clients that
	// fill it are dropped rather than blocking the broadcaster.
	sendBuffer = 64
	// writeWait is the time allowed to write a single message to a peer.
	writeWait = 10 * time.Second
	// pongWait is how long we wait for a pong before considering the peer dead.
	pongWait = 60 * time.Second
	// pingPeriod must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

// Client is a single websocket connection owned by exactly one writer goroutine.
type Client struct {
	conn *websocket.Conn
	send chan model.Message
	once sync.Once
}

// Hub tracks connected clients and fans out messages to them.
// All writes to a given connection go through that client's send channel,
// so no two goroutines ever write to the same socket concurrently.
type Hub struct {
	logger  *slog.Logger
	mu      sync.RWMutex
	clients map[*Client]struct{}
}

func NewHub(logger *slog.Logger) *Hub {
	return &Hub{
		logger:  logger,
		clients: make(map[*Client]struct{}),
	}
}

// Register adds a connection and starts its dedicated writer goroutine.
func (h *Hub) Register(conn *websocket.Conn) *Client {
	c := &Client{
		conn: conn,
		send: make(chan model.Message, sendBuffer),
	}

	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()

	go h.writePump(c)
	return c
}

// Unregister removes a client and closes its send channel exactly once.
func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	if _, ok := h.clients[c]; ok {
		delete(h.clients, c)
	}
	h.mu.Unlock()

	c.once.Do(func() { close(c.send) })
}

// Send queues a message to a single client. Returns false if the client's
// buffer is full (slow consumer) — caller may choose to drop it.
func (h *Hub) Send(c *Client, msg model.Message) bool {
	select {
	case c.send <- msg:
		return true
	default:
		h.logger.Warn("client send buffer full, dropping")
		return false
	}
}

// Broadcast queues a message to every connected client.
func (h *Hub) Broadcast(msg model.Message) {
	h.mu.RLock()
	clients := make([]*Client, 0, len(h.clients))
	for c := range h.clients {
		clients = append(clients, c)
	}
	h.mu.RUnlock()

	for _, c := range clients {
		if !h.Send(c, msg) {
			h.Unregister(c)
			_ = c.conn.CloseNow()
		}
	}
}

// writePump is the single writer for a client's connection. It drains the
// send channel and emits periodic pings to keep the connection alive.
func (h *Hub) writePump(c *Client) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.CloseNow()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				_ = c.conn.Close(websocket.StatusNormalClosure, "")
				return
			}
			ctx, cancel := context.WithTimeout(context.Background(), writeWait)
			err := wsjson.Write(ctx, c.conn, msg)
			cancel()
			if err != nil {
				h.logger.Debug("write failed, closing client", slog.String("error", err.Error()))
				return
			}
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), writeWait)
			err := c.conn.Ping(ctx)
			cancel()
			if err != nil {
				return
			}
		}
	}
}
