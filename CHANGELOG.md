# Changelog

Alle nennenswerten Änderungen an diesem Projekt werden hier dokumentiert.

Format: [Keep a Changelog](https://keepachangelog.com/de/1.1.0/).
Versionierung nach [SemVer](https://semver.org/lang/de/).

## [Unreleased]

## [0.1.0] — 2026-04-24

Erste nutzbare Version. Live-Viewer für Claude-Code-Sessions mit Browser-UI,
Session-Switcher, Favoriten, Markdown-Export.

### Added — Core

- Go-Modul `github.com/cuber-it/ccview`, Go 1.23
- `internal/parse` — JSONL-Event-Parser, typsicher, forward-kompatibel
  (unbekannte Event- und Block-Typen überleben ohne Fehler)
- `internal/tail` — Polling-Tailer (100 ms) mit Partial-Line-Handling
  für beliebig lange Zeilen, race-free
- `internal/session` — Projekt-Discovery, `List` pro Projekt, `ListAll`
  über alle Projekte, `Resolve` mit Prefix-Match, `ReadFirstEventTime`
  (Session-Start aus JSONL — stabil gegen Cross-Writes)
- `internal/srv` — HTTP-Server, SSE-Hub, Session-Switching zur Laufzeit,
  embedded Frontend (`//go:embed`)
- `internal/export` — Markdown-Renderer mit Prompt-Nummerierung,
  Edit-Diffs, Tool-Input-Prettifier, Thinking in `<details>`, Images
- `cmd/ccview` — CLI-Entry, Port-Fallback 12100..12199, Browser-Auto-Open

### Added — Frontend

- Vanilla-JS, ein embedded `index.html`, kein Build-Schritt
- Drei Themes (Dark / Light / Sepia), in `localStorage` persistiert
- Sidepanel ein-/ausklappbar mit Tabs `Prompts` / `Sessions`
- Prompt-Index mit `#NNNN`-Nummerierung, Sprungmarken, Filter-Input,
  Hover-Popup mit vollem Prompt-Text
- Sessions-Tab listet alle Sessions aller Projekte, mit Projekt-Label,
  Pin-Stern, Klick = Switch zur Laufzeit, Heute-Rahmen
- Favoriten-Leiste oben (max 5), hellgrün wenn neue Events seit letztem
  Öffnen (15-s-Poll, visibility-aware)
- Markdown-Light (Code-Fences, Bold, Italic, Inline-Code)
- Tool-Input-Prettifier (Bash / Read / Edit / Write / Grep / Glob)
- Image-Block-Rendering (base64 + URL)
- Echte User-Prompts visuell hervorgehoben, Tool-Result-User umgelabelt
- Kopier-Button pro Event-Karte (strukturierter Plain-Text)
- Burger-Menü: Speichern, Speichern unter…, Über ccview
- Info-Modal mit Version + Links (uc-it.de, GitHub)
- Bottom-Command-Bar: Scroll-Pause-Toggle, Anfang, Live, Keyboard-Hint

### Added — Produktivität

- Auto-Scroll-Pause mit "↓ Live"-Pill bei neuen Events wenn pausiert
- Live-Suche (`/` öffnen) filtert Events, zählt Treffer, scrollt ins Bild
- Event-/Prompt-Zähler in der Toolbar
- Keyboard-Nav: `/` `j` `k` `g g` `G` `Esc`
- Filter-Input im Prompts-Tab
- Interrupt-Prompts (`queue-operation enqueue`) werden als User-Prompts gerendert

### Added — Build & Release

- Makefile (`build`, `test`, `vet`, `race`, `cross`, `clean`)
- Cross-Compile für linux-amd64/arm64, darwin-amd64/arm64, windows-amd64
- `--version` Flag mit Versionsinjektion via `-ldflags -X main.version`
- `/api/version` Endpoint

### Config

- **CLI** `ccview` (ohne Args) startet den Viewer und merkt die zuletzt
  geöffnete Session via `localStorage`
- Keine Config-Datei — alles per localStorage (Theme, Sidepanel-State,
  Favoriten, Tab, zuletzt geöffnete Session) oder CLI-Flag

[Unreleased]: https://github.com/cuber-it/ccview
[0.1.0]: https://github.com/cuber-it/ccview/releases/tag/v0.1.0
