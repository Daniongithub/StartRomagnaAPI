package sse

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

var SSE_HUB Hub

type Event struct {
	Type string // "vehicle_positions", "trip_updates", "service_alerts"
	Data string // JSON payload, or an empty string for a notification without data.
}

type Hub struct {
	mu      sync.RWMutex
	clients map[chan Event]struct{}
}

func NewHub() *Hub {
	return &Hub{clients: make(map[chan Event]struct{})}
}

func InitHub() {
	SSE_HUB = *NewHub()
}

func (h *Hub) Subscribe() chan Event {
	ch := make(chan Event, 8)
	h.mu.Lock()
	if h.clients == nil {
		h.clients = make(map[chan Event]struct{})
	}
	h.clients[ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

func (h *Hub) Unsubscribe(ch chan Event) {
	h.mu.Lock()
	if _, exists := h.clients[ch]; exists {
		delete(h.clients, ch)
		close(ch)
	}
	h.mu.Unlock()
}

func (h *Hub) Broadcast(e Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.clients {
		select {
		case ch <- e:
		default:
			// Drop the event if the client is too slow to receive it.
		}
	}
}

func (h *Hub) SSEHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Disable proxy buffering so SSE events are delivered immediately.

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming is not supported", http.StatusInternalServerError)
		return
	}

	ch := h.Subscribe()
	defer h.Unsubscribe(ch)
	controller := http.NewResponseController(w)
	write := func(message string) error {
		// A slow peer must not retain its handler forever.
		_ = controller.SetWriteDeadline(time.Now().Add(10 * time.Second))
		if _, err := fmt.Fprint(w, message); err != nil {
			return err
		}
		flusher.Flush()
		return nil
	}
	if err := write(": connected\n\n"); err != nil {
		return
	}

	clientGone := r.Context().Done()
	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-clientGone:
			return
		case ev, ok := <-ch:
			if !ok {
				return
			}
			if err := write(fmt.Sprintf("event: %s\ndata: %s\n\n", ev.Type, ev.Data)); err != nil {
				return
			}
		case <-heartbeat.C:
			// Keep the connection alive through proxies and load balancers.
			if err := write(": heartbeat\n\n"); err != nil {
				return
			}
		}
	}
}
