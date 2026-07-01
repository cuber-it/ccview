// Package protocol maintains a continuously growing HTML transcript ("tail.html")
// for a session that can be toggled on and off. It tails the session's JSONL
// directly — independent of the live viewer's hub — so recording keeps going
// even when the viewer switches to another session.
//
// The transcript is a derived cache: it is always rebuildable from the JSONL,
// which stays the single source of truth. Every appended event is followed by a
// machine-readable resume marker (an HTML comment); the file's last marker is
// therefore self-consistent even after a crash, and is what a re-enabled or
// restarted recorder reads to continue appending instead of rebuilding.
package protocol

import (
	"bufio"
	"context"
	"fmt"
	"html"
	"io"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/cuber-it/ccview/internal/export"
	"github.com/cuber-it/ccview/internal/parse"
	"github.com/cuber-it/ccview/internal/tail"
)

// markerRe matches the resume marker appended after each event. n is the number
// of parsed events already on disk; prompt is the running #NNNN counter.
var markerRe = regexp.MustCompile(`<!-- ccview:marker n=(\d+) prompt=(\d+) -->`)

func markerComment(n, prompt int) string {
	return fmt.Sprintf("<!-- ccview:marker n=%d prompt=%d -->\n", n, prompt)
}

// boundary renders a visible divider marking an on/off window edge, so a reader
// can see where a recording was resumed or paused. Self-contained inline style.
func boundary(text string) string {
	return fmt.Sprintf(
		"<section class=\"event\" style=\"text-align:center;background:#fafafa\">"+
			"<div class=\"label\" style=\"color:#999\">%s</div></section>\n",
		html.EscapeString(text))
}

func stamp() string { return time.Now().Format("2006-01-02 15:04") }

// readLastMarker reads the resume marker from the tail of an existing transcript.
// Returns ok=false if the file is missing or carries no marker (→ build fresh).
func readLastMarker(path string) (n, prompt int, ok bool) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, false
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return 0, 0, false
	}
	const tailBytes = 64 * 1024
	start := int64(0)
	if fi.Size() > tailBytes {
		start = fi.Size() - tailBytes
	}
	buf := make([]byte, fi.Size()-start)
	if _, err := f.ReadAt(buf, start); err != nil && err != io.EOF {
		return 0, 0, false
	}
	m := markerRe.FindAllSubmatch(buf, -1)
	if len(m) == 0 {
		return 0, 0, false
	}
	last := m[len(m)-1]
	n, _ = strconv.Atoi(string(last[1]))
	prompt, _ = strconv.Atoi(string(last[2]))
	return n, prompt, true
}

// Recorder tails one session's JSONL and appends its events to an HTML file
// until its context is canceled.
type Recorder struct {
	sessionID  string
	jsonlPath  string
	htmlPath   string
	meta       export.Meta
	saveMarker func(sessionID string, events, prompt int) // mirror marker to the DB (may be nil)

	cancel context.CancelFunc
	done   chan struct{}
}

func newRecorder(sessionID, jsonlPath, htmlPath string, meta export.Meta,
	saveMarker func(string, int, int)) *Recorder {
	return &Recorder{
		sessionID: sessionID, jsonlPath: jsonlPath, htmlPath: htmlPath,
		meta: meta, saveMarker: saveMarker, done: make(chan struct{}),
	}
}

func (r *Recorder) start(parent context.Context) {
	ctx, cancel := context.WithCancel(parent)
	r.cancel = cancel
	go r.run(ctx)
}

// stop cancels the recorder and blocks until it has flushed and closed.
func (r *Recorder) stop() {
	if r.cancel != nil {
		r.cancel()
	}
	<-r.done
}

func (r *Recorder) run(ctx context.Context) {
	defer close(r.done)

	resumeN, resumePrompt, resuming := readLastMarker(r.htmlPath)

	renderer := export.NewHTMLRenderer()
	var f *os.File
	var w *bufio.Writer
	written := 0 // parsed events accounted for on disk

	// On stop: mark the window as paused and flush. No </body></html> is written
	// while recording — the file stays open so a later resume appends cleanly;
	// the serving handler closes it on the fly.
	defer func() {
		if w != nil {
			fmt.Fprint(w, boundary("Aufnahme pausiert · "+stamp()))
			fmt.Fprint(w, markerComment(written, renderer.PromptNum()))
			w.Flush()
		}
		if f != nil {
			f.Close()
		}
	}()

	persist := func() {
		if w != nil {
			w.Flush()
		}
		if r.saveMarker != nil {
			r.saveMarker(r.sessionID, written, renderer.PromptNum())
		}
	}

	var hist []parse.Event
	live := false

	for line := range tail.New(r.jsonlPath).Stream(ctx) {
		if line.Err != nil {
			return
		}
		if line.Live {
			// Full history read: decide fresh build vs. resume now.
			live = true
			// Resume is valid only if the marker fits inside the current file;
			// a shorter JSONL (trimmed/rewritten) means the marker is stale, so
			// rebuild from scratch — the JSONL is the source of truth.
			if resuming && resumeN <= len(hist) {
				file, err := os.OpenFile(r.htmlPath, os.O_APPEND|os.O_WRONLY, 0o644)
				if err != nil {
					return
				}
				f, w = file, bufio.NewWriter(file)
				renderer.SetPromptNum(resumePrompt)
				written = resumeN
				fmt.Fprint(w, boundary("Aufnahme fortgesetzt · "+stamp()))
				for _, ev := range hist[resumeN:] {
					fmt.Fprint(w, renderer.Event(ev))
					written++
				}
			} else {
				file, err := os.Create(r.htmlPath) // fresh or rebuild (truncate)
				if err != nil {
					return
				}
				f, w = file, bufio.NewWriter(file)
				fmt.Fprint(w, renderer.Head(r.meta, len(hist)))
				for _, ev := range hist {
					fmt.Fprint(w, renderer.Event(ev))
					written++
				}
			}
			fmt.Fprint(w, markerComment(written, renderer.PromptNum()))
			persist()
			hist = nil
			continue
		}
		ev, err := parse.Parse(line.Data)
		if err != nil {
			continue
		}
		if !live {
			hist = append(hist, ev)
			continue
		}
		if w == nil {
			continue
		}
		fmt.Fprint(w, renderer.Event(ev))
		written++
		fmt.Fprint(w, markerComment(written, renderer.PromptNum()))
		persist()
	}
}
