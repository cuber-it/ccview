// Package srv serves the ccview frontend and streams session events via SSE.
package srv

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cuber-it/ccview/internal/export"
	"github.com/cuber-it/ccview/internal/parse"
	"github.com/cuber-it/ccview/internal/session"
	"github.com/cuber-it/ccview/internal/tail"
)

//go:embed static
var staticFS embed.FS

const defaultExportDir = "Workspace/claude-code/sessions"

type Config struct {
	ProjectsRoot string
	ProjectDir   string
	Version      string
	Verbose      bool
}

type Server struct {
	mux          *http.ServeMux
	hub          *Hub
	projectsRoot string
	projectDir   string
	version      string
	verbose      bool

	mu         sync.Mutex
	currentID  string
	pumpCancel context.CancelFunc
	rootCtx    context.Context
}

func New(cfg Config) *Server {
	s := &Server{
		mux:          http.NewServeMux(),
		hub:          newHub(),
		projectsRoot: cfg.ProjectsRoot,
		projectDir:   cfg.ProjectDir,
		version:      cfg.Version,
		verbose:      cfg.Verbose,
	}
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(err)
	}
	s.mux.Handle("/", http.FileServer(http.FS(sub)))
	s.mux.HandleFunc("/stream", s.handleStream)
	s.mux.HandleFunc("/api/sessions", s.handleSessions)
	s.mux.HandleFunc("/api/switch", s.handleSwitch)
	s.mux.HandleFunc("/api/export", s.handleExport)
	s.mux.HandleFunc("/api/version", s.handleVersion)
	return s
}

func (s *Server) Hub() *Hub { return s.hub }

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.mux.ServeHTTP(w, r) }

// Serve runs the HTTP server until ctx is canceled. If initial is non-nil,
// that session is opened before serving begins.
func (s *Server) Serve(ctx context.Context, ln net.Listener, initial *session.Info) error {
	s.mu.Lock()
	s.rootCtx = ctx
	s.mu.Unlock()
	if initial != nil {
		if err := s.SetSession(*initial); err != nil {
			return err
		}
	}
	srv := &http.Server{
		Handler:     s.mux,
		ReadTimeout: 10 * time.Second,
		IdleTimeout: 120 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutCtx)
	}()
	if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// SetSession replaces the active tailer and resets all subscribed clients.
func (s *Server) SetSession(info session.Info) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.rootCtx == nil {
		return errors.New("server not serving yet")
	}
	if s.pumpCancel != nil {
		s.pumpCancel()
	}
	s.hub.Reset()
	s.currentID = info.ID
	pumpCtx, cancel := context.WithCancel(s.rootCtx)
	s.pumpCancel = cancel
	go s.pump(pumpCtx, info.Path)
	return nil
}

func (s *Server) pump(ctx context.Context, path string) {
	for line := range tail.New(path).Stream(ctx) {
		if line.Err != nil {
			if s.verbose {
				fmt.Fprintln(os.Stderr, "ccview: tail:", line.Err)
			}
			return
		}
		ev, err := parse.Parse(line.Data)
		if err != nil {
			continue
		}
		s.hub.Publish(ev)
	}
}

// ---- handlers ----

