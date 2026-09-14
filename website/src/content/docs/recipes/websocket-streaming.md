---
title: "Recipe: Real-Time WebSocket Streaming"
description: "How to build high-concurrency token streaming and bidirectional chat with Fiber WebSockets."
---

Real-time AI platforms require token-by-token LLM output streaming, live candidate chat, and telemetry. Loy provides dedicated WebSocket scaffolding over Fiber WebSockets ([ADR-021](/loy/adrs/)).

## 1. Architecture Flow

```
Browser Client ──► Fiber WebSocket Upgrader ──► Client Read Pump
                                            ├──► Client Write Pump (Heartbeat/Ping-Pong)
                                            └──► Thread-Safe Connection Hub
```

## 2. Scaffolding WebSocket Endpoints

Run `loy make ws` to generate connection infrastructure:

```bash
loy make ws chat
```

Scaffolded artifacts in `internal/chat/transport/ws/`:

- **`protocol.go`**: Strongly typed JSON frame format (`MessageType`, `SequenceID`, `Payload`).
- **`hub.go`**: Mutex-guarded connection registry with broadcast channels and safe unregister loops.
- **`client.go`**: Dedicated read pump, write pump, and ping-pong heartbeat tickers.

## 3. Registering the WebSocket Handler

Mount the WebSocket upgrader in your transport route registration:

```go title="internal/chat/transport/http/routes.go"
package http

import (
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"job-runner/internal/chat/transport/ws"
)

func RegisterWSRoutes(router fiber.Router, hub *ws.Hub) {
	go hub.Run()

	router.Get("/ws/chat/:session_id", websocket.New(func(c *websocket.Conn) {
		sessionID := c.Params("session_id")
		client := ws.NewClient(sessionID, hub, c)

		hub.Register(client)

		go client.WritePump()
		client.ReadPump()
	}))
}
```

## 4. Streaming Token Frames

Push streaming chunks directly to connected clients:

```go
func StreamAIResponse(hub *ws.Hub, sessionID string, tokenStream <-chan string) {
    for token := range tokenStream {
        hub.Broadcast(ws.Frame{
            Type: ws.MessageTypeData,
            Payload: map[string]string{
                "session_id": sessionID,
                "delta":      token,
            },
        })
    }
}
```
