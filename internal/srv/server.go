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
	"github.com/cuber-it/ccview/internal/session"
)

//go:embed static
var staticFS embed.FS

// Server hosts the embedded frontend and an SSE event stream.
type Server struct {
	mux              *http.ServeMux
	hub              *Hub
	projectsRoot     string
	projectDir       string
	currentSessionID string
}

// Config holds the project context needed to list sibling sessions.
type Config struct {
	ProjectsRoot     string // e.g. ~/.claude/projects
	ProjectDir       string // e.g. -home-ucuber-Workspace-foo
	CurrentSessionID string // full UUID of the session being viewed
}

// New constructs a Server with its own Hub. cfg may be zero-valued; in that
// case /api/sessions will return an empty list.
func New(cfg Config) *Server {
	s := &Server{
		mux:              http.NewServeMux(),
		hub:              newHub(),
		projectsRoot:     cfg.ProjectsRoot,
		projectDir:       cfg.ProjectDir,
		currentSessionID: cfg.CurrentSessionID,
	}
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(err) // programmer error: embed path wrong
	}
	s.mux.Handle("/", http.FileServer(http.FS(sub)))
	s.mux.HandleFunc("/stream", s.handleStream)
	s.mux.HandleFunc("/api/sessions", s.handleSessions)
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

// sessionDTO is the JSON shape returned by /api/sessions.
type sessionDTO struct {
	ID          string    `json:"id"`
	ShortID     string    `json:"short_id"`
	LastEvent   time.Time `json:"last_event,omitempty"`
	Size        int64     `json:"size"`
	FirstPrompt string    `json:"first_prompt,omitempty"`
	Current     bool      `json:"current"`
}

func (s *Server) handleSessions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if s.projectsRoot == "" || s.projectDir == "" {
		_ = json.NewEncoder(w).Encode([]sessionDTO{})
		return
	}
	sessions, err := session.List(s.projectsRoot, s.projectDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	out := make([]sessionDTO, 0, len(sessions))
	for _, si := range sessions {
		first, _ := session.ReadFirstUserPrompt(si.Path)
		short := si.ID
		if len(short) > 8 {
			short = short[:8]
		}
		out = append(out, sessionDTO{
			ID:          si.ID,
			ShortID:     short,
			LastEvent:   si.LastEventTime,
			Size:        si.Size,
			FirstPrompt: first,
			Current:     si.ID == s.currentSessionID,
		})
	}
	_ = json.NewEncoder(w).Encode(out)
}
