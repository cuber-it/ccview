// Package session locates Claude Code session JSONL files on disk and
// resolves user-provided session specs (id, prefix, "latest") into paths.
package session

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Info is metadata about a single session JSONL file.
type Info struct {
	ID      string    // full session UUID (filename without .jsonl)
	Path    string    // absolute path to the .jsonl
	Project string    // project dir name (with leading dash)
	Size    int64
	ModTime time.Time
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

// List returns all session JSONLs in projectsRoot/projectDir, newest mtime first.
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
		out = append(out, Info{
			ID:      strings.TrimSuffix(e.Name(), ".jsonl"),
			Path:    filepath.Join(dir, e.Name()),
			Project: projectDir,
			Size:    fi.Size(),
			ModTime: fi.ModTime(),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ModTime.After(out[j].ModTime) })
	return out, nil
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
