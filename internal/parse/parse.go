// Package parse converts raw Claude Code JSONL lines into typed Events.
package parse

import (
	"encoding/json"
	"fmt"
	"time"
)

// Kind is the top-level event type.
type Kind string

const (
	KindUser       Kind = "user"
	KindAssistant  Kind = "assistant"
	KindSystem     Kind = "system"
	KindPermission Kind = "permission-mode"
	KindAttachment Kind = "attachment"
	KindUnknown    Kind = "unknown"
)

// BlockKind is the type of a content block inside an Event.
type BlockKind string

const (
	BlockText       BlockKind = "text"
	BlockThinking   BlockKind = "thinking"
	BlockToolUse    BlockKind = "tool_use"
	BlockToolResult BlockKind = "tool_result"
	BlockUserPrompt BlockKind = "user_prompt"
)

// Event is a single parsed JSONL line.
type Event struct {
	Kind       Kind            `json:"kind"`
	Timestamp  time.Time       `json:"timestamp,omitempty"`
	UUID       string          `json:"uuid,omitempty"`
	ParentUUID string          `json:"parent_uuid,omitempty"`
	SessionID  string          `json:"session_id,omitempty"`
	Blocks     []Block         `json:"blocks,omitempty"`
	Raw        json.RawMessage `json:"-"`
}

// Block is one content block inside an Event.
type Block struct {
	Kind      BlockKind       `json:"kind"`
	Text      string          `json:"text,omitempty"`
	ToolName  string          `json:"tool_name,omitempty"`
	ToolID    string          `json:"tool_id,omitempty"`
	ToolInput json.RawMessage `json:"tool_input,omitempty"`
	ToolUseID string          `json:"tool_use_id,omitempty"`
	IsError   bool            `json:"is_error,omitempty"`
}

type rawLine struct {
	Type       string      `json:"type"`
	Timestamp  string      `json:"timestamp"`
	UUID       string      `json:"uuid"`
	ParentUUID string      `json:"parentUuid"`
	SessionID  string      `json:"sessionId"`
	Message    *rawMessage `json:"message"`
}

type rawMessage struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

type rawBlock struct {
	Type      string          `json:"type"`
	Text      string          `json:"text"`
	Thinking  string          `json:"thinking"`
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Input     json.RawMessage `json:"input"`
	ToolUseID string          `json:"tool_use_id"`
	Content   json.RawMessage `json:"content"`
	IsError   bool            `json:"is_error"`
}

// Parse converts a single JSONL line into an Event.
// Returns an error only if the line is not valid JSON.
// Unknown event types yield Kind=KindUnknown (forward-compatible).
func Parse(line []byte) (Event, error) {
	var raw rawLine
	if err := json.Unmarshal(line, &raw); err != nil {
		return Event{}, fmt.Errorf("parse jsonl line: %w", err)
	}
	ev := Event{
		UUID:       raw.UUID,
		ParentUUID: raw.ParentUUID,
		SessionID:  raw.SessionID,
		Kind:       kindOf(raw.Type),
		Raw:        append(json.RawMessage{}, line...),
	}
	if raw.Timestamp != "" {
		if ts, err := time.Parse(time.RFC3339Nano, raw.Timestamp); err == nil {
			ev.Timestamp = ts
		}
	}
	if raw.Message != nil {
		ev.Blocks = extractBlocks(ev.Kind, raw.Message)
	}
	return ev, nil
}

func kindOf(t string) Kind {
	switch t {
	case "user":
		return KindUser
	case "assistant":
		return KindAssistant
	case "system":
		return KindSystem
	case "permission-mode":
		return KindPermission
	case "attachment":
		return KindAttachment
	default:
		return KindUnknown
	}
}

func extractBlocks(k Kind, m *rawMessage) []Block {
	if len(m.Content) == 0 {
		return nil
	}
	if m.Content[0] == '"' {
		var s string
		if err := json.Unmarshal(m.Content, &s); err == nil {
			if k == KindUser {
				return []Block{{Kind: BlockUserPrompt, Text: s}}
			}
			return []Block{{Kind: BlockText, Text: s}}
		}
	}
	var raws []rawBlock
	if err := json.Unmarshal(m.Content, &raws); err != nil {
		return nil
	}
	out := make([]Block, 0, len(raws))
	for _, rb := range raws {
		if b := toBlock(rb); b.Kind != "" {
			out = append(out, b)
		}
	}
	return out
}

func toBlock(rb rawBlock) Block {
	switch rb.Type {
	case "text":
		return Block{Kind: BlockText, Text: rb.Text}
	case "thinking":
		return Block{Kind: BlockThinking, Text: rb.Thinking}
	case "tool_use":
		return Block{
			Kind:      BlockToolUse,
			ToolName:  rb.Name,
			ToolID:    rb.ID,
			ToolInput: rb.Input,
		}
	case "tool_result":
		return Block{
			Kind:      BlockToolResult,
			Text:      flattenResultContent(rb.Content),
			ToolUseID: rb.ToolUseID,
			IsError:   rb.IsError,
		}
	}
	return Block{}
}

// flattenResultContent coalesces a tool_result.content field into plain text.
// The field can be a JSON string OR an array of text-blocks.
func flattenResultContent(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	if raw[0] == '"' {
		var s string
		_ = json.Unmarshal(raw, &s)
		return s
	}
	var sub []rawBlock
	if err := json.Unmarshal(raw, &sub); err != nil {
		return ""
	}
	var out string
	for _, s := range sub {
		if s.Type == "text" {
			out += s.Text
		}
	}
	return out
}
