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
	"os"
	"strings"
	"sync"
	"time"

	"path/filepath"

	"github.com/cuber-it/ccview/internal/export"
	"github.com/cuber-it/ccview/internal/parse"
	"github.com/cuber-it/ccview/internal/session"
	"github.com/cuber-it/ccview/internal/tail"
)

//go:embed static
var staticFS embed.FS

// Server hosts the embedded frontend and an SSE event stream.
// It owns the tailer-pump for the currently-viewed session and can swap it
// at runtime via SetSession.
type Server struct {
	mux *http.ServeMux
	hub *Hub

	projectsRoot string
	projectDir   string

	mu               sync.Mutex
	currentSessionID string
	pumpCancel       context.CancelFunc
	rootCtx          context.Context
}

// Config holds the project context needed to list sibling sessions.
type Config struct {
	ProjectsRoot     string // e.g. ~/.claude/projects
	ProjectDir       string // e.g. -home-ucuber-Workspace-foo
	CurrentSessionID string // full UUID of the initially-viewed session
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
	s.mux.HandleFunc("/api/switch", s.handleSwitch)
	s.mux.HandleFunc("/api/export", s.handleExport)
	return s
}

// SetSession stops the current tailer (if any) and starts a new one on path.
// Connected clients receive a reset signal and the new session's events.
// Must be called after Serve so the server has a lifetime context.
func (s *Server) SetSession(info session.Info) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.rootCtx == nil {
		return fmt.Errorf("Server not yet serving (call Serve before SetSession)")
	}
	if s.pumpCancel != nil {
		s.pumpCancel()
	}
	s.hub.Reset()
	s.currentSessionID = info.ID

	pumpCtx, cancel := context.WithCancel(s.rootCtx)
	s.pumpCancel = cancel
	go s.pump(pumpCtx, info.Path)
	return nil
}

func (s *Server) pump(ctx context.Context, path string) {
	ch := tail.New(path).Stream(ctx)
	for l := range ch {
		if l.Err != nil {
			fmt.Fprintln(os.Stderr, "ccview: tail:", l.Err)
			return
		}
		ev, err := parse.Parse(l.Data)
		if err != nil {
			continue
		}
		s.hub.Publish(ev)
	}
}

// Hub returns the Server's event hub. Publish events here.
func (s *Server) Hub() *Hub { return s.hub }

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// Serve runs the HTTP server on ln until ctx is canceled.
// The context becomes the lifetime for any tailer-pumps started via SetSession.
func (s *Server) Serve(ctx context.Context, ln net.Listener) error {
	s.mu.Lock()
	s.rootCtx = ctx
	s.mu.Unlock()

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
		if ev.Kind == kindReset {
			if _, err := fmt.Fprintf(w, "event: reset\ndata: {}\n\n"); err != nil {
				return false
			}
			flusher.Flush()
			return true
		}
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
	ID           string    `json:"id"`
	ShortID      string    `json:"short_id"`
	Project      string    `json:"project"`
	ProjectLabel string    `json:"project_label"`
	FirstEvent   time.Time `json:"first_event,omitempty"`
	LastEvent    time.Time `json:"last_event,omitempty"`
	Size         int64     `json:"size"`
	FirstPrompt  string    `json:"first_prompt,omitempty"`
	Current      bool      `json:"current"`
	SameProject  bool      `json:"same_project"`
}

// projectLabel turns "-home-ucuber-Workspace-cuber-it-sps-sim-go"
// into ".../cuber-it/sps-sim/go" for readable display.
func projectLabel(raw string) string {
	// Reverse the Claude Code encoding: "-" → "/".
	path := strings.ReplaceAll(raw, "-", "/")
	// Trim the last 3 path components for a compact label.
	parts := strings.Split(path, "/")
	if len(parts) <= 3 {
		return path
	}
	return ".../" + strings.Join(parts[len(parts)-3:], "/")
}

// defaultExportDir is where "Save" (no explicit path) writes to.
const defaultExportDir = "Workspace/claude-code/sessions"

// defaultExportPath returns ~/Workspace/claude-code/sessions/<proj>_<date>_<shortid>.md
// based on the currently-viewed session and its header info.
func defaultExportPath(sessionID, projectDir string, started time.Time) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	base := filepath.Join(home, defaultExportDir)
	short := sessionID
	if len(short) > 8 {
		short = short[:8]
	}
	proj := projectShortName(projectDir)
	if proj == "" {
		proj = "project"
	}
	date := time.Now()
	if !started.IsZero() {
		date = started
	}
	name := fmt.Sprintf("%s_%s_%s.md", proj, date.Local().Format("2006-01-02"), short)
	return filepath.Join(base, name), nil
}

