package chat

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"mygochat/internal/chat"
	"mygochat/internal/config"
	"mygochat/internal/model"
)

// Handler upgrades HTTP connections to websockets and bridges them to the chat service.
type Handler struct {
	svc      *chat.Service
	hub      *chat.Hub
	logger   *slog.Logger
	upgrader websocket.Upgrader
	cfg      config.ChatConfig
}

func NewHandler(svc *chat.Service, hub *chat.Hub, logger *slog.Logger, cfg config.ChatConfig, allowedOrigins []string) *Handler {
	return &Handler{
		svc:      svc,
		hub:      hub,
		logger:   logger,
		cfg:      cfg,
		upgrader: newUpgrader(allowedOrigins),
	}
}

func newUpgrader(allowedOrigins []string) websocket.Upgrader {
	return websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			// Empty allow-list keeps the demo permissive; configure
			// CORS_TRUSTED_ORIGINS in production to lock this down.
			if len(allowedOrigins) == 0 {
				return true
			}
			origin := r.Header.Get("Origin")
			for _, allowed := range allowedOrigins {
				if origin == allowed {
					return true
				}
			}
			return false
		},
	}
}

const (
	maxMessageBytes = 4 * 1024
	pongWait        = 60 * time.Second
)

// ServeHTTP handles GET /ws.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Warn("websocket upgrade failed", slog.String("error", err.Error()))
		return
	}

	client := h.hub.Register(conn)
	defer func() {
		h.hub.Unregister(client)
		_ = conn.Close()
	}()

	conn.SetReadLimit(maxMessageBytes)
	_ = conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	ctx := r.Context()

	h.sendHistory(ctx, client)
	h.scheduleWelcome(client)
	h.readLoop(ctx, conn, client)
}

func (h *Handler) sendHistory(ctx context.Context, client *chat.Client) {
	history, err := h.svc.History(ctx)
	if err != nil {
		h.logger.Error("failed to load history", slog.String("error", err.Error()))
		return
	}
	for _, msg := range history {
		h.hub.Send(client, msg)
	}
}

func (h *Handler) scheduleWelcome(client *chat.Client) {
	if h.cfg.WelcomeMessage == "" {
		return
	}
	welcome := chat.SystemMessage(h.cfg.WelcomeMessage)
	if h.cfg.WelcomeTimeout <= 0 {
		h.hub.Send(client, welcome)
		return
	}
	time.AfterFunc(h.cfg.WelcomeTimeout, func() {
		// Safe even if the client already disconnected: Send is a
		// non-blocking enqueue and Unregister closes the channel via sync.Once.
		defer func() { _ = recover() }()
		h.hub.Send(client, welcome)
	})
}

func (h *Handler) readLoop(ctx context.Context, conn *websocket.Conn, client *chat.Client) {
	for {
		var msg model.Message
		if err := conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				h.logger.Warn("unexpected websocket close", slog.String("error", err.Error()))
			}
			return
		}

		if msg.Username == "" || msg.Message == "" {
			continue
		}

		if err := h.svc.Publish(ctx, msg); err != nil {
			h.logger.Error("failed to publish message", slog.String("error", err.Error()))
			return
		}
	}
}
