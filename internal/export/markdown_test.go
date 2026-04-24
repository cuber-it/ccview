package export

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/cuber-it/ccview/internal/parse"
)

func TestMarkdown_Basic(t *testing.T) {
	start := time.Date(2026, 4, 24, 11, 26, 23, 0, time.UTC)
	events := []parse.Event{
		{
			Kind:      parse.KindUser,
			Timestamp: start,
			Blocks:    []parse.Block{{Kind: parse.BlockUserPrompt, Text: "Hallo"}},
		},
		{
			Kind:      parse.KindAssistant,
			Timestamp: start.Add(2 * time.Second),
			Blocks: []parse.Block{
				{Kind: parse.BlockText, Text: "Hi there"},
				{Kind: parse.BlockToolUse, ToolName: "Bash", ToolInput: json.RawMessage(`{"command":"ls"}`)},
			},
		},
		{
			Kind:      parse.KindUser,
			Timestamp: start.Add(3 * time.Second),
			Blocks:    []parse.Block{{Kind: parse.BlockToolResult, Text: "file1\nfile2"}},
		},
	}
	md := Markdown(Meta{
		SessionID:   "abc12345-xxxx",
		ProjectPath: "/tmp/proj",
		Started:     start,
		Exported:    start,
	}, events)

	want := []string{
		"# Claude Code Session",
		"**ID:** abc12345-xxxx",
		"**Projekt:** /tmp/proj",
		"## #0001 User",
		"Hallo",
		"## Assistant",
		"Hi there",
		"**Tool: Bash**",
		"$ ls",
		"**Result:**",
		"file1\nfile2",
	}
	for _, w := range want {
		if !strings.Contains(md, w) {
			t.Errorf("markdown missing %q\n--- output ---\n%s", w, md)
		}
	}
}

func TestMarkdown_Thinking(t *testing.T) {
	md := Markdown(Meta{}, []parse.Event{
		{
			Kind:   parse.KindAssistant,
			Blocks: []parse.Block{{Kind: parse.BlockThinking, Text: "pondering..."}},
		},
	})
	if !strings.Contains(md, "<details><summary>thinking</summary>") {
		t.Errorf("expected thinking details block:\n%s", md)
	}
	if !strings.Contains(md, "pondering...") {
		t.Error("thinking text missing")
	}
}

func TestMarkdown_ToolResultError(t *testing.T) {
	md := Markdown(Meta{}, []parse.Event{
		{
			Kind:   parse.KindUser,
			Blocks: []parse.Block{{Kind: parse.BlockToolResult, Text: "boom", IsError: true}},
		},
	})
	if !strings.Contains(md, "**Result (error):**") {
		t.Errorf("expected error marker:\n%s", md)
	}
}

func TestMarkdown_EditDiff(t *testing.T) {
	in := json.RawMessage(`{"file_path":"/a","old_string":"foo","new_string":"bar"}`)
	ev := parse.Event{
		Kind: parse.KindAssistant,
		Blocks: []parse.Block{
			{Kind: parse.BlockToolUse, ToolName: "Edit", ToolInput: in},
		},
	}
	md := Markdown(Meta{}, []parse.Event{ev})
	for _, w := range []string{"- foo", "+ bar", "**file:** `/a`", "```diff"} {
		if !strings.Contains(md, w) {
			t.Errorf("missing %q in:\n%s", w, md)
		}
	}
}

func TestMarkdown_SkipsPermissionMode(t *testing.T) {
	md := Markdown(Meta{}, []parse.Event{
		{Kind: parse.KindPermission},
		{Kind: parse.KindUser, Blocks: []parse.Block{{Kind: parse.BlockUserPrompt, Text: "hi"}}},
	})
	if strings.Contains(md, "permission") {
		t.Errorf("permission-mode leaked:\n%s", md)
	}
	if !strings.Contains(md, "## #0001 User") {
		t.Error("prompt numbering broken")
	}
}
