package tail

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTailer_History(t *testing.T) {
	path := writeTemp(t, "line1\nline2\nline3\n")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	ch := New(path).WithInterval(10 * time.Millisecond).Stream(ctx)
	got := collect(ch, 3, time.Second)
	assertLines(t, got, []string{"line1", "line2", "line3"})
}

func TestTailer_Live(t *testing.T) {
	path := writeTemp(t, "initial\n")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	ch := New(path).WithInterval(10 * time.Millisecond).Stream(ctx)

	select {
	case l := <-ch:
		if l.Err != nil || string(l.Data) != "initial" {
			t.Fatalf("initial: %+v", l)
		}
	case <-time.After(time.Second):
		t.Fatal("no initial line")
	}

	appendTo(t, path, "live1\nlive2\n")
	got := collect(ch, 2, time.Second)
	assertLines(t, got, []string{"live1", "live2"})
}

func TestTailer_PartialLineCompleted(t *testing.T) {
	path := writeTemp(t, "part")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	ch := New(path).WithInterval(10 * time.Millisecond).Stream(ctx)

	select {
	case l := <-ch:
		t.Fatalf("unexpected line before newline: %+v", l)
	case <-time.After(50 * time.Millisecond):
	}

	appendTo(t, path, "ial\n")

	select {
	case l := <-ch:
		if string(l.Data) != "partial" {
			t.Errorf("got %q, want %q", l.Data, "partial")
		}
	case <-time.After(time.Second):
		t.Fatal("no line after completion")
	}
}

func TestTailer_MissingFile(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ch := New(filepath.Join(t.TempDir(), "does-not-exist.jsonl")).Stream(ctx)
	select {
	case l, ok := <-ch:
		if !ok {
			t.Fatal("channel closed without error")
		}
		if l.Err == nil {
			t.Error("expected error in Line")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout")
	}
	if _, ok := <-ch; ok {
		t.Error("channel should be closed after fatal error")
	}
}

func TestTailer_CancelStops(t *testing.T) {
	path := writeTemp(t, "x\n")
	ctx, cancel := context.WithCancel(context.Background())
	ch := New(path).WithInterval(10 * time.Millisecond).Stream(ctx)
	<-ch // drain initial
	cancel()
	deadline := time.After(500 * time.Millisecond)
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				return
			}
		case <-deadline:
			t.Fatal("channel not closed after cancel")
		}
	}
}

func TestTailer_LongLine(t *testing.T) {
	big := make([]byte, 200_000)
	for i := range big {
		big[i] = 'x'
	}
	path := writeTemp(t, string(big)+"\n")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	ch := New(path).WithInterval(10 * time.Millisecond).Stream(ctx)
	select {
	case l := <-ch:
		if len(l.Data) != 200_000 {
			t.Errorf("len = %d, want 200000", len(l.Data))
		}
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "t.jsonl")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func appendTo(t *testing.T, path, content string) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
}

func collect(ch <-chan Line, n int, timeout time.Duration) []Line {
	var out []Line
	deadline := time.After(timeout)
	for len(out) < n {
		select {
		case l, ok := <-ch:
			if !ok {
				return out
			}
			out = append(out, l)
		case <-deadline:
			return out
		}
	}
	return out
}

func assertLines(t *testing.T, got []Line, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d: %+v", len(got), len(want), got)
	}
	for i, w := range want {
		if got[i].Err != nil {
			t.Errorf("[%d] err: %v", i, got[i].Err)
			continue
		}
		if string(got[i].Data) != w {
			t.Errorf("[%d] = %q, want %q", i, got[i].Data, w)
		}
	}
}