// projectShortName reverses Claude Code's lossy `/`→`-` encoding by probing the
// filesystem greedily for existing directories, then returns the last 2 real
// path segments joined by "-". Falls back to the last hyphen-segment if no
// directory lookup succeeds.
func projectShortName(encoded string) string {
	encoded = strings.TrimLeft(encoded, "-")
	if encoded == "" {
		return ""
	}
	parts := strings.Split(encoded, "-")
	prefix := string(filepath.Separator)
	real := []string{}
	i := 0
	for i < len(parts) {
		matched := false
		for j := len(parts); j > i; j-- {
			candidate := strings.Join(parts[i:j], "-")
			if candidate == "" {
				continue
			}
			test := filepath.Join(prefix, candidate)
			if stat, err := os.Stat(test); err == nil && stat.IsDir() {
				real = append(real, candidate)
				prefix = test
				i = j
				matched = true
				break
			}
		}
		if !matched {
			i++ // skip this part; shouldn't happen for real sessions
		}
	}
	switch len(real) {
	case 0:
		return parts[len(parts)-1]
	case 1:
		return real[0]
	default:
		return real[len(real)-2] + "-" + real[len(real)-1]
	}
}

func (s *Server) handleExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Path string `json:"path"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	s.mu.Lock()
	curID := s.currentSessionID
	curProject := s.projectDir
	s.mu.Unlock()

	// Try to enrich meta by looking up the session in the project list.
	var started time.Time
	var projectPath string
	if s.projectsRoot != "" && curProject != "" {
		if list, err := session.List(s.projectsRoot, curProject); err == nil {
			for _, si := range list {
				if si.ID == curID {
					started = si.FirstEventTime
					break
				}
			}
		}
	}
	projectPath = decodeProjectDir(curProject)

	target := body.Path
	if target == "" {
		p, err := defaultExportPath(curID, curProject, started)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		target = p
	}
	// Expand leading ~
	if strings.HasPrefix(target, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			target = filepath.Join(home, target[2:])
		}
	}
	// Relative path without extension → put in default dir
	if !filepath.IsAbs(target) && !strings.ContainsRune(target, filepath.Separator) {
		home, _ := os.UserHomeDir()
		target = filepath.Join(home, defaultExportDir, target)
	}
	if filepath.Ext(target) == "" {
		target += ".md"
	}

	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		http.Error(w, "mkdir: "+err.Error(), http.StatusInternalServerError)
		return
	}

	events := s.hub.History()
	md := export.Markdown(export.Meta{
		SessionID:   curID,
		ProjectPath: projectPath,
		Started:     started,
		Exported:    time.Now(),
	}, events)

	if err := os.WriteFile(target, []byte(md), 0o644); err != nil {
		http.Error(w, "write: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"path":   target,
		"bytes":  len(md),
		"events": len(events),
	})
}

// decodeProjectDir best-effort reconstructs a real filesystem path from the
// encoded project-dir name by probing existing directories.
func decodeProjectDir(raw string) string {
	if raw == "" {
		return ""
	}
	raw = strings.TrimLeft(raw, "-")
	parts := strings.Split(raw, "-")
	prefix := string(filepath.Separator)
	i := 0
	for i < len(parts) {
		matched := false
		for j := len(parts); j > i; j-- {
			candidate := strings.Join(parts[i:j], "-")
			if candidate == "" {
				continue
			}
			test := filepath.Join(prefix, candidate)
			if stat, err := os.Stat(test); err == nil && stat.IsDir() {
				prefix = test
				i = j
				matched = true
				break
			}
		}
		if !matched {
			i++
		}
	}
	if prefix == string(filepath.Separator) {
		return "/" + strings.ReplaceAll(raw, "-", "/")
	}
	return prefix
}

func (s *Server) handleSwitch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if s.projectsRoot == "" {
		http.Error(w, "no projects configured", http.StatusBadRequest)
		return
	}
	sessions, err := session.ListAll(s.projectsRoot)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	info, err := session.Resolve(sessions, body.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err := s.SetSession(info); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"id": info.ID, "project": info.Project})
}

func (s *Server) handleSessions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if s.projectsRoot == "" {
		_ = json.NewEncoder(w).Encode([]sessionDTO{})
		return
	}
	sessions, err := session.ListAll(s.projectsRoot)
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
		s.mu.Lock()
		curID := s.currentSessionID
		s.mu.Unlock()
		out = append(out, sessionDTO{
			ID:           si.ID,
			ShortID:      short,
			Project:      si.Project,
			ProjectLabel: projectLabel(si.Project),
			FirstEvent:   si.FirstEventTime,
			LastEvent:    si.LastEventTime,
			Size:         si.Size,
			FirstPrompt:  first,
			Current:      si.ID == curID,
			SameProject:  si.Project == s.projectDir,
		})
	}
	_ = json.NewEncoder(w).Encode(out)
}
