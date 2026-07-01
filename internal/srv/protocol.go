package srv

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cuber-it/ccview/internal/export"
	"github.com/cuber-it/ccview/internal/protocol"
	"github.com/cuber-it/ccview/internal/store"
)

// initProtocols creates the recorder manager and resumes recording for every
// session still flagged active in the DB (survives a server/systemd restart).
func (s *Server) initProtocols(ctx context.Context) {
	saveMarker := func(sessionID string, events, prompt int) {
		if err := s.store.SetProtocolMarker(sessionID, events, prompt); err != nil && s.verbose {
			fmt.Fprintf(os.Stderr, "ccview: protocol marker: %v\n", err)
		}
	}
	mgr := protocol.NewManager(ctx, protocolsDir(), saveMarker)
	s.mu.Lock()
	s.protocols = mgr
	s.mu.Unlock()

	active, err := s.store.ActiveProtocols()
	if err != nil {
		if s.verbose {
			fmt.Fprintf(os.Stderr, "ccview: protocol resume: %v\n", err)
		}
		return
	}
	for _, id := range active {
		info, ok := s.findSession(id)
		if !ok {
			continue // session gone; leave the flag, nothing to record
		}
		if err := mgr.Start(id, info.Path, s.protocolMeta(id, info.Project, info.FirstEventTime)); err != nil && s.verbose {
			fmt.Fprintf(os.Stderr, "ccview: protocol resume %s: %v\n", id, err)
		}
	}
}

// handleProtocol toggles a session's tail-html recording (POST {on}) or serves
// the growing transcript (GET) for opening in a browser tab.
func (s *Server) handleProtocol(w http.ResponseWriter, r *http.Request) {
	sess := r.URL.Query().Get("session")
	if sess == "" || strings.ContainsAny(sess, "/\\") || strings.Contains(sess, "..") {
		http.Error(w, "invalid session", http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodGet:
		s.serveProtocolFile(w, r, sess)
	case http.MethodPost:
		var body struct {
			On bool `json:"on"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := s.setProtocol(sess, body.On); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]any{"active": body.On})
	default:
		http.Error(w, "GET or POST required", http.StatusMethodNotAllowed)
	}
}

// setProtocol starts or stops a recorder and persists the active flag, leaving
// the resume marker untouched so turning it back on continues the same file.
func (s *Server) setProtocol(sess string, on bool) error {
	if s.protocols == nil {
		return fmt.Errorf("recorder not ready")
	}
	if on {
		info, ok := s.findSession(sess)
		if !ok {
			return fmt.Errorf("session not found")
		}
		if err := s.protocols.Start(sess, info.Path, s.protocolMeta(sess, info.Project, info.FirstEventTime)); err != nil {
			return err
		}
		return s.store.SetProtocolActive(sess, true)
	}
	s.protocols.Stop(sess)
	return s.store.SetProtocolActive(sess, false)
}

// serveProtocolFile streams a session's transcript, closing the still-open
// document on the fly (the file itself stays append-friendly while recording).
func (s *Server) serveProtocolFile(w http.ResponseWriter, r *http.Request, sess string) {
	if s.protocols == nil {
		http.Error(w, "recorder not ready", http.StatusServiceUnavailable)
		return
	}
	data, err := os.ReadFile(s.protocols.HTMLPath(sess))
	if err != nil {
		http.Error(w, "no protocol recorded for this session", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Write(data)
	// While recording, make the served document tail itself: reload only while
	// the reader is at the bottom, so scrolling up to read pauses the refresh.
	if s.protocols.Active(sess) {
		fmt.Fprint(w, tailScript(s.protocolRefreshSecs(r)))
	}
	fmt.Fprint(w, export.NewHTMLRenderer().Foot())
}

// defaultProtocolRefresh is the auto-refresh interval (seconds) used when no
// config value and no query override are given.
const defaultProtocolRefresh = 3

// protocolRefreshSecs resolves the auto-refresh interval for a served
// transcript: the DB config default (protocol_refresh_secs, else 3), overridden
// by a ?refresh= query value when present. A value <= 0 disables auto-refresh.
func (s *Server) protocolRefreshSecs(r *http.Request) int {
	secs := defaultProtocolRefresh
	if v, ok, _ := s.store.GetConfig("protocol_refresh_secs"); ok {
		secs = atoiDefault(v, secs)
	}
	if q := r.URL.Query().Get("refresh"); q != "" {
		secs = atoiDefault(q, secs)
	}
	return secs
}

// tailScript returns a "tail -f" script that reloads the transcript every secs
// seconds, but only while the reader is at the bottom (scrolling up to read
// pauses it). Returns "" when secs <= 0, disabling auto-refresh entirely.
func tailScript(secs int) string {
	if secs <= 0 {
		return ""
	}
	return fmt.Sprintf(`<script>
(function(){
  function atBottom(){ return window.innerHeight + window.scrollY >= document.body.scrollHeight - 60; }
  if (sessionStorage.getItem('ccviewTail') === '1') window.scrollTo(0, document.body.scrollHeight);
  setInterval(function(){
    if (atBottom()) { sessionStorage.setItem('ccviewTail','1'); location.reload(); }
    else { sessionStorage.setItem('ccviewTail','0'); }
  }, %d);
})();
</script>
`, secs*1000)
}

// protocolMeta builds the transcript header metadata for a session.
func (s *Server) protocolMeta(sessionID, project string, started time.Time) export.Meta {
	return export.Meta{
		SessionID:   sessionID,
		ProjectPath: decodeProjectDir(project),
		Started:     started,
	}
}

// protocolsDir is the on-disk directory for transcripts, next to the DB.
func protocolsDir() string {
	db, err := store.DefaultPath()
	if err != nil {
		return "protocols"
	}
	return filepath.Join(filepath.Dir(db), "protocols")
}
