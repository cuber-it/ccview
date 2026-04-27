package srv

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cuber-it/ccview/internal/parse"
)

func TestServer_IndexServed(t *testing.T) {
	s := New(Config{})
	ts := httptest.NewServer(s)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		t.Fatalf("status = %d", res.StatusCode)
	}
	body, _ := io.ReadAll(res.Body)
	if !strings.Contains(string(body), "ccview") {
		t.Error("index body missing 'ccview'")
	}
	if !strings.Contains(string(body), "app.js") {
		t.Error("index body missing app.js script reference")
	}

	jsRes, err := http.Get(ts.URL + "/app.js")
	if err != nil {
		t.Fatal(err)
	}
	defer jsRes.Body.Close()
	if jsRes.StatusCode != 200 {
		t.Fatalf("app.js status = %d", jsRes.StatusCode)
	}
	js, _ := io.ReadAll(jsRes.Body)
	if !strings.Contains(string(js), "EventSource") {
		t.Error("app.js missing EventSource client")
	}
}

func TestServer_StreamHistoryAndLive(t *testing.T) {
	s := New(Config{})
	// publish history BEFORE client connects
	s.Hub().Publish(parse.Event{
		Kind: parse.KindUser,
		Blocks: []parse.Block{
			{Kind: parse.BlockUserPrompt, Text: "hallo"},
		},
	})

	ts := httptest.NewServer(s)
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, "GET", ts.URL+"/stream", nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if ct := res.Header.Get("Content-Type"); !strings.Contains(ct, "text/event-stream") {
		t.Errorf("Content-Type = %q", ct)
	}

	events := make(chan parse.Event, 8)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		br := bufio.NewReader(res.Body)
		for {
			line, err := br.ReadString('\n')
			if err != nil {
				return
			}
			line = strings.TrimRight(line, "\r\n")
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			payload := strings.TrimPrefix(line, "data: ")
			var ev parse.Event
			if err := json.Unmarshal([]byte(payload), &ev); err != nil {
				continue
			}
			events <- ev
		}
	}()

	// 1st event: history
	select {
	case ev := <-events:
		if ev.Kind != parse.KindUser || len(ev.Blocks) != 1 || ev.Blocks[0].Text != "hallo" {
			t.Fatalf("history event wrong: %+v", ev)
		}
	case <-time.After(time.Second):
		t.Fatal("no history event")
	}

	// publish live event
	s.Hub().Publish(parse.Event{
		Kind: parse.KindAssistant,
		Blocks: []parse.Block{
			{Kind: parse.BlockText, Text: "live!"},
		},
	})

	select {
	case ev := <-events:
		if ev.Kind != parse.KindAssistant || ev.Blocks[0].Text != "live!" {
			t.Fatalf("live event wrong: %+v", ev)
		}
	case <-time.After(time.Second):
		t.Fatal("no live event")
	}

	cancel()
	wg.Wait()
}

func TestHub_SlowConsumerDoesNotBlock(t *testing.T) {
	h := newHub()
	// subscribe but never drain
	_, _, unsub := h.Subscribe()
	defer unsub()

	// publishing should not block even with full channel
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 10_000; i++ {
			h.Publish(parse.Event{Kind: parse.KindUser})
		}
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Publish blocked on full subscriber buffer")
	}
}
