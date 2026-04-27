# Changelog

All notable changes to this project are documented here.

Format: [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
Versioning follows [SemVer](https://semver.org/).

## [Unreleased]

### Refactor

- Frontend: split monolithic `internal/srv/static/index.html` (1945 lines) into
  `index.html` (82 lines), `style.css` (781 lines), `app.js` (1080 lines).
  No build step, served via the same `//go:embed static` tree. Test updated
  to verify the `app.js` reference and fetch.
- `srv/server.go: handleStream` split into a new `srv/sse.go` with `sseWriter`
  (SSE protocol) and helpers `replayHistory` / `streamLive`. handleStream now
  orchestrates only.
- `srv/server.go: handleExport` split: `lookupStartedTime` and
  `writeMarkdownExport` extracted; the handler now orchestrates only.

## [0.1.0] — 2026-04-24

First usable release. Live viewer for Claude Code sessions with a browser UI,
runtime session switching, favorites, and markdown export.

### Core

- Go module `github.com/cuber-it/ccview`, Go 1.23
- `internal/parse` — typed JSONL event parser, forward-compatible
  (unknown event and block types are tolerated)
- `internal/tail` — polling tailer (100 ms) with partial-line handling
  for arbitrarily long lines, race-free
- `internal/session` — project discovery, `List` per project, `ListAll`
  across all projects, `Resolve` with prefix match, `ReadFirstEventTime`
  (session start from the JSONL — stable against cross-writes)
- `internal/srv` — HTTP server, SSE hub, runtime session switching,
  embedded frontend (`//go:embed`)
- `internal/export` — markdown renderer with prompt numbering,
  Edit diffs, tool-input prettifier, thinking in `<details>`, images
- `cmd/ccview` — CLI entrypoint, port fallback 12100..12199, auto-open browser

### Frontend

- Vanilla JS, a single embedded `index.html`, no build step
- Three themes (Dark / Light / Sepia), persisted in `localStorage`
- Collapsible sidepanel with `Prompts` / `Sessions` tabs
- Prompt index with `#NNNN` numbering, anchor links, filter input,
  hover popup showing the full prompt
- Sessions tab lists every session across every project, with project
  label, pin star, click = runtime switch, today frame
- Favorites bar (max 5) with light-green tint when new events arrive
  since last viewed (15-second poll, visibility-aware)
- Main-session star (exclusive) that loads on startup
- Light markdown (code fences, bold, italic, inline code)
- Tool-input prettifier (Bash / Read / Edit / Write / Grep / Glob)
- Image block rendering (base64 + URL)
- Real user prompts visually highlighted, tool-result users relabeled
- Per-event copy button (structured plain text)
- Burger menu: Save, Save As…, About ccview
- About modal with version + links (uc-it.de, GitHub)
- Bottom command bar: scroll-pause toggle, top, live, keyboard hint
- DE / EN language toggle, persisted

### Productivity

- Auto-scroll pause with "↓ Live" pill on new events when paused
- Live search (press `/`) filters events, counts hits, scrolls to first match
- Event and prompt counter in the toolbar
- Keyboard: `/` `j` `k` `g g` `G` `Esc`
- Filter input in the Prompts tab
- Interrupt prompts (`queue-operation enqueue`) rendered as user prompts

### Build & release

- Makefile (`build`, `test`, `vet`, `race`, `cross`, `clean`)
- Cross-compile for linux-amd64/arm64, darwin-amd64/arm64, windows-amd64
- `--version` flag, version injected via `-ldflags -X main.version`
- `/api/version` endpoint

### Config

- **CLI** `ccview` (no args) starts the viewer and remembers the last-opened
  session via `localStorage`
- No config file — everything is either a CLI flag or `localStorage`
  (theme, language, sidepanel state, favorites, main session, last session, tab)

[Unreleased]: https://github.com/cuber-it/ccview
[0.1.0]: https://github.com/cuber-it/ccview/releases/tag/v0.1.0
