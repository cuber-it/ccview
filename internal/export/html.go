package export

import (
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/cuber-it/ccview/internal/parse"
)

// HTML renders events as a self-contained HTML transcript. It uses inline CSS
// only, so it opens directly in a browser tab: the browser lays out the whole
// session natively in one pass (no JS, no incremental rendering) — fast even
// for huge sessions, and fully searchable with Ctrl-F or printable to PDF.
func HTML(meta Meta, events []parse.Event) string {
	var b strings.Builder
	b.WriteString("<!doctype html>\n<html lang=\"de\">\n<head>\n<meta charset=\"utf-8\">\n")
	b.WriteString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n")
	fmt.Fprintf(&b, "<title>ccview · %s</title>\n", html.EscapeString(shortID(meta.SessionID)))
	b.WriteString(htmlStyle)
	b.WriteString("</head>\n<body>\n")
	htmlHeader(&b, meta, len(events))

	promptNum := 0
	for _, ev := range events {
		switch ev.Kind {
		case parse.KindUser:
			if firstKind(ev) == parse.BlockUserPrompt {
				promptNum++
				htmlUserPrompt(&b, ev, promptNum)
			} else {
				htmlToolResults(&b, ev)
			}
		case parse.KindAssistant:
			htmlAssistant(&b, ev)
		}
	}
	b.WriteString("</body>\n</html>\n")
	return b.String()
}

const htmlStyle = `<style>
:root{--fg:#1e1e1e;--muted:#6a6a6a;--border:#e3e3e3;--bg:#fff;--code-bg:#f5f5f5;--user:#0b66c3;--tool:#7a3e9d}
*{box-sizing:border-box}
body{margin:0;background:var(--bg);color:var(--fg);font:15px/1.6 -apple-system,Segoe UI,Roboto,sans-serif;padding-bottom:60px}
header{max-width:900px;margin:0 auto;padding:20px 16px;border-bottom:1px solid var(--border)}
header h1{font-size:18px;margin:0 0 8px}
header .meta{color:var(--muted);font-size:12px}
.event{max-width:900px;margin:0 auto;padding:14px 16px;border-bottom:1px solid var(--border)}
.event.prompt{background:#f7fbff}
.label{font-size:11px;font-weight:600;color:var(--muted);margin:0 0 6px;text-transform:uppercase;letter-spacing:.04em}
.event.prompt .label{color:var(--user)}
.num{font-weight:700}
.text{white-space:pre-wrap;word-wrap:break-word;margin:4px 0}
pre{background:var(--code-bg);border:1px solid var(--border);border-radius:6px;padding:10px;overflow:auto;font:12px/1.5 ui-monospace,Menlo,Consolas,monospace;white-space:pre-wrap;word-wrap:break-word}
.tool{color:var(--tool);font-weight:600;font-size:13px;margin:8px 0 4px}
details{margin:6px 0}summary{cursor:pointer;color:var(--muted);font-size:13px}
img{max-width:100%;border:1px solid var(--border);border-radius:6px}
@media print{body{padding-bottom:0}.event,header{border-color:#ccc}}
</style>
`

func htmlHeader(b *strings.Builder, m Meta, n int) {
	b.WriteString("<header>\n<h1>Claude Code Session</h1>\n<div class=\"meta\">")
	fmt.Fprintf(b, "ID %s", html.EscapeString(m.SessionID))
	if m.ProjectPath != "" {
		fmt.Fprintf(b, " &middot; %s", html.EscapeString(m.ProjectPath))
	}
	if !m.Started.IsZero() {
		fmt.Fprintf(b, " &middot; Start %s", m.Started.Local().Format("2006-01-02 15:04"))
	}
	if !m.Exported.IsZero() {
		fmt.Fprintf(b, " &middot; Export %s", m.Exported.Local().Format("2006-01-02 15:04"))
	}
	fmt.Fprintf(b, " &middot; %d Events", n)
	b.WriteString("</div>\n</header>\n")
}

