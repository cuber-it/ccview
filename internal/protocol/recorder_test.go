package protocol

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cuber-it/ccview/internal/export"
)

func userLine(text string) string {
	return `{"type":"user","message":{"role":"user","content":"` + text + `"}}`
}
func asstLine(text string) string {
	return `{"type":"assistant","message":{"role":"assistant","content":"` + text + `"}}`
}

func writeLines(t *testing.T, path string, lines ...string) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	for _, l := range lines {
		if _, err := f.WriteString(l + "\n"); err != nil {
			t.Fatal(err)
		}
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for: %s", what)
}

func markerN(path string) int {
	n, _, ok := readLastMarker(path)
	if !ok {
		return -1
	}
	return n
}

func TestRecorderFreshThenResume(t *testing.T) {
	dir := t.TempDir()
	jsonl := filepath.Join(dir, "s.jsonl")
	writeLines(t, jsonl, userLine("Frage eins"), asstLine("Antwort eins"))

	m := NewManager(context.Background(), filepath.Join(dir, "protocols"), nil)
	htmlPath := m.HTMLPath("s")

	if err := m.Start("s", jsonl, export.Meta{SessionID: "s"}); err != nil {
		t.Fatal(err)
	}
	waitFor(t, "fresh build marker n>=2", func() bool { return markerN(htmlPath) >= 2 })
	m.Stop("s")

	h1 := readFile(t, htmlPath)
	if strings.Count(h1, "<!doctype html>") != 1 {
		t.Fatalf("head should appear exactly once:\n%s", h1)
	}
	if !strings.Contains(h1, "Frage eins") || !strings.Contains(h1, "Antwort eins") {
		t.Fatal("history events missing")
	}
	if !strings.Contains(h1, "Aufnahme pausiert") {
		t.Fatal("pause boundary missing on stop")
	}

	// An event arrives while recording is off; resume must append, not rebuild.
	writeLines(t, jsonl, userLine("Frage zwei"))
	if err := m.Start("s", jsonl, export.Meta{SessionID: "s"}); err != nil {
		t.Fatal(err)
	}
	waitFor(t, "resume marker n>=3", func() bool { return markerN(htmlPath) >= 3 })
	m.Stop("s")

	h2 := readFile(t, htmlPath)
	if strings.Count(h2, "<!doctype html>") != 1 {
		t.Fatalf("resume must not rebuild the head:\n%s", h2)
	}
	if !strings.Contains(h2, "Aufnahme fortgesetzt") {
		t.Fatal("resume boundary missing")
	}
	if !strings.Contains(h2, "Frage zwei") {
		t.Fatal("new event not appended on resume")
	}
	if !strings.Contains(h2, "#0002") {
		t.Fatalf("prompt numbering must continue across resume:\n%s", h2)
	}
}

func TestRecorderRebuildsWhenMarkerStale(t *testing.T) {
	dir := t.TempDir()
	jsonl := filepath.Join(dir, "s.jsonl")
	writeLines(t, jsonl, userLine("nur eins"))

	protoDir := filepath.Join(dir, "protocols")
	if err := os.MkdirAll(protoDir, 0o755); err != nil {
		t.Fatal(err)
	}
	htmlPath := filepath.Join(protoDir, "s.html")
	// A stale transcript claiming 99 events already written; the JSONL has one.
	stale := "<!doctype html>\nOLD CONTENT\n" + markerComment(99, 40)
	if err := os.WriteFile(htmlPath, []byte(stale), 0o644); err != nil {
		t.Fatal(err)
	}

	m := NewManager(context.Background(), protoDir, nil)
	if err := m.Start("s", jsonl, export.Meta{SessionID: "s"}); err != nil {
		t.Fatal(err)
	}
	waitFor(t, "rebuild (marker back to 1)", func() bool {
		return markerN(htmlPath) == 1 && !strings.Contains(readFile(t, htmlPath), "OLD CONTENT")
	})
	m.Stop("s")

	out := readFile(t, htmlPath)
	if strings.Contains(out, "OLD CONTENT") {
		t.Fatal("stale transcript should have been rebuilt from the JSONL")
	}
	if !strings.Contains(out, "nur eins") {
		t.Fatal("rebuilt transcript missing content")
	}
}
