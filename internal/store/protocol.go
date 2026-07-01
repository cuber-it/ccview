package store

import "database/sql"

// ProtocolState is a session's tail-html recording state and resume marker.
// EventsWritten and PromptNum let the recorder append to an existing transcript
// after an off/on cycle instead of rebuilding it: skip that many events of the
// JSONL, and continue the #NNNN numbering from PromptNum. The HTML file carries
// the same marker as a self-describing backup; this DB row is authoritative.
type ProtocolState struct {
	Active        bool `json:"active"`
	EventsWritten int  `json:"events_written"`
	PromptNum     int  `json:"prompt_num"`
}

// GetProtocol returns a session's recording state, or the zero value (inactive,
// no marker) if the session has never been recorded.
func (s *Store) GetProtocol(sessionID string) (ProtocolState, error) {
	var ps ProtocolState
	var active int
	err := s.db.QueryRow(
		`SELECT active, events_written, prompt_num FROM session_protocol WHERE session_id=?`,
		sessionID).Scan(&active, &ps.EventsWritten, &ps.PromptNum)
	if err == sql.ErrNoRows {
		return ProtocolState{}, nil
	}
	if err != nil {
		return ProtocolState{}, err
	}
	ps.Active = active != 0
	return ps, nil
}

// SetProtocolActive flips a session's recording on or off, leaving the resume
// marker untouched so turning it back on continues where it left off.
func (s *Store) SetProtocolActive(sessionID string, active bool) error {
	a := 0
	if active {
		a = 1
	}
	_, err := s.db.Exec(
		`INSERT INTO session_protocol(session_id,active,updated_at) VALUES(?,?,?)
		 ON CONFLICT(session_id) DO UPDATE SET active=excluded.active, updated_at=excluded.updated_at`,
		sessionID, a, now())
	return err
}

// SetProtocolMarker records how far the transcript has been written: how many
// events are on disk and the running prompt number. Called by the recorder as
// it appends, so an interrupted or restarted server resumes accurately.
func (s *Store) SetProtocolMarker(sessionID string, eventsWritten, promptNum int) error {
	_, err := s.db.Exec(
		`INSERT INTO session_protocol(session_id,events_written,prompt_num,updated_at) VALUES(?,?,?,?)
		 ON CONFLICT(session_id) DO UPDATE SET events_written=excluded.events_written, prompt_num=excluded.prompt_num, updated_at=excluded.updated_at`,
		sessionID, eventsWritten, promptNum, now())
	return err
}

// ActiveProtocols returns the session IDs currently flagged for recording, so
// the server can resume their recorders after a restart.
func (s *Store) ActiveProtocols() ([]string, error) {
	rows, err := s.db.Query(`SELECT session_id FROM session_protocol WHERE active=1`)
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
