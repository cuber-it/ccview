package srv

import (
	"sync"

	"github.com/cuber-it/ccview/internal/parse"
)

// kindReset is an internal control event signalling subscribers that the
// active session has changed and they should clear their view. Never
// serialized to clients as a normal "data:" message — translated to an
// "event: reset" SSE frame by the stream handler.
const kindReset parse.Kind = "__ccview_reset__"

// Hub holds event history and fans out live events to subscribed clients.
// Safe for concurrent use.
type Hub struct {
	mu      sync.Mutex
	history []parse.Event
	clients map[chan parse.Event]struct{}
}

func newHub() *Hub {
	return &Hub{clients: map[chan parse.Event]struct{}{}}
}

// Publish appends ev to history and delivers it to all subscribers.
// Slow subscribers (buffer full) drop the event rather than blocking the publisher.
func (h *Hub) Publish(ev parse.Event) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.history = append(h.history, ev)
	for c := range h.clients {
		select {
		case c <- ev:
		default:
		}
	}
}

// Subscribe registers a new client. Returns a copy of the current history
// and a channel on which future events will be delivered. Call the
// returned unsubscribe func when done.
func (h *Hub) Subscribe() ([]parse.Event, <-chan parse.Event, func()) {
	h.mu.Lock()
	defer h.mu.Unlock()
	hist := make([]parse.Event, len(h.history))
	copy(hist, h.history)
	ch := make(chan parse.Event, 256)
	h.clients[ch] = struct{}{}
	return hist, ch, func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		if _, ok := h.clients[ch]; ok {
			delete(h.clients, ch)
			close(ch)
		}
	}
}

// Reset clears event history and tells every subscriber to drop its view.
// Called when the active session changes.
func (h *Hub) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.history = nil
	for c := range h.clients {
		select {
		case c <- parse.Event{Kind: kindReset}:
		default:
		}
	}
}
