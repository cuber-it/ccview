package store

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func now() string { return time.Now().UTC().Format(time.RFC3339) }

// --- config (key/value) ---

// GetConfig returns the value for key and whether it existed.
func (s *Store) GetConfig(key string) (string, bool, error) {
	var v string
	err := s.db.QueryRow(`SELECT value FROM config WHERE key=?`, key).Scan(&v)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	return v, err == nil, err
}

// SetConfig upserts a config value.
func (s *Store) SetConfig(key, value string) error {
	_, err := s.db.Exec(
		`INSERT INTO config(key,value) VALUES(?,?)
		 ON CONFLICT(key) DO UPDATE SET value=excluded.value`, key, value)
	return err
}

// --- roots (stored as a JSON array under config key "roots") ---

// GetRoots returns the configured projects roots, or (nil,false) if unset.
func (s *Store) GetRoots() ([]string, bool, error) {
	v, ok, err := s.GetConfig("roots")
	if err != nil || !ok {
		return nil, ok, err
	}
	var roots []string
	if err := json.Unmarshal([]byte(v), &roots); err != nil {
		return nil, true, err
	}
	return roots, true, nil
}

// SetRoots persists the projects roots.
func (s *Store) SetRoots(roots []string) error {
	b, err := json.Marshal(roots)
	if err != nil {
		return err
	}
	return s.SetConfig("roots", string(b))
}

// --- notes (one per session) ---

// GetNote returns a session's note, or "" if none.
func (s *Store) GetNote(sessionID string) (string, error) {
	var c string
	err := s.db.QueryRow(`SELECT content FROM notes WHERE session_id=?`, sessionID).Scan(&c)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return c, err
}

// SetNote upserts a session's note.
func (s *Store) SetNote(sessionID, content string) error {
	_, err := s.db.Exec(
		`INSERT INTO notes(session_id,content,updated_at) VALUES(?,?,?)
		 ON CONFLICT(session_id) DO UPDATE SET content=excluded.content, updated_at=excluded.updated_at`,
		sessionID, content, now())
	return err
}

// --- session meta (name, favorite) ---

// Meta is the per-session metadata shown in the sidebar.
type Meta struct {
	Name     string `json:"name"`
	Favorite bool   `json:"favorite"`
}

// AllMeta returns every session's metadata, keyed by session ID.
func (s *Store) AllMeta() (map[string]Meta, error) {
	rows, err := s.db.Query(`SELECT session_id, COALESCE(name,''), favorite FROM session_meta`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[string]Meta)
	for rows.Next() {
		var id, name string
		var fav int
		if err := rows.Scan(&id, &name, &fav); err != nil {
			return nil, err
		}
		out[id] = Meta{Name: name, Favorite: fav != 0}
	}
	return out, rows.Err()
}

// SetName sets (or clears, when name is empty) a session's custom name.
func (s *Store) SetName(sessionID, name string) error {
	_, err := s.db.Exec(
		`INSERT INTO session_meta(session_id,name,updated_at) VALUES(?,?,?)
		 ON CONFLICT(session_id) DO UPDATE SET name=excluded.name, updated_at=excluded.updated_at`,
		sessionID, name, now())
	return err
}

// SetFavorite toggles a session's favorite flag.
func (s *Store) SetFavorite(sessionID string, fav bool) error {
	f := 0
	if fav {
		f = 1
	}
	_, err := s.db.Exec(
		`INSERT INTO session_meta(session_id,favorite,updated_at) VALUES(?,?,?)
		 ON CONFLICT(session_id) DO UPDATE SET favorite=excluded.favorite, updated_at=excluded.updated_at`,
		sessionID, f, now())
	return err
}

// --- project groups ---

// Group is a sidebar project group's display config.
type Group struct {
	Key    string `json:"key"`
	Label  string `json:"label"`
	Order  int    `json:"order"`
	Hidden bool   `json:"hidden"`
}

// Groups returns all stored project-group configs.
func (s *Store) Groups() ([]Group, error) {
	rows, err := s.db.Query(
		`SELECT project_key, COALESCE(label,''), COALESCE(sort_order,0), hidden FROM project_groups`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Group
	for rows.Next() {
		var g Group
		var hidden int
		if err := rows.Scan(&g.Key, &g.Label, &g.Order, &hidden); err != nil {
			return nil, err
		}
		g.Hidden = hidden != 0
		out = append(out, g)
	}
	return out, rows.Err()
}

// SaveGroups upserts the given project-group configs in one transaction.
func (s *Store) SaveGroups(groups []Group) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, g := range groups {
		hidden := 0
		if g.Hidden {
			hidden = 1
		}
		if _, err := tx.Exec(
			`INSERT INTO project_groups(project_key,label,sort_order,hidden) VALUES(?,?,?,?)
			 ON CONFLICT(project_key) DO UPDATE SET label=excluded.label, sort_order=excluded.sort_order, hidden=excluded.hidden`,
			g.Key, g.Label, g.Order, hidden); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// --- one-time migration from the old file-based layout ---

// MigrateFromFiles imports legacy state into the DB, idempotently:
//   - rootsPath (roots.json) → config "roots", only if unset in the DB
//   - every <id>.md in notesDir → notes table, unless that session already
//     has a note in the DB
func (s *Store) MigrateFromFiles(rootsPath, notesDir string) error {
	if _, ok, _ := s.GetConfig("roots"); !ok && rootsPath != "" {
		if data, err := os.ReadFile(rootsPath); err == nil {
			var cfg struct {
				Roots []string `json:"roots"`
			}
			if json.Unmarshal(data, &cfg) == nil && len(cfg.Roots) > 0 {
				if err := s.SetRoots(cfg.Roots); err != nil {
					return err
				}
			}
		}
	}
	if notesDir != "" {
		entries, _ := os.ReadDir(notesDir)
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
				continue
			}
			id := strings.TrimSuffix(e.Name(), ".md")
			if existing, _ := s.GetNote(id); existing != "" {
				continue
			}
			if data, err := os.ReadFile(filepath.Join(notesDir, e.Name())); err == nil && len(data) > 0 {
				if err := s.SetNote(id, string(data)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
