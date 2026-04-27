package srv

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cuber-it/ccview/internal/parse"
)

const sseKeepalive = 20 * time.Second

// sseWriter wraps an http.ResponseWriter for Server-Sent Events. Each write
// returns true on success and false if the connection has been broken.
type sseWriter struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

// newSSEWriter sets the SSE response headers and returns a writer.
// Returns false if the underlying ResponseWriter cannot flush.
func newSSEWriter(w http.ResponseWriter) (*sseWriter, bool) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, false
	}
	h := w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache")
	h.Set("Connection", "keep-alive")
	h.Set("X-Accel-Buffering", "no")
	return &sseWriter{w: w, flusher: flusher}, true
}

// open emits an initial comment so the browser sees the stream open even
// before any real event arrives.
func (s *sseWriter) open() bool { return s.writeRaw(": ok\n\n") }

// keepalive emits a comment frame to keep proxies from closing the connection.
func (s *sseWriter) keepalive() bool { return s.writeRaw(": keepalive\n\n") }

// writeEvent emits ev as a "data:" frame, or as an "event: reset" frame for
// the internal reset signal. JSON-encode failures skip the event but keep
// the stream alive.
func (s *sseWriter) writeEvent(ev parse.Event) bool {
	if ev.Kind == kindReset {
		return s.writeRaw("event: reset\ndata: {}\n\n")
	}
	data, err := json.Marshal(ev)
	if err != nil {
		return true
	}
	return s.writeRaw(fmt.Sprintf("data: %s\n\n", data))
}

func (s *sseWriter) writeRaw(line string) bool {
	if _, err := io.WriteString(s.w, line); err != nil {
		return false
	}
	s.flusher.Flush()
	return true
}

// replayHistory dumps all buffered events to the client. Returns false if
// the connection broke mid-replay.
func replayHistory(sw *sseWriter, hist []parse.Event) bool {
	for _, ev := range hist {
		if !sw.writeEvent(ev) {
			return false
		}
	}
	return true
}

// streamLive forwards events from ch to the client and emits keepalive
// comments. Exits when ctx is canceled, ch closes, or a write fails.
func streamLive(ctx context.Context, sw *sseWriter, ch <-chan parse.Event) {
	tick := time.NewTicker(sseKeepalive)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-ch:
			if !ok || !sw.writeEvent(ev) {
				return
			}
		case <-tick.C:
			if !sw.keepalive() {
				return
			}
		}
	}
}
