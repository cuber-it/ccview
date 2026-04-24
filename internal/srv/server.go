// Package srv provides an HTTP server that hosts the ccview frontend
// and streams session events via SSE to connected browsers.
package srv

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"time"

	"github.com/cuber-it/ccview/internal/parse"
)

//go:embed static
var staticFS embed.FS

// Server hosts the embedded frontend and an SSE event stream.
type Server struct {
	mux *http.ServeMux
	hub *Hub
}

// New constructs a Server with its own Hub.
func New() *Server {
	s := &Server{
		mux: http.NewServeMux(),
		hub: newHub(),
	}
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(err) // programmer error: embed path wrong
	}
	s.mux.Handle("/", http.FileServer(http.FS(sub)))
	s.mux.HandleFunc("/stream", s.handleStream)
	return s
}

// Hub returns the Server's event hub. Publish events here.
func (s *Server) Hub() *Hub { return s.hub }

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// Serve runs the HTTP server on ln until ctx is canceled.
func (s *Server) Serve(ctx context.Context, ln net.Listener) error {
	srv := &http.Server{
		Handler:      s.mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 0, // SSE needs long-lived writes
		IdleTimeout:  120 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()
	if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) handleStream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	h := w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache")
	h.Set("Connection", "keep-alive")
	h.Set("X-Accel-Buffering", "no")

	hist, ch, unsubscribe := s.hub.Subscribe()
	defer unsubscribe()

	write := func(ev parse.Event) bool {
		data, err := json.Marshal(ev)
		if err != nil {
			return true // skip bad event, don't tear down
		}
		if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
			return false
		}
		flusher.Flush()
		return true
	}

	for _, ev := range hist {
		if !write(ev) {
			return
		}
	}
	flusher.Flush()

	ctx := r.Context()
	keepalive := time.NewTicker(20 * time.Second)
	defer keepalive.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-ch:
			if !ok {
				return
			}
			if !write(ev) {
				return
			}
		case <-keepalive.C:
			if _, err := fmt.Fprintf(w, ": keepalive\n\n"); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}
