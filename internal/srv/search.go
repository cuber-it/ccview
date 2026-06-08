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

// handleSearch runs a (case-insensitive) regex across every session JSONL in
// the configured roots and returns matching sessions with hit count, activity
// days, and a snippet. GET /api/search?q=<regex>
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
	sessions := session.ListAllRoots(s.currentRoots())
	meta, _ := s.store.AllMeta()

	out := make([]searchHit, 0)
	for _, si := range sessions {
		n, days, snippet := scanFile(si.Path, re)
		if n == 0 {
			continue
		}
		short := si.ID
		if len(short) > 8 {
			short = short[:8]
		}
		out = append(out, searchHit{
			ID:           si.ID,
			ShortID:      short,
			Project:      si.Project,
			ProjectLabel: projectLabel(si.Project),
			Name:         meta[si.ID].Name,
			Matches:      n,
			Days:         days,
			LastEvent:    si.LastEventTime,
			Snippet:      snippet,
		})
	}
	// Most recently active first.
	sort.Slice(out, func(i, j int) bool { return out[i].LastEvent.After(out[j].LastEvent) })
	writeJSON(w, out)
}
