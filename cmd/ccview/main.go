// Command ccview is a live viewer for Claude Code sessions.
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

	"github.com/cuber-it/ccview/internal/session"
	"github.com/cuber-it/ccview/internal/srv"
)

// Set at build time via -ldflags "-X main.version=..."
var version = "dev"

func main() {
	var (
		sess        string
		port        int
		bind        string
		noBrowser   bool
		verbose     bool
		showVersion bool
	)
	flag.StringVar(&sess, "session", "", "session id, prefix, or 'latest'")
	flag.StringVar(&sess, "s", "", "shorthand for --session")
	flag.IntVar(&port, "port", 0, "HTTP port (0 = auto-pick 12100..12199)")
	flag.StringVar(&bind, "bind", "127.0.0.1", "bind address")
	flag.BoolVar(&noBrowser, "no-browser", false, "do not open browser")
	flag.BoolVar(&verbose, "verbose", false, "print status messages to stdout/stderr")
	flag.BoolVar(&verbose, "v", false, "shorthand for --verbose")
	flag.BoolVar(&showVersion, "version", false, "print version and exit")
	flag.Parse()

	if showVersion {
		fmt.Println("ccview", version)
		return
	}
	if err := run(sess, port, bind, noBrowser, verbose); err != nil {
		fmt.Fprintln(os.Stderr, "ccview:", err)
		os.Exit(1)
	}
}

func run(sessSpec string, port int, bind string, noBrowser, verbose bool) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	projectsRoot, err := session.ProjectsDir()
	if err != nil {
		return err
	}
	projectDir := session.ProjectDirFromCwd(cwd)

	var initial *session.Info
	if sessSpec != "" {
		sessions, err := session.List(projectsRoot, projectDir)
		if err != nil {
			return fmt.Errorf("list sessions for %s: %w", projectDir, err)
		}
		info, err := session.Resolve(sessions, sessSpec)
		if err != nil {
			return err
		}
		initial = &info
	}

	ln, addr, err := listenWithFallback(bind, port)
	if err != nil {
		return err
	}
	url := "http://" + addr
	fmt.Println(url)
	if verbose {
		if initial != nil {
			fmt.Fprintf(os.Stderr, "ccview: session %s\n", initial.ID)
		} else {
			fmt.Fprintln(os.Stderr, "ccview: Session im Browser wählen")
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		if verbose {
			fmt.Fprintln(os.Stderr, "\nccview: shutting down")
		}
		cancel()
	}()

	s := srv.New(srv.Config{
		ProjectsRoot: projectsRoot,
		ProjectDir:   projectDir,
		Version:      version,
		Verbose:      verbose,
	})
	if !noBrowser {
		go openBrowser(url)
	}
	return s.Serve(ctx, ln, initial)
}

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
		if ln, addr, err := try(p); err == nil {
			return ln, addr, nil
		} else {
			lastErr = err
		}
	}
	return nil, "", fmt.Errorf("no free port in 12100..12199: %w", lastErr)
}

func openBrowser(url string) {
	time.Sleep(150 * time.Millisecond)
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
