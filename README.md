# ccview

Live-Viewer für Claude Code Sessions.

Tailt die JSONL einer Session unter `~/.claude/projects/<projekt>/<session-id>.jsonl`
und rendert sie im Browser — Messages, Tool-Calls und Tool-Results im Stil der
claude.ai-Oberfläche, lokal, ohne Cloud.

## Status

Work in progress. Alpha.

## Nutzung

```bash
ccview                         # Liste aller Sessions im aktuellen Projekt
ccview -s <id|prefix|latest>   # Session im Browser öffnen (live)
ccview --session <...>         # Langform
ccview --no-browser            # nur URL ausgeben (für SSH-Tunnel / headless)
ccview --port N                # Port überschreiben (Default: ab 12100)
ccview --bind 0.0.0.0          # für LAN-Zugriff freigeben
```

Ein Prozess pro geöffnete Session. Strg-C beendet.

## Installation

```bash
go install github.com/cuber-it/ccview@latest
```

Oder Binary aus dem GitHub-Release ziehen (Linux / macOS / Windows).

## Build aus Quelltext

```bash
git clone https://github.com/cuber-it/ccview
cd ccview
go build -o ccview ./cmd/ccview
```

## Lizenz

Apache 2.0 © 2026 Ulrich Cuber / cuber IT service
