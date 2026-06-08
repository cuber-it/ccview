// Package store is ccview's central SQLite-backed metadata store: config,
// session names/favorites, per-session notes, and project-group settings.
//
// Session JSONL files live wherever Claude Code writes them (scattered by
// cwd). This single database is the clasp over that scatter: it never stores
// session content, only metadata tied to a session by ID, so a note or name
// for a session is the same no matter which directory it was opened from.
package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// Store wraps the SQLite database handle.
type Store struct{ db *sql.DB }

const schema = `
CREATE TABLE IF NOT EXISTS config (
  key   TEXT PRIMARY KEY,
  value TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS session_meta (
  session_id TEXT PRIMARY KEY,
  name       TEXT,
  favorite   INTEGER NOT NULL DEFAULT 0,
  updated_at TEXT
);
CREATE TABLE IF NOT EXISTS notes (
  session_id TEXT PRIMARY KEY,
  content    TEXT NOT NULL DEFAULT '',
  updated_at TEXT
);
CREATE TABLE IF NOT EXISTS project_groups (
  project_key TEXT PRIMARY KEY,
  label       TEXT,
  sort_order  INTEGER,
  hidden      INTEGER NOT NULL DEFAULT 0
);`

// DefaultPath returns the on-disk location of the ccview database, honouring
// CLAUDE_CONFIG_DIR the same way session.ProjectsDir does, else ~/.claude.
func DefaultPath() (string, error) {
	base := os.Getenv("CLAUDE_CONFIG_DIR")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".claude")
	}
	return filepath.Join(base, "ccview", "ccview.db"), nil
}

// Open opens the store at path, creating the parent directory and applying
// the schema (idempotent). Pass an empty path to use DefaultPath.
func Open(path string) (*Store, error) {
	if path == "" {
		var err error
		if path, err = DefaultPath(); err != nil {
			return nil, err
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("mkdir store dir: %w", err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1) // SQLite: serialise access, avoids "database is locked"
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}
	return &Store{db: db}, nil
}

// Close closes the underlying database.
func (s *Store) Close() error { return s.db.Close() }
