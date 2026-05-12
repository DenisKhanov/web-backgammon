package ws

import (
	"context"
	"log/slog"
	"time"

	"nhooyr.io/websocket"

	"github.com/denis/web-backgammon/internal/game"
)

const (
	pingInterval    = 15 * time.Second
	writeTimeout    = 10 * time.Second
	maxMessageBytes = 1024
)

// Client represents one connected WebSocket peer.
type Client struct {
	hub          *Hub
	conn         *websocket.Conn
	roomCode     string
	sessionToken string
	color        game.Color // assigned when the room starts
	name         string
	send         chan []byte
}

// run starts the read and write goroutines and blocks until the connection closes.
func (c *Client) run(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
		c.hub.unregister(c)
		c.conn.Close(websocket.StatusGoingAway, "disconnected")
	}()

	go c.writePump(ctx)
	c.readPump(ctx)
}

func (c *Client) readPump(ctx context.Context) {
	c.conn.SetReadLimit(maxMessageBytes)
	for {
		_, data, err := c.conn.Read(ctx)
		if err != nil {
			return
		}
		c.hub.dispatch(c, data)
	}
}

func (c *Client) writePump(ctx context.Context) {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				return
			}
			wCtx, cancel := context.WithTimeout(ctx, writeTimeout)
			err := c.conn.Write(wCtx, websocket.MessageText, msg)
			cancel()
			if err != nil {
				slog.Warn("ws write error", "err", err, "color", c.color)
				return
			}

		case <-ticker.C:
			pCtx, cancel := context.WithTimeout(ctx, writeTimeout)
			err := c.conn.Ping(pCtx)
			cancel()
			if err != nil {
				return
			}

		case <-ctx.Done():
			return
		}
	}
}

// sendMsg queues a pre-encoded JSON byte slice for delivery.
func (c *Client) sendMsg(b []byte) {
	select {
	case c.send <- b:
	default:
		slog.Warn("ws send buffer full, dropping message", "color", c.color)
	}
}
