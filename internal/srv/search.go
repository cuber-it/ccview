package srv

import (
	"bufio"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/cuber-it/ccview/internal/session"
)

// searchHit is one matching session in a cross-session search.
type searchHit struct {
	ID           string    `json:"id"`
	ShortID      string    `json:"short_id"`
	Project      string    `json:"project"`
	ProjectLabel string    `json:"project_label"`
	Name         string    `json:"name,omitempty"`
	Matches      int       `json:"matches"`
	Days         []string  `json:"days"`
	LastEvent    time.Time `json:"last_event,omitempty"`
	Snippet      string    `json:"snippet,omitempty"`
}

// scanFile counts regex matches in a session JSONL, collects the distinct
// activity days (from timestamps), and grabs a short snippet of the first hit.
// It reads line by line so very large files don't blow up memory.
func scanFile(path string, re *regexp.Regexp) (matches int, days []string, snippet string) {
	f, err := os.Open(path)
	if err != nil {
		return 0, nil, ""
	}
	defer f.Close()

	daySet := make(map[string]struct{})
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 8*1024*1024) // up to 8 MB per line
	for sc.Scan() {
		line := sc.Text()
		if i := strings.Index(line, `"timestamp":"`); i >= 0 && len(line) >= i+23 {
			daySet[line[i+13:i+23]] = struct{}{}
		}
		if loc := re.FindStringIndex(line); loc != nil {
			matches++
			if snippet == "" {
				snippet = snippetAround(line, loc[0], loc[1])
			}
		}
	}
	for d := range daySet {
		days = append(days, d)
	}
	sort.Strings(days)
	return matches, days, snippet
}

// snippetAround returns up to ~120 chars of context around a match.
func snippetAround(line string, start, end int) string {
	const pad = 50
	from := start - pad
	if from < 0 {
		from = 0
	}
	to := end + pad
	if to > len(line) {
		to = len(line)
	}
	s := strings.TrimSpace(line[from:to])
	s = strings.ReplaceAll(s, "\\n", " ")
	if len(s) > 160 {
		s = s[:160] + "…"
	}
	return s
}

func shortID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

// scanFileSnippets returns up to max snippets for lines that match re — used
// when searching within a single session.
func scanFileSnippets(path string, re *regexp.Regexp, max int) []string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	out := []string{}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	for sc.Scan() && len(out) < max {
		line := sc.Text()
		if loc := re.FindStringIndex(line); loc != nil {
			out = append(out, snippetAround(line, loc[0], loc[1]))
		}
	}
	return out
}

// handleSearch runs a case-insensitive regex. The scope query param selects
// the target: "" / "all" → every session JSONL, "notes" → every saved note,
// "session" (+ session=<id>) → a single session's JSONL.
// GET /api/search?q=<regex>&scope=<all|notes|session>&session=<id>
func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		writeJSON(w, []searchHit{})
		return
	}
	re, err := regexp.Compile("(?i)" + q)
	if err != nil {
		http.Error(w, "bad regex: "+err.Error(), http.StatusBadRequest)
		return
	}
	switch r.URL.Query().Get("scope") {
	case "notes":
		s.searchNotes(w, re)
	case "session":
		s.searchOneSession(w, re, r.URL.Query().Get("session"))
	default:
		s.searchAllSessions(w, re)
	}
}

func (s *Server) searchAllSessions(w http.ResponseWriter, re *regexp.Regexp) {
	meta, _ := s.store.AllMeta()
	out := make([]searchHit, 0)
	for _, si := range session.ListAllRoots(s.currentRoots()) {
		n, days, snippet := scanFile(si.Path, re)
		if n == 0 {
			continue
		}
		out = append(out, searchHit{
			ID: si.ID, ShortID: shortID(si.ID), Project: si.Project,
			ProjectLabel: projectLabel(si.Project), Name: meta[si.ID].Name,
			Matches: n, Days: days, LastEvent: si.LastEventTime, Snippet: snippet,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].LastEvent.After(out[j].LastEvent) })
	writeJSON(w, out)
}

func (s *Server) searchNotes(w http.ResponseWriter, re *regexp.Regexp) {
	notes, _ := s.store.AllNotes()
	meta, _ := s.store.AllMeta()
	byID := make(map[string]session.Info)
	for _, si := range session.ListAllRoots(s.currentRoots()) {
		byID[si.ID] = si
	}
	out := make([]searchHit, 0)
	for id, content := range notes {
		locs := re.FindAllStringIndex(content, -1)
		if len(locs) == 0 {
			continue
		}
		si := byID[id]
		out = append(out, searchHit{
			ID: id, ShortID: shortID(id), Project: si.Project,
			ProjectLabel: projectLabel(si.Project), Name: meta[id].Name,
			Matches: len(locs), LastEvent: si.LastEventTime,
			Snippet: snippetAround(content, locs[0][0], locs[0][1]),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Matches > out[j].Matches })
	writeJSON(w, out)
}

func (s *Server) searchOneSession(w http.ResponseWriter, re *regexp.Regexp, sid string) {
	info, ok := s.findSession(sid)
	if !ok {
		writeJSON(w, map[string]any{"matches": 0, "snippets": []string{}})
		return
	}
	snippets := scanFileSnippets(info.Path, re, 200)
	writeJSON(w, map[string]any{"matches": len(snippets), "snippets": snippets})
}
