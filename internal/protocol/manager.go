package protocol

import (
	"context"
	"os"
	"path/filepath"
	"sync"

	"github.com/cuber-it/ccview/internal/export"
)

// Manager owns the set of running recorders, one per session being recorded.
// It is safe for concurrent use. Recorders are children of the parent context
// passed to New, so canceling it (server shutdown) stops them all.
type Manager struct {
	dir        string
	parent     context.Context
	saveMarker func(sessionID string, events, prompt int)

	mu   sync.Mutex
	recs map[string]*Recorder
}

// NewManager returns a manager writing transcripts under dir. saveMarker (may be
// nil) mirrors each recorder's resume marker to the DB for the UI and query box.
func NewManager(parent context.Context, dir string, saveMarker func(string, int, int)) *Manager {
	return &Manager{parent: parent, dir: dir, saveMarker: saveMarker, recs: map[string]*Recorder{}}
}

// HTMLPath is the on-disk location of a session's transcript.
func (m *Manager) HTMLPath(sessionID string) string {
	return filepath.Join(m.dir, sessionID+".html")
}

// Active reports whether a session is currently being recorded.
func (m *Manager) Active(sessionID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.recs[sessionID]
	return ok
}

// Start begins (or resumes) recording a session. Idempotent: starting an
// already-recording session is a no-op.
func (m *Manager) Start(sessionID, jsonlPath string, meta export.Meta) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.recs[sessionID]; ok {
		return nil
	}
	if err := os.MkdirAll(m.dir, 0o755); err != nil {
		return err
	}
	rec := newRecorder(sessionID, jsonlPath, m.HTMLPath(sessionID), meta, m.saveMarker)
	rec.start(m.parent)
	m.recs[sessionID] = rec
	return nil
}

// Stop halts recording a session and blocks until its transcript is flushed.
// Stopping a session that is not recording is a no-op.
func (m *Manager) Stop(sessionID string) {
	m.mu.Lock()
	rec, ok := m.recs[sessionID]
	delete(m.recs, sessionID)
	m.mu.Unlock()
	if ok {
		rec.stop()
	}
}

// StopAll halts every recorder (server shutdown).
func (m *Manager) StopAll() {
	m.mu.Lock()
	recs := m.recs
	m.recs = map[string]*Recorder{}
	m.mu.Unlock()
	for _, r := range recs {
		r.stop()
	}
}
