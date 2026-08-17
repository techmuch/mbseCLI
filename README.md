# mbsecli

> **Website & Live Demo:** [https://techmuch.github.io/mbseCLI/](https://techmuch.github.io/mbseCLI/)

A CLI + local web server for live visualization and review of SysML v2
textual models. Point it at a `.sysml` file (or a directory of them) and it
watches the file, parses it, and serves an interactive UI — object tree,
dockable diagram views, element inspector — that updates as the model
changes on disk.

Same shape as mbseCLI's sibling tools: a single Go binary with an embedded
React frontend, no separate server process or install step for end users.

## Status

Early scaffold. The parser is a deliberately naive, line-oriented tokenizer
covering common declaration forms (`package`, `part def`, `part`, `port`,
`action`, `state`, `requirement`, `attribute`) — enough to drive the tree,
diagram, and live-reload pipeline end to end. It is **not** a conformant
SysML v2 / KerML parser. See `internal/parser/parser.go` for the intended
seam to swap in a generated ANTLR4 parser later.

## Quick start

```sh
make build
./mbsecli start --open examples/drone.sysml
```

Then edit `examples/drone.sysml` in your editor and watch the tree and diagram update.

## Development

```sh
# Terminal 1: Go server in --dev mode (serves /api and /ws only)
make dev

# Terminal 2: Vite dev server with hot reload, proxying /api and /ws to :4173
cd web && npm run dev
```

Open http://localhost:5173 during development.

## Architecture

```
.sysml file on disk
  -> internal/watch   (fsnotify, debounced)
  -> internal/parser  (naive tokenizer -> internal/model.Graph)
  -> internal/server  (WebSocket broadcast + REST API)
  -> web/ (React + nexus-shell)
       left pane:   object tree (TreeWidget)
       center pane: dockable diagram views (GraphCanvas, registered per view type)
       right pane:  element inspector + feedback notes (PropertyPanel + custom notes UI)
```

Feedback notes are anchored to an element's fully-qualified name (FQN) and
persisted to a JSON sidecar file next to the model
(`.<model_name>.feedback.json`) — plain text, diffable, travels with the
model in version control. Notes whose FQN no longer exists after a reparse
(element renamed/deleted) are flagged as orphaned rather than dropped.

## Testing

```sh
make test    # Go unit tests: internal/parser, internal/feedback
make e2e     # Playwright, against a real running mbsecli instance
```

The e2e suite (`web/e2e/`) boots `mbsecli` against a scratch copy of
`examples/drone.sysml` (see `web/e2e/start-server.sh`) so it's free to edit
the file on disk — that's what the live-reload test exercises. First-time
setup needs the browser binary once:

```sh
cd web && npx playwright install chromium
```

## CLI

```sh
mbsecli start [file-or-directory]   # default: current directory
  --port 4173        # web UI / API port
  --debounce 200ms    # file-change debounce window
  --open, -o          # open browser automatically once started
  --dev               # skip embedded assets; pair with `npm run dev`
mbsecli version
```

## Next steps

- Swap the naive parser for a generated ANTLR4 parser from the OMG
  `SysML.g4` / `KerML.g4` grammars (`antlr4 -Dlanguage=Go`), preserving
  `internal/model.Graph` as the output shape.
- Additional view types on the same `Graph` IR: interconnection/IBD (ports +
  `connect`), behavioral (`action`/`state` flow), requirement traceability
  (`satisfy`/`verify` matrix).
- "Export notes to model" — write resolved feedback back into the `.sysml`
  source as `doc /* ... */` annotations.
