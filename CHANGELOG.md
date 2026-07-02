# Changelog

All notable changes to this project are documented here.

Format: [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
Versioning follows [SemVer](https://semver.org/).

## [Unreleased]

## [0.4.0] - 2026-07-02

### Added

- **tail-html — a growing HTML transcript per session, toggle on/off.** Context
  menu → "Protokoll starten/stoppen" records a session's events into a
  continuously growing `~/.claude/ccview/protocols/<id>.html` while on, and
  freezes it when off — turn it on for a long sequence, off, then read it in
  peace while the session keeps running. Turning it back on **appends onto the
  existing file** (a resume marker in the HTML says where to continue; the
  #NNNN numbering carries over) instead of rebuilding, with a visible
  "Aufnahme fortgesetzt/pausiert" divider marking each window. The recorder
  tails the JSONL directly (independent of the live viewer, so it keeps
  recording even when you switch sessions), survives a server/systemd restart
  (active sessions resume), and is always rebuildable from the JSONL if a marker
  goes stale. "Protokoll öffnen" serves the file; while recording it auto-tails
  (reloads only while scrolled to the bottom). The auto-refresh interval is
  configurable in Settings (seconds, 0 = off) and can be overridden per open
  with a `?refresh=N` query parameter.
- **Session start and last-event times** (date & time) in the sidebar hover
  popup, read from the JSONL content (stable across cross-writes). The tile line
  next to the project path now shows the start–last span instead of a single
  relative stamp: `dd.mm.yy–dd.mm.yy` across days, or `dd.mm.yy HH:MM–HH:MM` for
  a same-day session (start and end always both visible, never one ambiguous date).

### Changed

- **Named sessions get their own group at the very top** of the sidebar, sorted
  alphabetically, and are excluded from the "Active" and per-project groups
  (a session with a custom title no longer appears twice).
- **Pinned (favorite) sessions are always in the "Active" group, pinned to the
  top** — not only when they have activity today. They stay there from pinning
  until you unpin, and remain reorderable among themselves (▲▼).

## [0.3.3] - 2026-06-10

### Changed

- **Per-session resume button.** Each session tile has a ▶ button (left of the
  burger) that copies `cd '<cwd>' && claude --resume <id>` to the clipboard —
  paste into any terminal to resume in the right directory (fixes the old
  "wrong cwd → no sessions match" trap). Replaces the single top-left current-
  session copy button, which is removed.
- **Session tile header reworked.** The two indistinguishable stars are now
  distinct: a **gold ★** marks a favorite (favbar), a **📌** marks the main
  (startup-default) session — both with tooltips. New header layout puts the two
  toggles and the id on the left and the actions burger on the right (inline, so
  it no longer overlaps a star), with the custom name on its own line below.

## [0.3.2] - 2026-06-09

### Fixed

- **Project labels resolve correctly** instead of guessing. The sidebar showed
  `.../it/sps/sim` for a session in `cuber-it/sps-sim` — it just replaced every
  `-` in the encoded project-dir name with `/`, which mangles directory names
  that contain hyphens. Labels now read back the real path from the filesystem
  (resolving the lossy hyphen encoding) and show `org/repo`, e.g.
  `cuber-it/sps-sim`. Memoized so the filesystem probe runs once per project.

### Added

- **Full project path in the session hover popup**, on its own line above the
  start prompt, so it is clear which working directory a session belongs to.

## [0.3.1] - 2026-06-09

### Added

- **Notes Markdown preview toggle** (👁 / ✎ in the notes toolbar): switches the
  textarea for a rendered view and back. The preview reuses the main timeline
  Markdown renderer (shared via `window.ccviewRenderMd`) — no second Markdown
  implementation, nothing heavy to load, and nothing that can break on
  session-switch the way the old EasyMDE editor did. Re-renders on switch while
  open; the formatting buttons hide in preview mode.
- **GFM tables now render** (header + `---` separator + rows) — in the notes
  preview *and* in rendered session text (`text` / `user_prompt` blocks), which
  previously dropped table markup as plain paragraphs.

### Changed

- **Pinned notes resize handle widened** from 6px to a 12px hit area with a
  visible center grip that brightens on hover and while dragging, so the panel
  edge is far easier to grab.

### Docs / Packaging

- **INSTALL.md** handbook (systemd + Docker setup) plus `Dockerfile`,
  `docker-compose.yml`, `.dockerignore`. README install section points to it.
- **Release pipeline** (`.goreleaser.yaml` + `release.yml` workflow): tagging
  `vX.Y.Z` builds binaries for Linux/macOS/Windows (amd64 + arm64), attaches them
  with checksums to a GitHub release, and pushes a multi-arch image to
  `ghcr.io/cuber-it/ccview`.

### Fixed

- **Stale frontend after updates**: the embedded assets were served without
  cache headers, so browsers kept an old `app.js`/`style.css` across a server
  update — features the new HTML expected (e.g. pinned-notes layout) silently
  broke. Embedded assets are now served `Cache-Control: no-store`.
- **Notes editor crashed** after the toolbar tooltips were translated: the
  notes floater is a separate IIFE and could not see the `t()` translator,
  so EasyMDE init threw `ReferenceError: t is not defined` and the editor never
  loaded. The translator is now shared and guarded with a fallback.

### Changed

- **Notes autosave silently to the DB** (while typing, on session-switch, on
  close) — no confirm dialog. The save button (and Ctrl-S) now **exports the
  note as a downloadable `.md` file** instead of saving (persistence is
  automatic). Added a **table** button to the notes toolbar.
- **Favorites bar always highlights the current session green**, not just chips
  with new events — so the one you're viewing is never the colorless one.
- **Notes are now a plain textarea** with a small built-in Markdown toolbar
  (bold/italic/heading/code/quote/lists/link), replacing the EasyMDE/CodeMirror
  editor. The rich editor needed display-refresh juggling that broke on open and
  on session-switch; the textarea opens/closes reliably and switches cleanly.
  Closing with unsaved edits asks before discarding; the white writing surface
  stays.
- **Large sessions render fast**: a fresh stream now replays only the last 800
  events with a "load older" control to page further back, clamp measurement is
  deferred to blocks scrolling into view (no per-event layout thrash), and
  scroll-to-bottom is coalesced to one per frame.
- **Switching into a huge session no longer freezes**: the server now separates
  history from live. `pump` reads the whole file into the hub without streaming
  it to clients, then pushes only the capped tail (reset + meta + last 800) on
  the live boundary; live events fan out as before. The browser renders ~hundreds
  of nodes instead of tens of thousands (longest main-thread block dropped from
  >20 s to ~0.1 s), with "load older" paging the rest from the hub.
- **Central SQLite store** (`~/.claude/ccview/ccview.db`, pure-Go driver, honours
  `CLAUDE_CONFIG_DIR`) now holds config, project roots, per-session notes, custom
  names/favorites, and project-group settings — replacing the earlier mix of
  `roots.json`, `notes/*.md`, and browser localStorage. Names and notes are now
  the same across browsers and ports instead of per-origin. Legacy files migrate
  into the DB on first start (idempotent, never overwrites). Also fixes a
  long-standing bug where per-session notes were never actually persisted.

### Added

- **HTML transcript** per session (context menu → "HTML-Transkript"): renders the
  full session as a standalone HTML document and opens it in a new tab. The
  browser lays it out natively in one pass — fast and Ctrl-F-searchable even for
  huge sessions that freeze the live viewer. Cached in memory (invalidated by
  file modTime), never written to disk.
- **Generic config store** (`/api/config`, `config` table): the "hide done"
  filter now persists server-side in the DB instead of browser localStorage, so
  it survives across browsers/profiles (legacy localStorage value migrated once).
- **Multiple project roots**: scan several Claude projects directories at
  once. Manage the list in Settings → "Projekt-Pfade" (stored in the central
  SQLite DB; `~` expanded, duplicates dropped). New
  `--projects-root` flag and `CLAUDE_CONFIG_DIR` support.
- **Settings dialog** (burger menu): curate sidebar project groups —
  display name, order, and visibility (localStorage).
- **Sidebar grouping**: an "Active" group (current / same-project / today)
  on top, remaining sessions grouped by project, each collapsible.
- **Per-session notes**: file-backed Markdown notepad with an EasyMDE editor
  (toolbar, live preview, syntax highlighting); floats or pins to the right
  with a draggable width.
- **Session context menu** (mini-burger / right-click): favorite, rename
  (persistent custom name shown instead of the ID), save, copy ID.
- **Session-list filter** and **auto-refresh** of the session list.
- **Reversible prompt index** and **resizable sidebar** width (both persisted).
- **Dev mode**: `CCVIEW_DEV=<dir>` serves `static/` from disk (no-cache) for
  live editing without a rebuild.
- **Collapse/expand all** sidebar groups with one button in the Sessions tab.
- **Done marker** per session (context menu) plus a filter to hide/show done
  sessions — the basis for later archiving/deletion.
- **Search modal** (toolbar button, next to Notes): case-insensitive regex over
  three scopes — current session, all sessions, all notes — with hit count,
  activity days, and snippets. Reads line by line, handles huge logs.
- **Delete to trash**: context-menu "Löschen…" moves a session's JSONL to
  `~/.claude/ccview/trash/` (reversible) and drops its DB metadata; confirm first.
- **Read-only query box** in its own modal (burger menu): a single SELECT against the metadata DB.
- **Built-in cheatsheet** of Claude Code slash commands, opened in its own tab
  with `?` (or the burger menu), linking to the official docs.

### Changed

- The sidebar "Active" group no longer counts `same_project` (same cwd as the
  server process) as active — only the currently-open session and today's
  sessions. Running ccview as a service from a fixed directory no longer
  floods "Active"; those sessions form a normal collapsible group instead.
- Settings lists every detected project (previously only those with inactive
  sessions), so any group can be curated.

## [0.2.0] — 2026-04-27

### Changed

- Stronger visual separation in the timeline: real user prompts get a 6 px
  orange left bar, assistant events a 6 px cyan/teal bar (complementary
  hues, distinguishable for color-blind users), tool-context user events
  stay muted. Header labels colored to match. Per-theme tuning for
  Dark / Light / Sepia.
- Markdown rendering in `text` and `user_prompt` blocks now covers headings
  (`#`–`######`), unordered and ordered lists, blockquotes, links, horizontal
  rules, and strikethrough — on top of the existing fenced/inline code,
  bold, and italic. `tool_result` stays raw `<pre>` (matches Claude Code).
- Per-block collapse with type-specific defaults: real user prompts 5 lines,
  assistant blocks (text / thinking / tool_use) 3 lines each, tool_result
  1 line. Above the cap, a `mehr` / `more` button appears. For very long
  content (> 10 lines), a second click reveals a 10-line preview marked
  `<WEITER>` / `<MORE>` (localized); a third click shows everything.
  Removes the previous fixed 320 px cap on tool_result and the hardcoded
  `truncate(...)` of Edit and Write tool inputs (200 / 400 chars) — full
  content is now reachable.

### Refactor

- Frontend: split monolithic `internal/srv/static/index.html` (1945 lines) into
  `index.html` (82 lines), `style.css` (~830 lines), `app.js` (~1140 lines).
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

[Unreleased]: https://github.com/cuber-it/ccview/compare/v0.4.0...HEAD
[0.4.0]: https://github.com/cuber-it/ccview/releases/tag/v0.4.0
[0.3.3]: https://github.com/cuber-it/ccview/releases/tag/v0.3.3
[0.3.2]: https://github.com/cuber-it/ccview/releases/tag/v0.3.2
[0.3.1]: https://github.com/cuber-it/ccview/releases/tag/v0.3.1
[0.3.0]: https://github.com/cuber-it/ccview/releases/tag/v0.3.0
[0.2.0]: https://github.com/cuber-it/ccview/releases/tag/v0.2.0
[0.1.0]: https://github.com/cuber-it/ccview/releases/tag/v0.1.0
