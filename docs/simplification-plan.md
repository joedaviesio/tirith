# Plan: Simplify Tirith to brew-install + one command + browser

## Context

Tirith currently requires building from source, running `tirith start`, then separately running `tirith dashboard` to open the browser. The goal is a dead-simple developer experience:

```
brew install joedaviesio/tap/tirith
tirith
# → proxy starts, dashboard starts, browser opens automatically
```

## Changes

### 1. Make bare `tirith` (no subcommand) start everything + open browser

**File: `cmd/costwatch/main.go`**

- Extract the start logic from `startCmd().RunE` into a standalone function:
  ```go
  func runStart(port int, openBrowser bool) error { ... }
  ```
- Set the root command's `RunE` to call `runStart(0, true)` — no args = start + open browser
- Keep `startCmd` but have its `RunE` call `runStart(port, !noOpen)` with a `--no-open` flag
- Add browser auto-open logic after servers are listening:
  - Wait for dashboard health (poll `localhost:{port}` briefly)
  - Use `runtime.GOOS` switch: `open` (macOS), `xdg-open` (Linux), `cmd /c start` (Windows)
- Keep the process in the foreground (Ctrl+C to stop) — no daemon mode

### 2. Add goreleaser config for cross-platform builds

**New file: `.goreleaser.yaml`**

- Build targets: `darwin/amd64`, `darwin/arm64`, `linux/amd64`, `linux/arm64`
- Main: `./cmd/costwatch`
- Binary name: `tirith`
- Ldflags: `-X main.version={{.Version}}`
- Before hook: build frontend (`cd frontend && npm run build && cp -r out/* ../internal/dashboard/frontend/`)
- Archives: tar.gz (Linux), zip (macOS)
- Homebrew tap: `joedaviesio/homebrew-tap`

### 3. Add GitHub Actions release workflow

**New file: `.github/workflows/release.yaml`**

- Trigger: tag push `v*`
- Steps: checkout → setup Go → setup Node → `npm ci` in frontend → goreleaser release
- goreleaser publishes to GitHub Releases and auto-updates the Homebrew tap formula

### 4. Create Homebrew tap repo (manual step)

- Create `joedaviesio/homebrew-tap` on GitHub
- goreleaser auto-pushes the formula on release
- Users install with: `brew install joedaviesio/tap/tirith`

## Files modified

| File | Action |
|------|--------|
| `cmd/costwatch/main.go` | Extract `runStart()`, add root `RunE` with browser open, cross-platform open, `--no-open` flag |
| `.goreleaser.yaml` | New — cross-compile + Homebrew tap config |
| `.github/workflows/release.yaml` | New — tag-triggered release workflow |

## End-to-end experience after implementation

```bash
# Install
brew install joedaviesio/tap/tirith

# Start (one command, browser opens automatically)
tirith

# Or with options
tirith --no-open          # start without opening browser
tirith start --port 5678  # custom port, no auto-open
tirith report --last 7d   # terminal report still works
tirith run -- python app.py  # wrapper mode still works
```

## Verification

1. `make build` produces a binary named `tirith`
2. `./tirith` starts proxy + dashboard + opens browser
3. `./tirith --no-open` starts without opening browser
4. `./tirith start` still works (backward compat)
5. `goreleaser release --snapshot --clean` builds all platforms
6. After creating tap repo + tagging a release, `brew install joedaviesio/tap/tirith` works
