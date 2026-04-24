// Package session locates Claude Code session JSONL files on disk and
// resolves user-provided session specs (id, prefix, "latest") into paths.
package session

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Info is metadata about a single session JSONL file.
type Info struct {
	ID             string    // full session UUID (filename without .jsonl)
	Path           string    // absolute path to the .jsonl
	Project        string    // project dir name (with leading dash)
	Size           int64
	ModTime        time.Time // file mtime — unreliable as "last activity" (Claude Code touches files across sessions)
	FirstEventTime time.Time // timestamp of the first event inside the JSONL — session start, stable
	LastEventTime  time.Time // timestamp of the last event inside the JSONL; zero if none found
}

// ReadFirstEventTime returns the timestamp of the first event in the JSONL
// that carries a "timestamp" field. Scans only the first 64 KB.
// This corresponds to the session start and is stable — unlike mtime
// or last-event-time, it doesn't drift when Claude Code writes cross-session.
func ReadFirstEventTime(path string) (time.Time, error) {
	f, err := os.Open(path)
	if err != nil {
		return time.Time{}, err
	}
	defer f.Close()

	buf := make([]byte, 64*1024)
	n, err := io.ReadFull(f, buf)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return time.Time{}, err
	}
	data := buf[:n]

	for len(data) > 0 {
		idx := bytes.IndexByte(data, '\n')
		var line []byte
		if idx >= 0 {
			line = data[:idx]
			data = data[idx+1:]
		} else {
			line = data
			data = nil
		}
		if len(line) == 0 {
			continue
		}
		var tmp struct {
			Timestamp string `json:"timestamp"`
		}
		if json.Unmarshal(line, &tmp) == nil && tmp.Timestamp != "" {
			if t, err := time.Parse(time.RFC3339Nano, tmp.Timestamp); err == nil {
				return t, nil
			}
		}
	}
	return time.Time{}, errors.New("no timestamped event in first 64 KB")
}

// ReadLastEventTime scans the tail of a JSONL file for the most recent
// event that carries a "timestamp" field and returns it.
// Cheap: reads only the last 16 KB.
func ReadLastEventTime(path string) (time.Time, error) {
	f, err := os.Open(path)
	if err != nil {
		return time.Time{}, err
	}
	defer f.Close()

	size, err := f.Seek(0, io.SeekEnd)
	if err != nil {
		return time.Time{}, err
	}
	if size == 0 {
		return time.Time{}, errors.New("empty file")
	}

	const tailLen = 16 * 1024
	start := int64(0)
	if size > tailLen {
		start = size - tailLen
	}
	if _, err := f.Seek(start, io.SeekStart); err != nil {
		return time.Time{}, err
	}
	buf, err := io.ReadAll(f)
	if err != nil {
		return time.Time{}, err
	}
	buf = bytes.TrimRight(buf, "\n\r")

	// Walk lines from the end; return the first one with a usable timestamp.
	for {
		idx := bytes.LastIndexByte(buf, '\n')
		var line []byte
		if idx >= 0 {
			line = buf[idx+1:]
			buf = buf[:idx]
		} else {
			line = buf
			buf = nil
		}
		if len(line) == 0 {
			if len(buf) == 0 {
				break
			}
			continue
		}
		var tmp struct {
			Timestamp string `json:"timestamp"`
		}
		if json.Unmarshal(line, &tmp) == nil && tmp.Timestamp != "" {
			if t, err := time.Parse(time.RFC3339Nano, tmp.Timestamp); err == nil {
				return t, nil
			}
		}
		if len(buf) == 0 {
			break
		}
	}
	return time.Time{}, errors.New("no timestamped event in tail")
}

// ProjectsDir returns the default Claude Code projects directory.
func ProjectsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude", "projects"), nil
}

// ProjectDirFromCwd derives the Claude Code project-dir name from a working directory.
// Example: /home/ucuber/Workspace/x → -home-ucuber-Workspace-x
func ProjectDirFromCwd(cwd string) string {
	return strings.ReplaceAll(cwd, string(filepath.Separator), "-")
}

