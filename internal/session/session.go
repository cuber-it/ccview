// Package session locates Claude Code session JSONL files and resolves
// user-provided session specs (id, prefix, "latest") into paths.
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
	ID             string
	Path           string
	Project        string
	Size           int64
	ModTime        time.Time
	FirstEventTime time.Time // from the JSONL content — stable across cross-writes
	LastEventTime  time.Time // zero if no timestamped event found in tail
}

// ProjectsDir returns the default Claude Code projects directory.
func ProjectsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude", "projects"), nil
}

// ProjectDirFromCwd derives the Claude Code project-dir name from a cwd.
// Example: /home/u/Workspace/x → -home-u-Workspace-x.
// On Windows, drive-letter colons are stripped (best guess — Claude Code's
// exact Windows encoding is untested).
func ProjectDirFromCwd(cwd string) string {
	encoded := strings.ReplaceAll(cwd, string(filepath.Separator), "-")
	encoded = strings.ReplaceAll(encoded, ":", "")
	return encoded
}

// List returns all session JSONLs in projectsRoot/projectDir, sorted
// newest-first by FirstEventTime (session start), ModTime as fallback.
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

// ListAll scans every project under projectsRoot and returns all sessions,
// sorted newest-first by FirstEventTime.
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

func sortKey(i Info) time.Time {
	switch {
	case !i.FirstEventTime.IsZero():
		return i.FirstEventTime
	case !i.LastEventTime.IsZero():
		return i.LastEventTime
	default:
		return i.ModTime
	}
}

// Resolve turns a session spec into a single Info:
//
//	"latest"               → newest session by FirstEventTime
//	<full-uuid>            → exact match
//	<prefix> (≥ 4 chars)   → unique prefix match
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

// ReadFirstEventTime returns the timestamp of the first timestamped event.
// Scans the first 64 KB. Session-start time — stable, doesn't drift.
func ReadFirstEventTime(path string) (time.Time, error) {
	var out time.Time
	err := scanHead(path, 64*1024, func(line []byte) bool {
		if t, ok := extractTimestamp(line); ok {
			out = t
			return true
		}
		return false
	})
	if err != nil {
		return time.Time{}, err
	}
	if out.IsZero() {
		return time.Time{}, errors.New("no timestamped event in first 64 KB")
	}
	return out, nil
}

// ReadLastEventTime returns the timestamp of the most recent timestamped event.
// Scans the last 16 KB.
func ReadLastEventTime(path string) (time.Time, error) {
	var out time.Time
	err := scanTail(path, 16*1024, func(line []byte) bool {
		if t, ok := extractTimestamp(line); ok {
			out = t
			return true
		}
		return false
	})
	if err != nil {
		return time.Time{}, err
	}
	if out.IsZero() {
		return time.Time{}, errors.New("no timestamped event in last 16 KB")
	}
	return out, nil
}

// ReadFirstUserPrompt returns the text of the first user-string prompt
// (i.e. not a tool-result). Scans the first 64 KB.
func ReadFirstUserPrompt(path string) (string, error) {
	var out string
	err := scanHead(path, 64*1024, func(line []byte) bool {
		var tmp struct {
			Type    string `json:"type"`
			Message struct {
				Content json.RawMessage `json:"content"`
			} `json:"message"`
		}
		if json.Unmarshal(line, &tmp) != nil {
			return false
		}
		if tmp.Type != "user" || len(tmp.Message.Content) == 0 || tmp.Message.Content[0] != '"' {
			return false
		}
		var s string
		if json.Unmarshal(tmp.Message.Content, &s) == nil && s != "" {
			out = s
			return true
		}
		return false
	})
	if err != nil {
		return "", err
	}
	if out == "" {
		return "", errors.New("no user prompt found in first 64 KB")
	}
	return out, nil
}

// scanHead reads up to limit bytes from the start and calls fn per non-empty
// line, in file order. Stops when fn returns true.
func scanHead(path string, limit int, fn func(line []byte) bool) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	buf := make([]byte, limit)
	n, err := io.ReadFull(f, buf)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return err
	}
	data := buf[:n]
	for len(data) > 0 {
		idx := bytes.IndexByte(data, '\n')
		var line []byte
		if idx >= 0 {
			line, data = data[:idx], data[idx+1:]
		} else {
			line, data = data, nil
		}
		if len(line) > 0 && fn(line) {
			return nil
		}
	}
	return nil
}

// scanTail reads up to limit bytes from the end and calls fn per non-empty
// line, in reverse file order. Stops when fn returns true.
func scanTail(path string, limit int, fn func(line []byte) bool) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	size, err := f.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}
	if size == 0 {
		return nil
	}
	start := int64(0)
	if size > int64(limit) {
		start = size - int64(limit)
	}
	if _, err := f.Seek(start, io.SeekStart); err != nil {
		return err
	}
	buf, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	buf = bytes.TrimRight(buf, "\n\r")
	for len(buf) > 0 {
		idx := bytes.LastIndexByte(buf, '\n')
		var line []byte
		if idx >= 0 {
			line, buf = buf[idx+1:], buf[:idx]
		} else {
			line, buf = buf, nil
		}
		if len(line) > 0 && fn(line) {
			return nil
		}
	}
	return nil
}

func extractTimestamp(line []byte) (time.Time, bool) {
	var tmp struct {
		Timestamp string `json:"timestamp"`
	}
	if json.Unmarshal(line, &tmp) != nil || tmp.Timestamp == "" {
		return time.Time{}, false
	}
	t, err := time.Parse(time.RFC3339Nano, tmp.Timestamp)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}
