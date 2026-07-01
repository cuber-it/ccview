package store

import "strings"

// Tag is a session tag together with its origin, so manual tags can be told
// apart from heuristic ("auto") and model-generated ("ki") ones.
type Tag struct {
	Tag    string `json:"tag"`
	Source string `json:"source"` // manual | auto | ki
}

// TagsFor returns the tags of one session, ordered by tag name.
func (s *Store) TagsFor(sessionID string) ([]Tag, error) {
	rows, err := s.db.Query(
		`SELECT tag, source FROM session_tags WHERE session_id=? ORDER BY tag`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Tag
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.Tag, &t.Source); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// AllTags returns every session's tags keyed by session ID (for list rendering).
func (s *Store) AllTags() (map[string][]Tag, error) {
	rows, err := s.db.Query(`SELECT session_id, tag, source FROM session_tags ORDER BY tag`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[string][]Tag)
	for rows.Next() {
		var id string
		var t Tag
		if err := rows.Scan(&id, &t.Tag, &t.Source); err != nil {
			return nil, err
		}
		out[id] = append(out[id], t)
	}
	return out, rows.Err()
}

// TagCounts returns each tag with the number of sessions carrying it, for
// sidebar facets. Ordered most-used first is the caller's job; this returns a map.
func (s *Store) TagCounts() (map[string]int, error) {
	rows, err := s.db.Query(`SELECT tag, COUNT(*) FROM session_tags GROUP BY tag`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[string]int)
	for rows.Next() {
		var tag string
		var n int
		if err := rows.Scan(&tag, &n); err != nil {
			return nil, err
		}
		out[tag] = n
	}
	return out, rows.Err()
}

// SessionsByTag returns the session IDs carrying the given tag.
func (s *Store) SessionsByTag(tag string) ([]string, error) {
	rows, err := s.db.Query(`SELECT session_id FROM session_tags WHERE tag=?`, tag)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

// normalizeTag lower-cases and trims a tag; empty after that means "skip".
func normalizeTag(tag string) string {
	return strings.ToLower(strings.TrimSpace(tag))
}

// AddTag adds one tag to a session (idempotent). source is manual|auto|ki.
func (s *Store) AddTag(sessionID, tag, source string) error {
	tag = normalizeTag(tag)
	if tag == "" {
		return nil
	}
	_, err := s.db.Exec(
		`INSERT INTO session_tags(session_id,tag,source) VALUES(?,?,?)
		 ON CONFLICT(session_id,tag) DO UPDATE SET source=excluded.source`,
		sessionID, tag, source)
	return err
}

// RemoveTag removes one tag from a session.
func (s *Store) RemoveTag(sessionID, tag string) error {
	_, err := s.db.Exec(
		`DELETE FROM session_tags WHERE session_id=? AND tag=?`, sessionID, normalizeTag(tag))
	return err
}

// ReplaceTagsBySource replaces all tags of one source for a session in a single
// transaction. This is how the heuristic and Lisa re-run safely: only their own
// source's tags are cleared and rewritten, manual tags are never touched.
func (s *Store) ReplaceTagsBySource(sessionID, source string, tags []string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(
		`DELETE FROM session_tags WHERE session_id=? AND source=?`, sessionID, source); err != nil {
		return err
	}
	stmt, err := tx.Prepare(
		`INSERT INTO session_tags(session_id,tag,source) VALUES(?,?,?)
		 ON CONFLICT(session_id,tag) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	seen := map[string]bool{}
	for _, t := range tags {
		t = normalizeTag(t)
		if t == "" || seen[t] {
			continue
		}
		seen[t] = true
		if _, err := stmt.Exec(sessionID, t, source); err != nil {
			return err
		}
	}
	return tx.Commit()
}