// List returns all session JSONLs in projectsRoot/projectDir, sorted newest first
// by FirstEventTime (session start) with ModTime as fallback.
func List(projectsRoot, projectDir string) ([]Info, error) {
	dir := filepath.Join(projectsRoot, projectDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	out := make([]Info, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".jsonl") {
			continue
		}
		fi, err := e.Info()
		if err != nil {
			continue
		}
		path := filepath.Join(dir, e.Name())
		first, _ := ReadFirstEventTime(path)
		last, _ := ReadLastEventTime(path)
		out = append(out, Info{
			ID:             strings.TrimSuffix(e.Name(), ".jsonl"),
			Path:           path,
			Project:        projectDir,
			Size:           fi.Size(),
			ModTime:        fi.ModTime(),
			FirstEventTime: first,
			LastEventTime:  last,
		})
	}
	sortInfos(out)
	return out, nil
}

// ListAll returns all session JSONLs across all projects under projectsRoot.
// Sorted newest-first by FirstEventTime (session start).
func ListAll(projectsRoot string) ([]Info, error) {
	projects, err := os.ReadDir(projectsRoot)
	if err != nil {
		return nil, err
	}
	var out []Info
	for _, p := range projects {
		if !p.IsDir() {
			continue
		}
		part, err := List(projectsRoot, p.Name())
		if err != nil {
			continue
		}
		out = append(out, part...)
	}
	sortInfos(out)
	return out, nil
}

func sortInfos(s []Info) {
	sort.Slice(s, func(i, j int) bool { return sortKey(s[i]).After(sortKey(s[j])) })
}

// sortKey prefers FirstEventTime (stable session start) over anything else.
// Falls back to LastEventTime, then ModTime.
func sortKey(i Info) time.Time {
	if !i.FirstEventTime.IsZero() {
		return i.FirstEventTime
	}
	if !i.LastEventTime.IsZero() {
		return i.LastEventTime
	}
	return i.ModTime
}

// ReadFirstUserPrompt returns the text of the first user event with string
// content inside path (i.e. a real prompt, not a tool-result). Scans only the
// first ~64 KB of the file.
func ReadFirstUserPrompt(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	buf := make([]byte, 64*1024)
	n, err := io.ReadFull(f, buf)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return "", err
	}
	if n == 0 {
		return "", errors.New("empty file")
	}
	data := buf[:n]

	for len(data) > 0 {
		idx := bytes.IndexByte(data, '\n')
		var line []byte
		if idx >= 0 {
			line = data[:idx]
			data = data[idx+1:]
		} else {
			line = data
			data = nil
		}
		if len(line) == 0 {
			continue
		}
		var tmp struct {
			Type    string `json:"type"`
			Message struct {
				Content json.RawMessage `json:"content"`
			} `json:"message"`
		}
		if json.Unmarshal(line, &tmp) != nil {
			continue
		}
		if tmp.Type != "user" || len(tmp.Message.Content) == 0 {
			continue
		}
		if tmp.Message.Content[0] == '"' {
			var s string
			if json.Unmarshal(tmp.Message.Content, &s) == nil && s != "" {
				return s, nil
			}
		}
	}
	return "", errors.New("no user prompt found in first 64 KB")
}

// Resolve turns a user-provided session spec into a single Info.
// Accepted specs:
//
//	"latest"                 → newest session by mtime
//	<full-uuid>              → exact match
//	<prefix> (min 4 chars)   → unique prefix match
//
// Returns an error for empty spec, short prefix, no match, or ambiguous prefix.
func Resolve(sessions []Info, spec string) (Info, error) {
	if len(sessions) == 0 {
		return Info{}, errors.New("no sessions found")
	}
	switch spec {
	case "":
		return Info{}, errors.New("empty session spec")
	case "latest":
		return sessions[0], nil
	}
	for _, s := range sessions {
		if s.ID == spec {
			return s, nil
		}
	}
	if len(spec) < 4 {
		return Info{}, fmt.Errorf("session prefix too short (need ≥4 chars): %q", spec)
	}
	var matches []Info
	for _, s := range sessions {
		if strings.HasPrefix(s.ID, spec) {
			matches = append(matches, s)
		}
	}
	switch len(matches) {
	case 0:
		return Info{}, fmt.Errorf("no session matching %q", spec)
	case 1:
		return matches[0], nil
	default:
		ids := make([]string, len(matches))
		for i, m := range matches {
			ids[i] = m.ID[:12] + "…"
		}
		return Info{}, fmt.Errorf("ambiguous prefix %q matches %d sessions: %s",
			spec, len(matches), strings.Join(ids, ", "))
	}
}
