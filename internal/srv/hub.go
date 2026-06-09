package srv

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/cuber-it/ccview/internal/parse"
)

// kindReset is an internal control event signalling subscribers that the
// active session has changed and they should clear their view. Never
// serialized to clients as a normal "data:" message — translated to an
// "event: reset" SSE frame by the stream handler.
const kindReset parse.Kind = "__ccview_reset__"

// kindMeta is an internal control event carrying {total,offset} JSON in Raw,
// translated to an "event: meta" SSE frame so the client knows how many older
// events exist above the pushed tail (for "load older").
const kindMeta parse.Kind = "__ccview_meta__"

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
	ch := make(chan parse.Event, 2048) // holds a full tail burst (replayCap) without dropping
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

// History returns a snapshot of all events published so far.
func (h *Hub) History() []parse.Event {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]parse.Event, len(h.history))
	copy(out, h.history)
	return out
}

// Switch installs the full history for a freshly loaded session and pushes a
// bounded view to every subscriber: a reset (clear), a meta frame (total count
// and the offset of the first pushed event), then the last `cap` events. Older
// events load on demand via /api/history. This is the one place a burst of many
// events is sent at once; the client channel buffer is sized to hold the tail.
func (h *Hub) Switch(history []parse.Event, cap int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.history = history
	total := len(history)
	start := total - cap
	if start < 0 {
		start = 0
	}
	meta := parse.Event{Kind: kindMeta, Raw: json.RawMessage(fmt.Sprintf(`{"total":%d,"offset":%d}`, total, start))}
	for c := range h.clients {
		send := func(ev parse.Event) {
			select {
			case c <- ev:
			default:
			}
		}
		send(parse.Event{Kind: kindReset})
		send(meta)
		for _, ev := range history[start:] {
			send(ev)
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
