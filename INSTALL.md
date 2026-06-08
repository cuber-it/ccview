# Installing ccview

ccview is a single Go binary that serves a local web UI for browsing your
Claude Code sessions. To run it you need:

- **read** access to your Claude Code sessions — `~/.claude/projects/`
- a **writable** spot for ccview's own small SQLite database
  (config, project roots, session names, notes, done flags, trash) —
  `~/.claude/ccview/ccview.db`

ccview has **no authentication**. Always bind it to `127.0.0.1` (or put it
behind your own reverse proxy / VPN). Default port is `12100`.

Two ways to run it as a persistent service: **systemd** (recommended, simplest)
or **Docker**. systemd is the natural fit because ccview wants direct access to
your home directory; Docker works too but needs explicit volume mounts.

---

## Build from source

Requires Go (see `go.mod` for the version). The version string is injected at
build time:

```sh
git clone https://github.com/cuber-it/ccview.git
cd ccview
go build -ldflags "-X main.version=$(git describe --tags --always)" -o ccview ./cmd/ccview
```

Quick check (no install, no browser): `./ccview --no-browser` then open the
printed URL. The SQLite driver is pure Go, so `CGO_ENABLED=0` builds a fully
static binary that runs anywhere.

---

## A. systemd user service (recommended)

Runs ccview as your user with direct filesystem access — no mounts, no
permission juggling.

**1. Install the binary**

```sh
go build -ldflags "-X main.version=$(git describe --tags --always)" \
  -o ~/.local/bin/ccview ./cmd/ccview
```

**2. Create the unit** — `~/.config/systemd/user/ccview.service`:

```ini
[Unit]
Description=ccview — Claude Code session viewer
After=network.target

[Service]
Type=simple
WorkingDirectory=%h
ExecStart=%h/.local/bin/ccview --no-browser --bind 127.0.0.1 --port 12100
Restart=on-failure
RestartSec=3

[Install]
WantedBy=default.target
```

**3. Enable, start, and keep it running without a login session**

```sh
systemctl --user daemon-reload
systemctl --user enable --now ccview.service
loginctl enable-linger "$USER"          # so it survives logout / runs after boot
```

**4. Verify**

```sh
systemctl --user status ccview.service
curl -s http://127.0.0.1:12100/api/version
```

**Updating:** rebuild the binary into `~/.local/bin/ccview` and
`systemctl --user restart ccview`. (If the binary is in use you'll get
`ETXTBSY` only when *copying* over it — `go build -o` does an atomic rename, so
it's fine; when copying from elsewhere, write `ccview.new` and `mv -f`.)

---

## B. Docker / docker-compose

Containerised, ccview lives *off* your host's `~/.claude`, so you mount it in.
Three things matter — all handled by the provided `Dockerfile` and
`docker-compose.yml`:

1. **Sessions are mounted read-only** — ccview never writes to them.
2. **The database is a host bind-mount, not a named volume.** A named volume is
   created `root:root`; the container runs as your user and couldn't write it
   (`SQLITE_CANTOPEN`). A host-owned bind-mount avoids that.
3. **The container runs as your host UID/GID** so the read-only sessions are
   readable and the bind-mounted DB stays host-owned.

### docker-compose (recommended)

```sh
mkdir -p ~/.claude/ccview            # host-owned DB dir (gotcha #2)
export UID=$(id -u) GID=$(id -g)     # gotcha #3
docker compose up -d --build
```

ccview is then on `http://127.0.0.1:12100`. Stop with `docker compose down`.

### Plain docker run

```sh
docker build -t ccview .
mkdir -p ~/.claude/ccview
docker run -d --name ccview \
  -p 127.0.0.1:12100:12100 \
  -e CLAUDE_CONFIG_DIR=/data \
  -v "$HOME/.claude/projects:/data/projects:ro" \
  -v "$HOME/.claude/ccview:/data/ccview" \
  --user "$(id -u):$(id -g)" \
  --restart unless-stopped \
  ccview
```

> **Don't run the systemd service and the container at the same time** — they'd
> share the same `~/.claude/ccview/ccview.db`. Pick one.

---

## Configuration

| Flag / env | Effect |
|---|---|
| `--port N` | Listen port (default `12100`, auto-increments on collision for the CLI) |
| `--bind 127.0.0.1` | Bind address. Keep it on loopback — there is no auth. |
| `--no-browser` | Don't open a browser (always use this for a service) |
| `--projects-root DIR` | Override the sessions directory |
| `CLAUDE_CONFIG_DIR` | Relocates the whole Claude tree; ccview reads `$CLAUDE_CONFIG_DIR/projects` and stores its DB under `$CLAUDE_CONFIG_DIR/ccview` |

### Multiple project roots

If your Claude sessions live under more than one directory, add them in
**Settings → Projekt-Pfade** (stored in the DB). Under Docker, each root must
also be **mounted into the container** at the path you enter in Settings — e.g.
mount `~/other-claude/projects:/data/projects2:ro` and add `/data/projects2`.

---

## Where ccview stores things

- **Sessions** (read): `~/.claude/projects/<encoded-cwd>/<id>.jsonl` — owned by
  Claude Code, ccview only reads them.
- **Database**: `~/.claude/ccview/ccview.db` — config, project roots, session
  names, favorites, done flags, notes.
- **Trash**: `~/.claude/ccview/trash/` — deleted session JSONLs land here
  (delete is reversible: move the file back).
- **Browser-local**: theme, language, favorites bar, active-session order — kept
  in `localStorage`, per browser.
