# ccview

Live viewer for Claude Code sessions.

`ccview` tails the JSONL of a Claude Code session under `~/.claude/projects/`
and renders it in your browser — messages, tool calls, tool results, thinking,
images — in a chat-client style. Fully local, no cloud, no account.

A single Go binary (~5 MB). No runtime, no dependencies.

## Features

- **Live tail** of any session in the browser, via SSE
- **Session browser** across all projects, runtime switching with a click
- **Multiple project roots** scanned together, managed in the Settings dialog
- **Grouped sidebar** (Active + per-project, collapsible) with resizable width,
  filter, auto-refresh, and per-session rename
- **Cross-session search**: regex over the current session, all sessions, or all
  notes — with hit counts and activity days
- **Done marker** + filter to hide finished sessions; **delete** moves the JSONL
  to a reversible trash
- **Per-session notes**: Markdown notepad (EasyMDE), stored in the central DB
- **Read-only SQL console** over ccview's metadata DB, plus a built-in
  **slash-command cheatsheet** (press `?`)
- **Prompt index** in the sidepanel: numbered anchors, filter, hover popup
- **Favorites** (up to 5) with light-green highlight on new events
- **Main session** (exclusive) auto-loads on startup regardless of cwd
- **Today frame** for sessions with activity on the current day
- **Themes** Dark / Light / Sepia, persisted in localStorage
- **Language** DE / EN toggle, persisted
- **Markdown export** of the active session (Save / Save As)
- **Copy button** per event — structured plain text to clipboard
- **Productivity**: auto-scroll pause, live search, keyboard nav, event counter
- **Interrupt prompts** (Claude Code `queue-operation`) are rendered visibly
- **Cross-platform** binaries for Linux, macOS, Windows (amd64 + arm64)
  — Linux/macOS tested, Windows untested (feedback welcome)

## Usage

```bash
ccview                         # start server, open browser, pick session in UI
ccview -s <id|prefix|latest>   # open a specific session right away
ccview --no-browser            # only print the URL (SSH tunnel / headless)
ccview --port N                # override port (default: 12100..12199)
ccview --bind 0.0.0.0          # expose to the LAN
ccview --verbose               # print status messages
ccview --version               # show version
```

One process per instance. The last-opened session is remembered in the
browser and auto-loaded next time. Mark a session as "main" (blue star)
to make it the startup default everywhere.

### Keyboard

```
/        live search in events
Esc      close search / menu / modal
j / k    next / previous prompt
g g      top
G        end (live)
```

### Burger menu (top right)

- **Save** — markdown export to `~/Workspace/claude-code/sessions/<proj>_<date>_<short-id>.md`
- **Save As…** — choose filename or path
- **About ccview** — version, links

## Install

```bash
go install github.com/cuber-it/ccview/cmd/ccview@latest
```

Or download a binary from
[GitHub Releases](https://github.com/cuber-it/ccview/releases).

To run it as a persistent **service** — **systemd** or **Docker** — see
**[INSTALL.md](INSTALL.md)**. ccview has no authentication, so bind it to
`127.0.0.1`. State lives in a small SQLite DB at `~/.claude/ccview/ccview.db`;
your Claude sessions are only ever read.

## Build from source

```bash
git clone https://github.com/cuber-it/ccview
cd ccview
make build          # → ./ccview
make test           # go test ./...
make cross          # cross-compile all platforms into dist/
```

Built with Go (see `go.mod`). The SQLite driver is pure Go, so
`CGO_ENABLED=0` produces a fully static binary.

## Architecture

```
~/.claude/projects/*/*.jsonl     (written by Claude Code)
        │
        ▼  polling tailer
  ┌────────────┐  parse  ┌──────┐  SSE   ┌──────────────┐
  │ internal/  │────────▶│ Hub  │───────▶│  browser tab │
  │ tail       │         │      │        │  (embedded)  │
  └────────────┘         └──────┘        └──────────────┘
         ▲                                        │
         └── /api/switch ◀── click ───────────────┘
```

- `internal/parse` — JSONL events → typed struct
- `internal/tail` — polling tailer with partial-line handling
- `internal/session` — project/session discovery, sort by `FirstEventTime`
- `internal/srv` — HTTP server, SSE hub, runtime session switching, embedded frontend
- `internal/export` — markdown renderer

Frontend: a single `index.html` served via `//go:embed`, vanilla JS, no build step.

## License

Apache 2.0 — © 2026 Ulrich Cuber / cuber IT service &middot;
[www.uc-it.de](https://www.uc-it.de)
