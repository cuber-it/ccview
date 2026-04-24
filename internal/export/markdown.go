// Package export turns a sequence of parsed events into a markdown transcript.
package export

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cuber-it/ccview/internal/parse"
)

// Meta captures the context written into the document header.
type Meta struct {
	SessionID   string
	ProjectPath string
	Started     time.Time
	Exported    time.Time
}

// Markdown renders events as a Markdown transcript.
func Markdown(meta Meta, events []parse.Event) string {
	var b strings.Builder
	writeHeader(&b, meta, len(events))

	promptNum := 0
	for _, ev := range events {
		switch ev.Kind {
		case parse.KindUser:
			if firstKind(ev) == parse.BlockUserPrompt {
				promptNum++
				writeUserPrompt(&b, ev, promptNum)
			} else {
				writeToolResults(&b, ev)
			}
		case parse.KindAssistant:
			writeAssistant(&b, ev)
		}
	}
	return b.String()
}

func writeHeader(b *strings.Builder, m Meta, n int) {
	fmt.Fprintf(b, "# Claude Code Session\n\n")
	fmt.Fprintf(b, "- **ID:** %s\n", m.SessionID)
	if m.ProjectPath != "" {
		fmt.Fprintf(b, "- **Projekt:** %s\n", m.ProjectPath)
	}
	if !m.Started.IsZero() {
		fmt.Fprintf(b, "- **Start:** %s\n", m.Started.Local().Format("2006-01-02 15:04:05"))
	}
	if !m.Exported.IsZero() {
		fmt.Fprintf(b, "- **Export:** %s\n", m.Exported.Local().Format("2006-01-02 15:04:05"))
	}
	fmt.Fprintf(b, "- **Events:** %d\n\n---\n\n", n)
}

func writeUserPrompt(b *strings.Builder, ev parse.Event, num int) {
	fmt.Fprintf(b, "## #%04d User %s\n\n", num, ts(ev.Timestamp))
	for _, blk := range ev.Blocks {
		if blk.Kind == parse.BlockUserPrompt && strings.TrimSpace(blk.Text) != "" {
			b.WriteString(strings.TrimSpace(blk.Text))
			b.WriteString("\n\n")
		}
	}
}

func writeAssistant(b *strings.Builder, ev parse.Event) {
	fmt.Fprintf(b, "## Assistant %s\n\n", ts(ev.Timestamp))
	for _, blk := range ev.Blocks {
		switch blk.Kind {
		case parse.BlockText:
			if strings.TrimSpace(blk.Text) == "" {
				continue
			}
			b.WriteString(strings.TrimSpace(blk.Text))
			b.WriteString("\n\n")
		case parse.BlockThinking:
			if strings.TrimSpace(blk.Text) == "" {
				continue
			}
			b.WriteString("<details><summary>thinking</summary>\n\n")
			b.WriteString(strings.TrimSpace(blk.Text))
			b.WriteString("\n\n</details>\n\n")
		case parse.BlockToolUse:
			writeToolUse(b, blk)
		case parse.BlockImage:
			writeImage(b, blk)
		}
	}
}

func writeToolUse(b *strings.Builder, blk parse.Block) {
	fmt.Fprintf(b, "**Tool: %s**\n\n", blk.ToolName)
	pretty := prettyInput(blk.ToolName, blk.ToolInput)
	if pretty != "" {
		b.WriteString(pretty)
		b.WriteString("\n\n")
	}
}

func writeToolResults(b *strings.Builder, ev parse.Event) {
	for _, blk := range ev.Blocks {
		if blk.Kind != parse.BlockToolResult {
			continue
		}
		if blk.IsError {
			b.WriteString("**Result (error):**\n\n")
		} else {
			b.WriteString("**Result:**\n\n")
		}
		text := strings.TrimRight(blk.Text, "\n")
		b.WriteString("```\n")
		b.WriteString(text)
		b.WriteString("\n```\n\n")
	}
}

func writeImage(b *strings.Builder, blk parse.Block) {
	if blk.ImageSource == "url" && blk.ImageData != "" {
		fmt.Fprintf(b, "![image](%s)\n\n", blk.ImageData)
		return
	}
	if blk.ImageData != "" {
		mt := blk.ImageMediaType
		if mt == "" {
			mt = "image/png"
		}
		fmt.Fprintf(b, "![image](data:%s;base64,%s)\n\n", mt, blk.ImageData)
	}
}

// prettyInput renders tool-input JSON in a readable way for the common tools.
func prettyInput(tool string, input json.RawMessage) string {
	if len(input) == 0 {
		return ""
	}
	switch tool {
	case "Bash":
		var v struct {
			Command string `json:"command"`
		}
		if json.Unmarshal(input, &v) == nil && v.Command != "" {
			return "```bash\n$ " + v.Command + "\n```"
		}
	case "Read":
		var v struct {
			FilePath string `json:"file_path"`
			Offset   int    `json:"offset"`
			Limit    int    `json:"limit"`
		}
		if json.Unmarshal(input, &v) == nil && v.FilePath != "" {
			extra := ""
			if v.Offset != 0 {
				extra += fmt.Sprintf(" @%d", v.Offset)
			}
			if v.Limit != 0 {
				extra += fmt.Sprintf(" limit=%d", v.Limit)
			}
			return "`" + v.FilePath + extra + "`"
		}
	case "Edit":
		var v struct {
			FilePath   string `json:"file_path"`
			OldString  string `json:"old_string"`
			NewString  string `json:"new_string"`
			ReplaceAll bool   `json:"replace_all"`
		}
		if json.Unmarshal(input, &v) == nil {
			var sb strings.Builder
			fmt.Fprintf(&sb, "**file:** `%s`", v.FilePath)
			if v.ReplaceAll {
				sb.WriteString("  _(replace_all)_")
			}
			sb.WriteString("\n\n```diff\n")
			for _, line := range strings.Split(v.OldString, "\n") {
				sb.WriteString("- ")
				sb.WriteString(line)
				sb.WriteString("\n")
			}
			for _, line := range strings.Split(v.NewString, "\n") {
				sb.WriteString("+ ")
				sb.WriteString(line)
				sb.WriteString("\n")
			}
			sb.WriteString("```")
			return sb.String()
		}
	case "Write":
		var v struct {
			FilePath string `json:"file_path"`
			Content  string `json:"content"`
		}
		if json.Unmarshal(input, &v) == nil {
			return fmt.Sprintf("**file:** `%s`\n\n```\n%s\n```", v.FilePath, v.Content)
		}
	}
	// fallback: pretty JSON
	var pretty json.RawMessage
	if err := json.Unmarshal(input, &pretty); err == nil {
		formatted, err := json.MarshalIndent(pretty, "", "  ")
		if err == nil {
			return "```json\n" + string(formatted) + "\n```"
		}
	}
	return "```\n" + string(input) + "\n```"
}

func firstKind(ev parse.Event) parse.BlockKind {
	for _, b := range ev.Blocks {
		return b.Kind
	}
	return ""
}

func ts(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return "· " + t.Local().Format("15:04:05")
}
