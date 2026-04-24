// Command ccview is a live viewer for Claude Code sessions.
//
// Usage:
//
//	ccview                         list sessions in the current project
//	ccview -s <id|prefix|latest>   open a session in the browser
//	ccview --session <...>         long form
//	ccview --no-browser            do not open browser; print URL only
//	ccview --port N                override port
//	ccview --bind 0.0.0.0          bind address (default 127.0.0.1)
package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/cuber-it/ccview/internal/parse"
	"github.com/cuber-it/ccview/internal/session"
	"github.com/cuber-it/ccview/internal/srv"
	"github.com/cuber-it/ccview/internal/tail"
)

// version is set via -ldflags -X at build time.
var version = "dev"

func main() {
	var (
		sess        string
		port        int
		bind        string
		noBrowser   bool
		showVersion bool
	)
	flag.StringVar(&sess, "session", "", "session id, prefix, or 'latest'")
	flag.StringVar(&sess, "s", "", "shorthand for --session")
	flag.IntVar(&port, "port", 0, "HTTP port (0 = auto-pick 12100..12199)")
	flag.StringVar(&bind, "bind", "127.0.0.1", "bind address")
	flag.BoolVar(&noBrowser, "no-browser", false, "do not open browser")
	flag.BoolVar(&showVersion, "version", false, "print version and exit")
	flag.Parse()

	if showVersion {
		fmt.Println("ccview", version)
		return
	}

	if err := run(sess, port, bind, noBrowser); err != nil {
		fmt.Fprintln(os.Stderr, "ccview:", err)
		os.Exit(1)
	}
}

func run(sessSpec string, port int, bind string, noBrowser bool) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	projectsRoot, err := session.ProjectsDir()
	if err != nil {
		return err
	}
	projectDir := session.ProjectDirFromCwd(cwd)
	sessions, err := session.List(projectsRoot, projectDir)
	if err != nil {
		return fmt.Errorf("cannot list sessions for %s: %w", projectDir, err)
	}

	if sessSpec == "" {
		printSessionList(sessions, projectDir)
		return nil
	}

	info, err := session.Resolve(sessions, sessSpec)
	if err != nil {
		return err
	}

	ln, addr, err := listenWithFallback(bind, port)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("http://%s", addr)

	fmt.Printf("ccview: session %s\nccview: %s\n", info.ID, url)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		fmt.Fprintln(os.Stderr, "\nccview: shutting down")
		cancel()
	}()

	s := srv.New()

	// tail → parse → publish
	go pump(ctx, info.Path, s.Hub(), cancel)

	if !noBrowser {
		go openBrowser(url)
	}

	return s.Serve(ctx, ln)
}

func pump(ctx context.Context, path string, hub *srv.Hub, cancel context.CancelFunc) {
	ch := tail.New(path).Stream(ctx)
	for l := range ch {
		if l.Err != nil {
			fmt.Fprintln(os.Stderr, "ccview: tail:", l.Err)
			cancel()
			return
		}
		ev, err := parse.Parse(l.Data)
		if err != nil {
			continue
		}
		hub.Publish(ev)
	}
}

// listenWithFallback binds bind:port. port==0 → try 12100..12199 in order.
func listenWithFallback(bind string, port int) (net.Listener, string, error) {
	try := func(p int) (net.Listener, string, error) {
		addr := fmt.Sprintf("%s:%d", bind, p)
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			return nil, "", err
		}
		return ln, addr, nil
	}
	if port != 0 {
		return try(port)
	}
	var lastErr error
	for p := 12100; p < 12200; p++ {
		ln, addr, err := try(p)
		if err == nil {
			return ln, addr, nil
		}
		lastErr = err
	}
	return nil, "", fmt.Errorf("no free port in 12100..12199: %w", lastErr)
}

func openBrowser(url string) {
	time.Sleep(150 * time.Millisecond) // let the server bind
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return
	}
	_ = cmd.Start()
}

func printSessionList(sessions []session.Info, projectDir string) {
	if len(sessions) == 0 {
		fmt.Printf("ccview: no sessions found in ~/.claude/projects/%s\n", projectDir)
		return
	}
	fmt.Printf("%-8s  %-19s  %8s  %s\n", "ID", "Last", "Size", "Path")
	for _, s := range sessions {
		fmt.Printf("%-8s  %-19s  %8s  %s\n",
			s.ID[:8],
			s.ModTime.Format("2006-01-02 15:04:05"),
			fmtSize(s.Size),
			s.Path,
		)
	}
}

func fmtSize(n int64) string {
	switch {
	case n < 1024:
		return fmt.Sprintf("%dB", n)
	case n < 1024*1024:
		return fmt.Sprintf("%.1fK", float64(n)/1024)
	default:
		return fmt.Sprintf("%.1fM", float64(n)/(1024*1024))
	}
}
