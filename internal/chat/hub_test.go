package chat

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"mygochat/internal/model"
)

func newTestHub() *Hub {
	return NewHub(slog.New(slog.NewTextHandler(io.Discard, nil)))
}

// dialServer spins up a websocket endpoint that registers each connection with hub.
func dialServer(t *testing.T, hub *Hub) (*websocket.Conn, func()) {
	t.Helper()

	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("upgrade: %v", err)
			return
		}
		hub.Register(conn)
	}))

	url := "ws" + strings.TrimPrefix(srv.URL, "http") + "/"
	client, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		srv.Close()
		t.Fatalf("dial: %v", err)
	}

	return client, func() {
		_ = client.Close()
		srv.Close()
	}
}

func TestHubBroadcastDelivers(t *testing.T) {
	hub := newTestHub()

	client, cleanup := dialServer(t, hub)
	defer cleanup()

	// Give the server goroutine time to register the client.
	time.Sleep(50 * time.Millisecond)

	hub.Broadcast(model.Message{Username: "u", Message: "hi", Color: "#fff"})

	_ = client.SetReadDeadline(time.Now().Add(time.Second))
	var got model.Message
	if err := client.ReadJSON(&got); err != nil {
		t.Fatalf("read: %v", err)
	}
	if got.Message != "hi" {
		t.Errorf("got message %q, want %q", got.Message, "hi")
	}
}

func TestHubUnregisterIsIdempotent(t *testing.T) {
	hub := newTestHub()

	c := &Client{conn: &websocket.Conn{}, send: make(chan model.Message, 1)}
	hub.mu.Lock()
	hub.clients[c] = struct{}{}
	hub.mu.Unlock()

	// Multiple concurrent unregisters must not panic on double-close.
	var wg sync.WaitGroup
	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			hub.Unregister(c)
		}()
	}
	wg.Wait()
}

func TestHubSendDropsWhenBufferFull(t *testing.T) {
	hub := newTestHub()
	c := &Client{conn: &websocket.Conn{}, send: make(chan model.Message, 1)}

	if !hub.Send(c, model.Message{Message: "1"}) {
		t.Fatal("first send should succeed")
	}
	if hub.Send(c, model.Message{Message: "2"}) {
		t.Fatal("second send should be dropped (buffer full)")
	}
}
