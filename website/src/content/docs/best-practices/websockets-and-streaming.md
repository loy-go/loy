---
title: "Best Practices: WebSockets & Real-Time Streaming"
description: "How to design thread-safe connection hubs, read/write pumps, and low-latency token streaming in Loy."
---

Generated via: `loy make ws <name>` ([ADR-021](/loy/adrs/))

Real-time streaming requires careful handling of goroutine lifecycles, connection heartbeats, and non-blocking channel dispatching to avoid memory leaks or deadlocks.

## Golden Rules

### 1. Dedicated Read & Write Pumps per Connection
- **DO NOT** read from and write to a WebSocket connection concurrently from multiple goroutines.
- **DO** use the scaffolded pattern:
  - **Read Pump:** Listens for inbound client messages and handles pong heartbeats.
  - **Write Pump:** Sole goroutine with write access to `websocket.Conn`, dispatching messages from the client's buffered `send` channel and emitting periodic ping heartbeats.

### 2. Thread-Safe Hub Unregistration
- Never mutate the hub's client map in place while holding a read lock (`RLock`).
- Route client disconnects through the hub's `unregister` channel:

```go
select {
case client.send <- message:
default:
    // Slow client: trigger unregistration safely
    go func(c *Client) {
        h.unregister <- c
    }(client)
}
```

### 3. Graceful Heartbeat Ping-Pong Timeouts
- Configure read deadlines on the client connection and reset them upon receiving pong frames. If a client drops connection silently, the read deadline triggers and closes the connection cleanly.
