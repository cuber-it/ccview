// Package tail reads a file from its start and keeps yielding new lines
// as they are appended (tail -f behaviour), using simple polling.
package tail

import (
	"bytes"
	"context"
	"io"
	"os"
	"time"
)

// Line is one extracted line from the tailed file. Data has any trailing
// \n/\r stripped. Err, if set, is terminal — the Stream channel closes
// after it.
type Line struct {
	Data []byte
	Err  error
}

// Tailer follows a file and streams its lines.
type Tailer struct {
	path     string
	interval time.Duration
}

// New returns a Tailer with the default poll interval of 100ms.
func New(path string) *Tailer {
	return &Tailer{path: path, interval: 100 * time.Millisecond}
}

// WithInterval returns a copy of t using the given poll interval.
func (t *Tailer) WithInterval(d time.Duration) *Tailer {
	c := *t
	c.interval = d
	return &c
}

// Stream opens the file and starts yielding lines on the returned channel.
// Existing content is delivered first (history), then the channel stays open
// and yields newly appended lines as they arrive. The channel closes when
// ctx is canceled or a fatal error occurs.
func (t *Tailer) Stream(ctx context.Context) <-chan Line {
	out := make(chan Line, 64)
	go t.run(ctx, out)
	return out
}

func (t *Tailer) run(ctx context.Context, out chan<- Line) {
	defer close(out)
	f, err := os.Open(t.path)
	if err != nil {
		send(ctx, out, Line{Err: err})
		return
	}
	defer f.Close()

	readbuf := make([]byte, 32*1024)
	var carry []byte // partial line held across reads

	for {
		n, err := f.Read(readbuf)
		if n > 0 {
			carry = append(carry, readbuf[:n]...)
			for {
				nl := bytes.IndexByte(carry, '\n')
				if nl < 0 {
					break
				}
				line := carry[:nl]
				if len(line) > 0 && line[len(line)-1] == '\r' {
					line = line[:len(line)-1]
				}
				cp := make([]byte, len(line))
				copy(cp, line)
				if !send(ctx, out, Line{Data: cp}) {
					return
				}
				carry = carry[nl+1:]
			}
			// compact if we're wasting a lot of capacity
			if cap(carry) > 64*1024 && cap(carry) > 4*len(carry) {
				carry = append(make([]byte, 0, len(carry)), carry...)
			}
		}
		if err == nil {
			continue
		}
		if err != io.EOF {
			send(ctx, out, Line{Err: err})
			return
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(t.interval):
		}
	}
}

func send(ctx context.Context, out chan<- Line, l Line) bool {
	select {
	case out <- l:
		return true
	case <-ctx.Done():
		return false
	}
}
