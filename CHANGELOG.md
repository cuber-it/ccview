# Changelog

Alle nennenswerten Änderungen an diesem Projekt werden hier dokumentiert.

Format: [Keep a Changelog](https://keepachangelog.com/de/1.1.0/).
Versionierung nach [SemVer](https://semver.org/lang/de/).

## [Unreleased]

### Added

- Projekt-Scaffold: Go-Modul, Repo-Struktur, README, Lizenz, .gitignore
- Konzept-Dokument `docs/CCVIEW-001-konzept.md`
- CCVIEW-002: `internal/parse` — typsicherer JSONL-Event-Parser (10 Tests)
- CCVIEW-002: `internal/tail` — Polling-Tailer mit Partial-Line-Handling (6 Tests, race-free)
- CCVIEW-002: `internal/session` — Projekt-Auflösung, List, Resolve (id/prefix/latest) (9 Tests)
- CCVIEW-003: `internal/srv` — HTTP-Server, SSE-Hub, embedded Frontend (3 Tests)
- CCVIEW-003: `internal/srv/static/index.html` — Vanilla-JS-Viewer mit Dark-Theme
- CCVIEW-004: `cmd/ccview` — CLI mit `-s/--session`, `--port`, `--bind`, `--no-browser`
- CCVIEW-004: Port-Fallback 12100..12199, Cross-Platform Browser-Open
- CCVIEW-005: Frontend-Polish — Minimal-Markdown (Code-Fences/Bold/Italic/Inline-Code)
- CCVIEW-005: Tool-Input-Prettifier für Bash/Read/Edit/Write/Grep/Glob
- CCVIEW-005: Timestamp-Formatierung, Empty-State, Fehler-Badge auf Tool-Results
- CCVIEW-006: Makefile — build / test / vet / race / cross / clean
- CCVIEW-006: `--version` Flag, Versionsinjektion via `-ldflags -X main.version`
- CCVIEW-006: Cross-Compile für linux-amd64/arm64, darwin-amd64/arm64, windows-amd64
- CCVIEW-007: Theme-Switcher Dark / Light / Sepia, Auswahl in localStorage persistiert
- CCVIEW-008: Prompt-Nummerierung #0001+, scrollbares Seitenpanel mit 20-Zeichen-Preview als Sprungmarke
- CCVIEW-009: Seitenpanel ein-/ausklappbar, Preview mit `...` Suffix wenn gekürzt
- CCVIEW-010: `latest` sortiert nach letztem Event-Timestamp in der JSONL statt mtime — robust gegen Claude-Code-Cross-Writes
- CCVIEW-011: Sidepanel-Tabs `Prompts` / `Sessions` — Sessions-Tab listet alle Sessions im Projekt, hover = erster Prompt als Tooltip, aktive Session markiert
- CCVIEW-011: `GET /api/sessions` Endpoint — liefert pro Session `short_id`, `last_event`, `size`, `first_prompt`, `current`
- CCVIEW-011: `session.ReadFirstUserPrompt` — liest ersten User-Prompt aus den ersten 64 KB einer JSONL
- CCVIEW-012: Image-Block-Support im Parser (`base64` + `url`), `<img>` im Frontend
- CCVIEW-013: `session.ListAll` — scannt alle Projekte unter `~/.claude/projects/*`
- CCVIEW-013: Sort nach `FirstEventTime` (Session-Start aus JSONL) — stabil gegen Cross-Writes
- CCVIEW-013: Badge "aktuell" → "offen"; Sessions-Tab zeigt Projekt-Label pro Eintrag
- CCVIEW-013: Custom Hover-Popup für Sidepanel-Items mit vollem Prompt-Text (bis 600 Zeichen) und Kopier-Command
- CCVIEW-014: Session-Switch zur Laufzeit — Klick auf Session im Sidepanel lädt diese in Live-Tail
- CCVIEW-014: `POST /api/switch` Endpoint, `Hub.Reset`, Server übernimmt Tailer-Verwaltung vom `main.go`
- CCVIEW-014: Rahmen für Sessions mit Event am aktuellen Tag (`.today` Klasse)
- CCVIEW-015: Echte User-Prompts visuell hervorgehoben (lila Left-Border, bold Label), Tool-Result-User als "tool-result" umgelabelt und gedämpft; leere Text-/Thinking-Blöcke gefiltert
- CCVIEW-016: Favoriten-Leiste oben — bis zu 5 pinnbare Sessions (★ im Sessions-Tab), hellgrüner Hintergrund wenn seit letztem Öffnen neue Events; 15-s-Poll mit Visibility-Check; Klick schaltet + markiert als gesehen
