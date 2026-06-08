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
	"github.com/cuber-it/ccview/internal/store"
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
	store        *store.Store
	projectsRoot string
	roots        []string
	projectDir   string
	version      string
	verbose      bool

	mu         sync.Mutex
	currentID  string
	pumpCancel context.CancelFunc
	rootCtx    context.Context
}

// noCache disables browser caching — used in CCVIEW_DEV mode so a reload
// always fetches the latest static file.
func noCache(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		h.ServeHTTP(w, r)
	})
}

func New(cfg Config) *Server {
	st, err := store.Open("")
	if err != nil {
		panic(fmt.Errorf("open ccview store: %w", err))
	}
	// Migrate the legacy file-based state into the DB once (idempotent).
	legacyRoots, _ := rootsConfigPath()
	if mErr := st.MigrateFromFiles(legacyRoots, legacyNotesDir()); mErr != nil && cfg.Verbose {
		fmt.Fprintf(os.Stderr, "ccview: migration warning: %v\n", mErr)
	}
	roots, ok, _ := st.GetRoots()
	if (!ok || len(roots) == 0) && cfg.ProjectsRoot != "" {
		roots = []string{cfg.ProjectsRoot}
	}
	s := &Server{
		mux:          http.NewServeMux(),
		hub:          newHub(),
		store:        st,
		projectsRoot: cfg.ProjectsRoot,
		roots:        roots,
		projectDir:   cfg.ProjectDir,
		version:      cfg.Version,
		verbose:      cfg.Verbose,
	}
	// Frontend: embedded by default. CCVIEW_DEV=<dir> serves static from disk
	// for live editing — change a file, reload the browser, no rebuild needed.
	var staticHandler http.Handler
	if dir := os.Getenv("CCVIEW_DEV"); dir != "" {
		staticHandler = noCache(http.FileServer(http.Dir(dir)))
	} else {
		sub, err := fs.Sub(staticFS, "static")
		if err != nil {
			panic(err)
		}
		staticHandler = http.FileServer(http.FS(sub))
	}
	s.mux.Handle("/", staticHandler)
	s.mux.HandleFunc("/stream", s.handleStream)
	s.mux.HandleFunc("/api/sessions", s.handleSessions)
	s.mux.HandleFunc("/api/switch", s.handleSwitch)
	s.mux.HandleFunc("/api/export", s.handleExport)
	s.mux.HandleFunc("/api/version", s.handleVersion)
	s.mux.HandleFunc("/api/notes", s.handleNotes)
	s.mux.HandleFunc("/api/roots", s.handleRoots)
	s.mux.HandleFunc("/api/session-meta", s.handleSessionMeta)
	s.mux.HandleFunc("/api/groups", s.handleGroups)
	s.mux.HandleFunc("/api/search", s.handleSearch)
	return s
}

// currentRoots returns a copy of the configured projects roots (thread-safe).
func (s *Server) currentRoots() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.roots...)
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
	sw, ok := newSSEWriter(w)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	hist, ch, unsub := s.hub.Subscribe()
	defer unsub()
	if !sw.open() {
		return
	}
	if !replayHistory(sw, hist) {
		return
	}
	streamLive(r.Context(), sw, ch)
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
	Name         string    `json:"name,omitempty"`
	Favorite     bool      `json:"favorite"`
	Done         bool      `json:"done"`
}

func (s *Server) handleSessions(w http.ResponseWriter, _ *http.Request) {
	roots := s.currentRoots()
	if len(roots) == 0 {
		writeJSON(w, []sessionDTO{})
		return
	}
	sessions := session.ListAllRoots(roots)
	meta, _ := s.store.AllMeta()
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
			Name:         meta[si.ID].Name,
			Favorite:     meta[si.ID].Favorite,
			Done:         meta[si.ID].Done,
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
	roots := s.currentRoots()
	if len(roots) == 0 {
		http.Error(w, "no projects configured", http.StatusBadRequest)
		return
	}
	sessions := session.ListAllRoots(roots)
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
		Path    string `json:"path"`
		Session string `json:"session"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	s.mu.Lock()
	curID := s.currentID
	projectDir := s.projectDir
	s.mu.Unlock()

	sessID := curID
	var events []parse.Event
	var started time.Time
	if body.Session != "" && body.Session != curID {
		// Export an arbitrary session: locate its file across all roots and parse it.
		info, ok := s.findSession(body.Session)
		if !ok {
			http.Error(w, "session not found: "+body.Session, http.StatusBadRequest)
			return
		}
		evs, perr := parseSessionFile(info.Path)
		if perr != nil {
			http.Error(w, perr.Error(), http.StatusInternalServerError)
			return
		}
		sessID, projectDir, events, started = body.Session, info.Project, evs, info.FirstEventTime
	} else {
		events = s.hub.History()
		if info, ok := s.findSession(sessID); ok {
			started = info.FirstEventTime
		}
	}

	target, err := resolveExportPath(body.Path, sessID, projectDir, started)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	meta := export.Meta{
		SessionID:   sessID,
		ProjectPath: decodeProjectDir(projectDir),
		Started:     started,
		Exported:    time.Now(),
	}
	n, err := writeMarkdownExport(target, meta, events)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{
		"path":   target,
		"bytes":  n,
		"events": len(events),
	})
}

// legacyNotesDir is the old file-based notes location, kept only as a one-time
// migration source into the DB.
func legacyNotesDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Workspace", "claude-code", "notes")
}

// handleNotes serves GET (load) and POST (save) of a per-session note, backed
// by the central store.
func (s *Server) handleNotes(w http.ResponseWriter, r *http.Request) {
	sess := r.URL.Query().Get("session")
	if sess == "" || strings.ContainsAny(sess, "/\\") || strings.Contains(sess, "..") {
		http.Error(w, "invalid session", http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodGet:
		content, err := s.store.GetNote(sess)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]any{"content": content})
	case http.MethodPost:
		var body struct {
			Content string `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := s.store.SetNote(sess, body.Content); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]any{"ok": true})
	default:
		http.Error(w, "GET or POST required", http.StatusMethodNotAllowed)
	}
}

