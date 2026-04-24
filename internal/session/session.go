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
	ID            string    // full session UUID (filename without .jsonl)
	Path          string    // absolute path to the .jsonl
	Project       string    // project dir name (with leading dash)
	Size          int64
	ModTime       time.Time // file mtime — unreliable as "last activity" (Claude Code touches files across sessions)
	LastEventTime time.Time // timestamp of the last event inside the JSONL; zero if none found
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
// by LastEventTime (from the JSONL content itself) with ModTime as fallback.
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
		last, _ := ReadLastEventTime(path) // zero on error — sort falls back to mtime
		out = append(out, Info{
			ID:            strings.TrimSuffix(e.Name(), ".jsonl"),
			Path:          path,
			Project:       projectDir,
			Size:          fi.Size(),
			ModTime:       fi.ModTime(),
			LastEventTime: last,
		})
	}
	sort.Slice(out, func(i, j int) bool { return sortKey(out[i]).After(sortKey(out[j])) })
	return out, nil
}

// sortKey picks LastEventTime if known, else ModTime.
func sortKey(i Info) time.Time {
	if !i.LastEventTime.IsZero() {
		return i.LastEventTime
	}
	return i.ModTime
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
