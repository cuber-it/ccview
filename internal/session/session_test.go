package session

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestProjectDirFromCwd(t *testing.T) {
	got := ProjectDirFromCwd("/home/ucuber/Workspace/cuber-it/sps-sim/go")
	want := "-home-ucuber-Workspace-cuber-it-sps-sim-go"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestList(t *testing.T) {
	root, proj := setupSessions(t, map[string]time.Duration{
		"78ff0cff-aaaa-bbbb-cccc-111111111111.jsonl": 0,
		"704435b8-aaaa-bbbb-cccc-222222222222.jsonl": -10 * time.Minute,
		"abcdef12-aaaa-bbbb-cccc-333333333333.jsonl": -5 * time.Minute,
		"readme.md":                                  0,
	})

	infos, err := List(root, proj)
	if err != nil {
		t.Fatal(err)
	}
	if len(infos) != 3 {
		t.Fatalf("len = %d, want 3: %+v", len(infos), infos)
	}
	// newest first
	if !strings.HasPrefix(infos[0].ID, "78ff0cff") {
		t.Errorf("first = %s, want 78ff0cff...", infos[0].ID)
	}
	if !strings.HasPrefix(infos[2].ID, "704435b8") {
		t.Errorf("last = %s, want 704435b8...", infos[2].ID)
	}
}

func TestResolve_Latest(t *testing.T) {
	ss := []Info{{ID: "abc", ModTime: time.Now()}, {ID: "def", ModTime: time.Now().Add(-time.Hour)}}
	got, err := Resolve(ss, "latest")
	if err != nil || got.ID != "abc" {
		t.Errorf("got %q err=%v", got.ID, err)
	}
}

func TestResolve_Exact(t *testing.T) {
	ss := []Info{{ID: "78ff0cff-c9ad-4344-bbea-f256af780b94"}}
	got, err := Resolve(ss, "78ff0cff-c9ad-4344-bbea-f256af780b94")
	if err != nil || got.ID != ss[0].ID {
		t.Errorf("got %q err=%v", got.ID, err)
	}
}

func TestResolve_Prefix(t *testing.T) {
	ss := []Info{{ID: "78ff0cff-xxx"}, {ID: "704435b8-yyy"}}
	got, err := Resolve(ss, "78ff")
	if err != nil || got.ID != "78ff0cff-xxx" {
		t.Errorf("got %q err=%v", got.ID, err)
	}
}

func TestResolve_PrefixTooShort(t *testing.T) {
	ss := []Info{{ID: "78ff0cff-xxx"}}
	_, err := Resolve(ss, "78f")
	if err == nil {
		t.Error("expected error for short prefix")
	}
}

func TestResolve_Ambiguous(t *testing.T) {
	ss := []Info{{ID: "78ff0cff-aaa"}, {ID: "78ff1234-bbb"}}
	_, err := Resolve(ss, "78ff")
	if err == nil || !strings.Contains(err.Error(), "ambiguous") {
		t.Errorf("err=%v", err)
	}
}

func TestResolve_NoMatch(t *testing.T) {
	ss := []Info{{ID: "78ff0cff-xxx"}}
	_, err := Resolve(ss, "deadbeef")
	if err == nil || !strings.Contains(err.Error(), "no session matching") {
		t.Errorf("err=%v", err)
	}
}

func TestResolve_Empty(t *testing.T) {
	_, err := Resolve(nil, "latest")
	if err == nil {
		t.Error("expected error for empty sessions")
	}
}

// setupSessions creates a fake projects dir with the given files.
// Each value is a mtime-offset relative to now.
func setupSessions(t *testing.T, files map[string]time.Duration) (root, project string) {
	t.Helper()
	root = t.TempDir()
	project = "-tmp-test-project"
	dir := filepath.Join(root, project)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	for name, offset := range files {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		mt := now.Add(offset)
		if err := os.Chtimes(path, mt, mt); err != nil {
			t.Fatal(err)
		}
	}
	return root, project
}
