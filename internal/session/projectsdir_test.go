package session

import (
	"strings"
	"testing"
)

func TestProjectsDir_HonoursClaudeConfigDir(t *testing.T) {
	t.Setenv("CLAUDE_CONFIG_DIR", "/tmp/cc")
	got, err := ProjectsDir()
	if err != nil {
		t.Fatal(err)
	}
	if got != "/tmp/cc/projects" {
		t.Errorf("got %q, want /tmp/cc/projects", got)
	}
}

func TestProjectsDir_DefaultUnderHome(t *testing.T) {
	t.Setenv("CLAUDE_CONFIG_DIR", "")
	got, err := ProjectsDir()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(got, ".claude/projects") {
		t.Errorf("got %q, want suffix .claude/projects", got)
	}
}
