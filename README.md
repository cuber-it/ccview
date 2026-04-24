# ccview

Live-Viewer für Claude Code Sessions.

`ccview` tailt die JSONL einer Claude-Code-Session unter `~/.claude/projects/`
und rendert sie im Browser — Messages, Tool-Calls, Tool-Results, Thinking,
Bilder — im Stil eines Chat-Clients. Rein lokal, keine Cloud, kein Account.

Ein Go-Binary (~5 MB). Kein Runtime, keine Dependencies.

## Features

- **Live-Tail** jeder Session im Browser, mit SSE-Stream
- **Sessions-Browser** über alle Projekte hinweg, Switch zur Laufzeit per Klick
- **Prompt-Index** im Sidepanel: nummerierte Sprungmarken, Filter, Hover-Popup
- **Favoriten** (max 5) mit Hellgrün-Highlight bei neuen Events
- **Heute-Rahmen** für Sessions mit Aktivität am aktuellen Tag
- **Themes** Dark / Light / Sepia, localStorage-persistent
- **Markdown-Export** der aktuellen Session (Speichern / Speichern unter)
- **Kopier-Button** pro Event — Inhalt strukturiert ins Clipboard
- **Produktivitäts-Bundle**: Auto-Scroll-Pause, Live-Suche, Keyboard-Nav, Event-Zähler
- **Interrupt-Prompts** (Claude Codes `queue-operation`) werden sichtbar
- **Cross-Platform** Binaries für Linux, macOS, Windows (amd64 + arm64) — Linux/macOS getestet, Windows untested (Feedback erwünscht)

## Nutzung

```bash
ccview                         # startet Server + Browser, Session wählen
ccview -s <id|prefix|latest>   # direkt eine Session öffnen
ccview --no-browser            # nur URL ausgeben (SSH-Tunnel / headless)
ccview --port N                # Port überschreiben (Default: 12100..12199)
ccview --bind 0.0.0.0          # LAN-Freigabe
ccview --version               # Version anzeigen
```

Ein Prozess pro Instanz. Die zuletzt geöffnete Session wird im Browser gemerkt
und beim nächsten Start automatisch geladen.

### Keyboard

```
/        Live-Suche in Events
Esc      Suche / Menü / Modal schließen
j / k    nächster / vorheriger Prompt
g g      Anfang
G        Ende (Live)
```

### Burger-Menü (oben rechts)

- **Speichern** — Markdown-Export nach `~/Workspace/claude-code/sessions/<proj>_<datum>_<short-id>.md`
- **Speichern unter…** — Dateiname oder Pfad wählbar
- **Über ccview** — Version, Links

## Installation

```bash
go install github.com/cuber-it/ccview/cmd/ccview@latest
```

Oder ein Release-Binary aus
[GitHub Releases](https://github.com/cuber-it/ccview/releases) ziehen.

## Build aus dem Quelltext

```bash
git clone https://github.com/cuber-it/ccview
cd ccview
make build          # → ./ccview
make test           # go test ./...
make cross          # alle Plattform-Binaries nach dist/
```

Getestet mit Go 1.23.

## Architektur (kurz)

```
~/.claude/projects/*/*.jsonl   (Claude Code schreibt)
        │
        ▼  polling-tailer
  ┌────────────┐  parse  ┌──────┐  SSE  ┌──────────────┐
  │ internal/  │────────▶│ Hub  │──────▶│  Browser-Tab │
  │ tail       │         │      │       │  (embed HTML)│
  └────────────┘         └──────┘       └──────────────┘
         ▲                                      │
         └── /api/switch ◀── click ─────────────┘
```

- `internal/parse` — JSONL-Events → typisierte Struct
- `internal/tail` — Polling-Tailer mit Partial-Line-Handling
- `internal/session` — Projekt-/Session-Discovery, Sort nach `FirstEventTime`
- `internal/srv` — HTTP-Server, SSE-Hub, Session-Switching, embedded Frontend
- `internal/export` — Markdown-Renderer

Frontend: eine einzige `index.html` via `//go:embed`, Vanilla-JS, ohne Build-Schritt.

## Lizenz

Apache 2.0 — © 2026 Ulrich Cuber / cuber IT service &middot;
[www.uc-it.de](https://www.uc-it.de)
