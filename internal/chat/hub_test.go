package chat

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"mygochat/internal/model"
)

func newTestHub() *Hub {
	return NewHub(slog.New(slog.NewTextHandler(io.Discard, nil)))
}

// dialServer spins up a websocket endpoint that registers each connection with hub.
func dialServer(t *testing.T, hub *Hub) (*websocket.Conn, func()) {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
		if err != nil {
			t.Errorf("upgrade: %v", err)
			return
		}
		hub.Register(conn)
	}))

	url := "ws" + strings.TrimPrefix(srv.URL, "http") + "/"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	client, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		srv.Close()
		t.Fatalf("dial: %v", err)
	}

	return client, func() {
		_ = client.CloseNow()
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

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var got model.Message
	if err := wsjson.Read(ctx, client, &got); err != nil {
		t.Fatalf("read: %v", err)
	}
	if got.Message != "hi" {
		t.Errorf("got message %q, want %q", got.Message, "hi")
	}
}

func TestHubUnregisterIsIdempotent(t *testing.T) {
	hub := newTestHub()

	c := &Client{send: make(chan model.Message, 1)}
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
	c := &Client{send: make(chan model.Message, 1)}

	if !hub.Send(c, model.Message{Message: "1"}) {
		t.Fatal("first send should succeed")
	}
	if hub.Send(c, model.Message{Message: "2"}) {
		t.Fatal("second send should be dropped (buffer full)")
	}
}