func (s *Server) handleVersion(w http.ResponseWriter, _ *http.Request) {
	v := s.version
	if v == "" {
		v = "dev"
	}
	writeJSON(w, map[string]string{"version": v})
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

	hist, ch, unsub := s.hub.Subscribe()
	defer unsub()

	// Initial comment opens the stream for the browser even with empty history.
	if _, err := fmt.Fprint(w, ": ok\n\n"); err != nil {
		return
	}
	flusher.Flush()

	write := func(ev parse.Event) bool {
		if ev.Kind == kindReset {
			if _, err := fmt.Fprint(w, "event: reset\ndata: {}\n\n"); err != nil {
				return false
			}
			flusher.Flush()
			return true
		}
		data, err := json.Marshal(ev)
		if err != nil {
			return true
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

	ctx := r.Context()
	keepalive := time.NewTicker(20 * time.Second)
	defer keepalive.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-ch:
			if !ok || !write(ev) {
				return
			}
		case <-keepalive.C:
			if _, err := fmt.Fprint(w, ": keepalive\n\n"); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

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

func (s *Server) handleSessions(w http.ResponseWriter, _ *http.Request) {
	if s.projectsRoot == "" {
		writeJSON(w, []sessionDTO{})
		return
	}
	sessions, err := session.ListAll(s.projectsRoot)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.mu.Lock()
	curID := s.currentID
	s.mu.Unlock()

	out := make([]sessionDTO, 0, len(sessions))
	for _, si := range sessions {
		first, _ := session.ReadFirstUserPrompt(si.Path)
		short := si.ID
		if len(short) > 8 {
			short = short[:8]
		}
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
	writeJSON(w, out)
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
	writeJSON(w, map[string]string{"id": info.ID, "project": info.Project})
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
	curID := s.currentID
	projectDir := s.projectDir
	s.mu.Unlock()

	var started time.Time
	if s.projectsRoot != "" && projectDir != "" {
		if list, err := session.List(s.projectsRoot, projectDir); err == nil {
			for _, si := range list {
				if si.ID == curID {
					started = si.FirstEventTime
					break
				}
			}
		}
	}

	target, err := resolveExportPath(body.Path, curID, projectDir, started)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		http.Error(w, "mkdir: "+err.Error(), http.StatusInternalServerError)
		return
	}
	events := s.hub.History()
	md := export.Markdown(export.Meta{
		SessionID:   curID,
		ProjectPath: decodeProjectDir(projectDir),
		Started:     started,
		Exported:    time.Now(),
	}, events)
	if err := os.WriteFile(target, []byte(md), 0o644); err != nil {
		http.Error(w, "write: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{
		"path":   target,
		"bytes":  len(md),
		"events": len(events),
	})
}

// ---- helpers ----

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func resolveExportPath(userPath, sessionID, projectDir string, started time.Time) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	defaultDir := filepath.Join(home, defaultExportDir)
	target := userPath
	if target == "" {
		target = defaultFilename(defaultDir, sessionID, projectDir, started)
	}
	if strings.HasPrefix(target, "~/") {
		target = filepath.Join(home, target[2:])
	}
	if !filepath.IsAbs(target) && !strings.ContainsRune(target, filepath.Separator) {
		target = filepath.Join(defaultDir, target)
	}
	if filepath.Ext(target) == "" {
		target += ".md"
	}
	return target, nil
}

func defaultFilename(dir, sessionID, projectDir string, started time.Time) string {
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
	return filepath.Join(dir, name)
}

func projectLabel(raw string) string {
	path := strings.ReplaceAll(raw, "-", "/")
	parts := strings.Split(path, "/")
	if len(parts) <= 3 {
		return path
	}
	return ".../" + strings.Join(parts[len(parts)-3:], "/")
}

// probeRealPath greedily rebuilds a filesystem path from the hyphen-encoded
// project-dir name (Claude Code replaces `/` with `-`, lossy if a real dir
// name contains a `-`). Returns matched segments and the deepest real path.
func probeRealPath(encoded string) (segments []string, full string) {
	encoded = strings.TrimLeft(encoded, "-")
	if encoded == "" {
		return nil, ""
	}
	parts := strings.Split(encoded, "-")
	prefix := string(filepath.Separator)
	for i := 0; i < len(parts); {
		matched := false
		for j := len(parts); j > i; j-- {
			cand := strings.Join(parts[i:j], "-")
			if cand == "" {
				continue
			}
			test := filepath.Join(prefix, cand)
			if stat, err := os.Stat(test); err == nil && stat.IsDir() {
				segments = append(segments, cand)
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
	return segments, prefix
}

func projectShortName(encoded string) string {
	segs, _ := probeRealPath(encoded)
	switch len(segs) {
	case 0:
		parts := strings.Split(strings.TrimLeft(encoded, "-"), "-")
		if len(parts) == 0 || parts[0] == "" {
			return ""
		}
		return parts[len(parts)-1]
	case 1:
		return segs[0]
	default:
		return segs[len(segs)-2] + "-" + segs[len(segs)-1]
	}
}

func decodeProjectDir(encoded string) string {
	if encoded == "" {
		return ""
	}
	_, full := probeRealPath(encoded)
	if full == string(filepath.Separator) {
		return "/" + strings.ReplaceAll(strings.TrimLeft(encoded, "-"), "-", "/")
	}
	return full
}
