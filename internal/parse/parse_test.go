package parse

import (
	"encoding/json"
	"testing"
	"time"
)

func TestParse_UserPrompt(t *testing.T) {
	line := []byte(`{"type":"user","uuid":"u1","timestamp":"2026-04-24T10:00:00Z","message":{"role":"user","content":"hello world"}}`)
	ev, err := Parse(line)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if ev.Kind != KindUser {
		t.Errorf("Kind = %q, want %q", ev.Kind, KindUser)
	}
	if len(ev.Blocks) != 1 {
		t.Fatalf("len(Blocks) = %d, want 1", len(ev.Blocks))
	}
	if ev.Blocks[0].Kind != BlockUserPrompt {
		t.Errorf("Block.Kind = %q, want %q", ev.Blocks[0].Kind, BlockUserPrompt)
	}
	if ev.Blocks[0].Text != "hello world" {
		t.Errorf("Block.Text = %q", ev.Blocks[0].Text)
	}
	if ev.UUID != "u1" {
		t.Errorf("UUID = %q", ev.UUID)
	}
}

func TestParse_UserToolResult(t *testing.T) {
	line := []byte(`{"type":"user","uuid":"u2","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"t1","content":"output text","is_error":false}]}}`)
	ev, err := Parse(line)
	if err != nil {
		t.Fatal(err)
	}
	if len(ev.Blocks) != 1 {
		t.Fatalf("len(Blocks) = %d", len(ev.Blocks))
	}
	b := ev.Blocks[0]
	if b.Kind != BlockToolResult {
		t.Errorf("Block.Kind = %q, want %q", b.Kind, BlockToolResult)
	}
	if b.ToolUseID != "t1" {
		t.Errorf("ToolUseID = %q", b.ToolUseID)
	}
	if b.Text != "output text" {
		t.Errorf("Text = %q", b.Text)
	}
	if b.IsError {
		t.Error("IsError = true, want false")
	}
}

func TestParse_UserToolResult_ArrayContent(t *testing.T) {
	line := []byte(`{"type":"user","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"t1","content":[{"type":"text","text":"part1"},{"type":"text","text":"-part2"}]}]}}`)
	ev, err := Parse(line)
	if err != nil {
		t.Fatal(err)
	}
	if ev.Blocks[0].Text != "part1-part2" {
		t.Errorf("Text = %q, want %q", ev.Blocks[0].Text, "part1-part2")
	}
}

func TestParse_AssistantMixed(t *testing.T) {
	line := []byte(`{"type":"assistant","uuid":"a1","message":{"role":"assistant","content":[{"type":"thinking","thinking":"hmm"},{"type":"text","text":"hi"},{"type":"tool_use","id":"t1","name":"Bash","input":{"command":"ls"}}]}}`)
	ev, err := Parse(line)
	if err != nil {
		t.Fatal(err)
	}
	if ev.Kind != KindAssistant {
		t.Errorf("Kind = %q", ev.Kind)
	}
	if len(ev.Blocks) != 3 {
		t.Fatalf("len(Blocks) = %d, want 3", len(ev.Blocks))
	}
	want := []BlockKind{BlockThinking, BlockText, BlockToolUse}
	for i, w := range want {
		if ev.Blocks[i].Kind != w {
			t.Errorf("Blocks[%d].Kind = %q, want %q", i, ev.Blocks[i].Kind, w)
		}
	}
	if ev.Blocks[0].Text != "hmm" {
		t.Errorf("thinking Text = %q", ev.Blocks[0].Text)
	}
	if ev.Blocks[1].Text != "hi" {
		t.Errorf("text Text = %q", ev.Blocks[1].Text)
	}
	tu := ev.Blocks[2]
	if tu.ToolName != "Bash" || tu.ToolID != "t1" {
		t.Errorf("tool_use Name/ID = %q/%q", tu.ToolName, tu.ToolID)
	}
	var inp struct {
		Command string `json:"command"`
	}
	if err := json.Unmarshal(tu.ToolInput, &inp); err != nil {
		t.Fatalf("unmarshal ToolInput: %v", err)
	}
	if inp.Command != "ls" {
		t.Errorf("ToolInput.command = %q", inp.Command)
	}
}

func TestParse_Timestamp(t *testing.T) {
	line := []byte(`{"type":"user","timestamp":"2026-04-24T10:13:07.368Z","message":{"role":"user","content":"x"}}`)
	ev, err := Parse(line)
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2026, 4, 24, 10, 13, 7, 368_000_000, time.UTC)
	if !ev.Timestamp.Equal(want) {
		t.Errorf("Timestamp = %v, want %v", ev.Timestamp, want)
	}
}

func TestParse_SkipsUnknownBlockType(t *testing.T) {
	line := []byte(`{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"keep"},{"type":"future-block-kind","foo":"bar"}]}}`)
	ev, err := Parse(line)
	if err != nil {
		t.Fatal(err)
	}
	if len(ev.Blocks) != 1 || ev.Blocks[0].Text != "keep" {
		t.Errorf("Blocks = %+v", ev.Blocks)
	}
}

func TestParse_UnknownEventType(t *testing.T) {
	ev, err := Parse([]byte(`{"type":"something-new","uuid":"x"}`))
	if err != nil {
		t.Fatal(err)
	}
	if ev.Kind != KindUnknown {
		t.Errorf("Kind = %q, want %q", ev.Kind, KindUnknown)
	}
}

func TestParse_PermissionMode(t *testing.T) {
	ev, err := Parse([]byte(`{"type":"permission-mode","permissionMode":"default","sessionId":"s1"}`))
	if err != nil {
		t.Fatal(err)
	}
	if ev.Kind != KindPermission {
		t.Errorf("Kind = %q", ev.Kind)
	}
	if ev.SessionID != "s1" {
		t.Errorf("SessionID = %q", ev.SessionID)
	}
	if len(ev.Blocks) != 0 {
		t.Errorf("expected no blocks, got %d", len(ev.Blocks))
	}
}

func TestParse_InvalidJSON(t *testing.T) {
	_, err := Parse([]byte(`{not-json`))
	if err == nil {
		t.Error("expected error")
	}
}

func TestParse_ToolResultError(t *testing.T) {
	line := []byte(`{"type":"user","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"t1","content":"boom","is_error":true}]}}`)
	ev, err := Parse(line)
	if err != nil {
		t.Fatal(err)
	}
	if !ev.Blocks[0].IsError {
		t.Error("IsError = false, want true")
	}
}
