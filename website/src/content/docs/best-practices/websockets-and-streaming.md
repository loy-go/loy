---
title: "Best Practices: WebSockets & Real-Time Streaming"
description: "How to design thread-safe connection hubs, read/write pumps, ping-pong heartbeats, and low-latency token streaming in Loy."
---

Generated via:
```bash
loy make ws <name>
```

Real-time streaming (chat applications, live order updates, LLM token streaming, dashboard feeds) introduces concurrency hazards that typical REST APIs don't have: concurrent writes to the same TCP connection, slow client head-of-line blocking, and silent connection drops.

Loy provides an enterprise-grade WebSocket architecture ([ADR-021](/loy/adrs/)) implementing dedicated read/write pumps, channel-based hub coordination, and heartbeat deadlines.

---

## Architectural Topography

```text
┌────────────────────────────────────────────────────────────────────────┐
│                        CENTRAL WEBSOCKET HUB                           │
│  register chan • unregister chan • broadcast chan • clients map[Client]│
└───────────────────▲────────────────────────────────┬───────────────────┘
                    │                                │
     Subscribes /   │                                │ Broadcasts to
     Unsubscribes   │                                │ buffered channels
                    │                                ▼
       ┌────────────┴──────────┐        ┌─────────────────────────┐
       │   CLIENT CONNECTION   │        │    CLIENT CONNECTION    │
       │  ┌──────────────────┐ │        │  ┌───────────────────┐  │
       │  │ Read Pump        │ │        │  │ Read Pump         │  │
       │  │ (inbound frames) │ │        │  │ (inbound frames)  │  │
       │  ├──────────────────┤ │        │  ├───────────────────┤  │
       │  │ Write Pump       │ │        │  │ Write Pump        │  │
       │  │ (outbound queue) │ │        │  │ (outbound queue)  │  │
       │  └──────────────────┘ │        │  └───────────────────┘  │
       └───────────────────────┘        └─────────────────────────┘
```

---

## Golden Rules

### 1. Dedicated Read and Write Pumps per Connection
Standard WebSocket connections (`gorilla/websocket` or `nhooyr/websocket`) **do not support concurrent writes**. If two goroutines write to the connection simultaneously, the connection crashes with a race condition or corrupts the TLS frame.

Loy strictly decouples read and write loops into dedicated goroutines:
- **Read Pump (`readPump`)**:
  - The only goroutine reading from the WebSocket.
  - Sets read deadlines and resets them when `pong` frames arrive.
  - Forwards inbound messages to application use cases or the Hub.
  - Closes connection and unregisters upon receiving EOF or error.
- **Write Pump (`writePump`)**:
  - The **sole goroutine** with write access to the WebSocket connection.
  - Listens to the client's buffered `send` channel (`chan []byte`).
  - Emits periodic `ping` frames to keep TCP and NAT proxies alive.

```go title="internal/chat/transport/ws/client.go"
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
```

---

### 2. Guard Against Slow Clients (Non-Blocking Hub Dispatch)
A common failure mode in real-time systems is a **slow client** (e.g. mobile user with degraded cellular connection). If the Hub attempts a blocking write to that client's send channel, the entire central Hub freezes, delaying message delivery to all thousands of other connected users!

Loy implements **non-blocking channel selection**:

```go title="internal/chat/transport/ws/hub.go"
case message := <-h.broadcast:
	for client := range h.clients {
		select {
		case client.send <- message:
			// Dispatched successfully into client's buffer
		default:
			// Client's buffer is full (slow client): unregister and close!
			close(client.send)
			delete(h.clients, client)
		}
	}
```

If a client's buffer fills up, Loy drops the slow connection immediately rather than allowing one slow client to degrade the entire platform.

---

### 3. Ping-Pong Heartbeat Deadlines

To detect dead connections (such as closed laptop lids, mobile tunnel transitions, or killed browser tabs):

```go
const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 * 1024 // 512KB
)

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		c.hub.broadcast <- message
	}
}
```

If a client stops responding to pings within 60 seconds, the read deadline expires, terminating the dead socket and reclaiming memory.

---

### 4. Integrating with Domain Events

When domain entities mutate (e.g. `OrderCreated` or `PaymentCompleted`), Application Services publish domain events ([`loy make event`](/loy/reference/cli/loy_make_event/)).

A WebSocket listener subscribes to the domain event and forwards it to the Hub:

```go title="internal/order/transport/ws/listener.go"
type OrderEventListener struct {
	hub *ws.Hub
}

func (l *OrderEventListener) HandleOrderCreated(ctx context.Context, event domain.OrderCreatedEvent) error {
	payload, _ := json.Marshal(event)
	l.hub.BroadcastToRoom(fmt.Sprintf("tenant:%s", event.TenantID), payload)
	return nil
}
```