// handleSessionMeta sets a session's custom name and/or favorite flag. Fields
// left out of the JSON body (nil) are not touched.
func (s *Server) handleSessionMeta(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Session  string  `json:"session"`
		Name     *string `json:"name"`
		Favorite *bool   `json:"favorite"`
		Done     *bool   `json:"done"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if body.Session == "" {
		http.Error(w, "session required", http.StatusBadRequest)
		return
	}
	if body.Name != nil {
		if err := s.store.SetName(body.Session, strings.TrimSpace(*body.Name)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if body.Favorite != nil {
		if err := s.store.SetFavorite(body.Session, *body.Favorite); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if body.Done != nil {
		if err := s.store.SetDone(body.Session, *body.Done); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	writeJSON(w, map[string]any{"ok": true})
}

// handleGroups returns (GET) or replaces (POST) the project-group display config.
func (s *Server) handleGroups(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		groups, err := s.store.Groups()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if groups == nil {
			groups = []store.Group{}
		}
		writeJSON(w, map[string]any{"groups": groups})
	case http.MethodPost:
		var body struct {
			Groups []store.Group `json:"groups"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := s.store.SaveGroups(body.Groups); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]any{"ok": true})
	default:
		http.Error(w, "GET or POST required", http.StatusMethodNotAllowed)
	}
}

// parseSessionFile reads a full session JSONL and parses it into events,
// skipping blank or malformed lines so one bad line never aborts the export.
func parseSessionFile(path string) ([]parse.Event, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out []parse.Event
	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if ev, perr := parse.Parse([]byte(line)); perr == nil {
			out = append(out, ev)
		}
	}
	return out, nil
}

// findSession locates a session by ID across all configured roots.
func (s *Server) findSession(sessionID string) (session.Info, bool) {
	for _, si := range session.ListAllRoots(s.currentRoots()) {
		if si.ID == sessionID {
			return si, true
		}
	}
	return session.Info{}, false
}

// handleRoots returns (GET) or replaces (POST) the list of projects roots.
// On POST the new list is persisted to disk *before* the in-memory state is
// updated, so a failed write leaves runtime and config consistent.
func (s *Server) handleRoots(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, map[string]any{"roots": s.currentRoots()})
	case http.MethodPost:
		var body struct {
			Roots []string `json:"roots"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		clean := cleanRoots(body.Roots)
		if err := s.store.SetRoots(clean); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		s.mu.Lock()
		s.roots = clean
		s.mu.Unlock()
		writeJSON(w, map[string]any{"roots": clean})
	default:
		http.Error(w, "GET or POST required", http.StatusMethodNotAllowed)
	}
}

// cleanRoots trims, expands a leading ~, drops blanks, and deduplicates.
func cleanRoots(in []string) []string {
	out := make([]string, 0, len(in))
	seen := make(map[string]bool)
	home, _ := os.UserHomeDir()
	for _, root := range in {
		root = strings.TrimSpace(root)
		if root == "" {
			continue
		}
		if home != "" && (root == "~" || strings.HasPrefix(root, "~/")) {
			root = filepath.Join(home, strings.TrimPrefix(root[1:], "/"))
		}
		if seen[root] {
			continue
		}
		seen[root] = true
		out = append(out, root)
	}
	return out
}

// rootsConfigPath is where the projects-root list is persisted.
func rootsConfigPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "ccview", "roots.json"), nil
}

// writeMarkdownExport renders events as Markdown, ensures the parent dir
// exists, writes target, and returns the byte count.
func writeMarkdownExport(target string, meta export.Meta, events []parse.Event) (int, error) {
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return 0, fmt.Errorf("mkdir: %w", err)
	}
	md := export.Markdown(meta, events)
	if err := os.WriteFile(target, []byte(md), 0o644); err != nil {
		return 0, fmt.Errorf("write: %w", err)
	}
	return len(md), nil
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