func htmlUserPrompt(b *strings.Builder, ev parse.Event, num int) {
	b.WriteString("<section class=\"event prompt\">\n")
	fmt.Fprintf(b, "<div class=\"label\"><span class=\"num\">#%04d</span> User %s</div>\n", num, htmlTs(ev.Timestamp))
	for _, blk := range ev.Blocks {
		if blk.Kind == parse.BlockUserPrompt && strings.TrimSpace(blk.Text) != "" {
			writeHTMLText(b, strings.TrimSpace(blk.Text))
		}
	}
	b.WriteString("</section>\n")
}

func htmlAssistant(b *strings.Builder, ev parse.Event) {
	b.WriteString("<section class=\"event assistant\">\n")
	fmt.Fprintf(b, "<div class=\"label\">Assistant %s</div>\n", htmlTs(ev.Timestamp))
	for _, blk := range ev.Blocks {
		switch blk.Kind {
		case parse.BlockText:
			if strings.TrimSpace(blk.Text) == "" {
				continue
			}
			writeHTMLText(b, strings.TrimSpace(blk.Text))
		case parse.BlockThinking:
			if strings.TrimSpace(blk.Text) == "" {
				continue
			}
			b.WriteString("<details><summary>thinking</summary>\n")
			writeHTMLText(b, strings.TrimSpace(blk.Text))
			b.WriteString("</details>\n")
		case parse.BlockToolUse:
			fmt.Fprintf(b, "<div class=\"tool\">Tool: %s</div>\n", html.EscapeString(blk.ToolName))
			writeHTMLText(b, prettyInput(blk.ToolName, blk.ToolInput))
		case parse.BlockImage:
			htmlImage(b, blk)
		}
	}
	b.WriteString("</section>\n")
}

func htmlToolResults(b *strings.Builder, ev parse.Event) {
	for _, blk := range ev.Blocks {
		if blk.Kind != parse.BlockToolResult {
			continue
		}
		b.WriteString("<section class=\"event toolresult\">\n")
		if blk.IsError {
			b.WriteString("<div class=\"label\">Result (error)</div>\n")
		} else {
			b.WriteString("<div class=\"label\">Result</div>\n")
		}
		b.WriteString("<pre>")
		b.WriteString(html.EscapeString(strings.TrimRight(blk.Text, "\n")))
		b.WriteString("</pre>\n</section>\n")
	}
}

func htmlImage(b *strings.Builder, blk parse.Block) {
	if blk.ImageSource == "url" && blk.ImageData != "" {
		fmt.Fprintf(b, "<img src=\"%s\" alt=\"image\">\n", html.EscapeString(blk.ImageData))
		return
	}
	if blk.ImageData != "" {
		mt := blk.ImageMediaType
		if mt == "" {
			mt = "image/png"
		}
		fmt.Fprintf(b, "<img src=\"data:%s;base64,%s\" alt=\"image\">\n", mt, blk.ImageData)
	}
}

// writeHTMLText emits text, turning ```fenced``` blocks into <pre><code> and
// everything else into whitespace-preserving paragraphs. Inline markdown
// (**bold**, `code`) is left as-is — readable, and avoids a markdown dependency.
func writeHTMLText(b *strings.Builder, s string) {
	parts := strings.Split(s, "```")
	for i, p := range parts {
		if i%2 == 1 {
			code := p
			if nl := strings.IndexByte(code, '\n'); nl >= 0 {
				if first := strings.TrimSpace(code[:nl]); first != "" && !strings.ContainsAny(first, " \t") {
					code = code[nl+1:] // drop a leading language token line
				}
			}
			b.WriteString("<pre><code>")
			b.WriteString(html.EscapeString(strings.TrimRight(code, "\n")))
			b.WriteString("</code></pre>\n")
		} else {
			if strings.TrimSpace(p) == "" {
				continue
			}
			b.WriteString("<div class=\"text\">")
			b.WriteString(html.EscapeString(p))
			b.WriteString("</div>\n")
		}
	}
}

func htmlTs(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return "&middot; " + t.Local().Format("15:04:05")
}

func shortID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}
