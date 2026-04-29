# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

`mdc` (multi-docker-commander) — a Go CLI that orchestrates `docker compose` (and arbitrary commands like `npm run dev`) across multiple repositories from a single YAML config. It also daemonizes foreground commands and tracks them with PID files so they can be listed, attached, stopped, restarted, or killed later.

Read `docs/OVERVIEW.md` first for the original requirements/design (Japanese). `docs/README_ja.md` is the Japanese README; `README.md` is the English one.

## Commands

```bash
make build           # go build -o mdc .
make build-v         # build with version embedded from `git describe`
make build-local     # check + build-v + ./mdc -v (full local validation)

make test            # go test ./internal/... -v   (unit tests only)
make test-integration # go test ./test/... -v      (integration tests in test/integration/)
make test-all        # go test ./... -v
make test-cover      # test-all + coverage report
make lint            # go vet ./... + golangci-lint run
make check           # lint + test-all  ← run this before committing per .cursor/rules
```

Run a single test: `go test ./internal/runner/ -run TestRunSequential -v`

`make install-hooks` installs `githooks/pre-push.sh` as the local pre-push hook.

## Architecture

Entry point is `main.go` → `cmd.Execute()` (Cobra root in `cmd/root.go`). Each top-level subcommand lives in its own file under `cmd/` (`up.go`, `down.go`, `proc.go`, `proc_kill.go`, `ps.go`, `init.go`, `edit.go`, `rm.go`, `list.go`). `loadAndRun` in `cmd/root.go` is the shared entry for `up`/`down`: it loads the config and dispatches to `runner.Run` (or `runner.DryRun`).

Configs live in `~/.config/mdc/<name>.yml`. `internal/config` parses them into `Config → Projects[] → Commands{Up, Down}` where each command is a `CommandItem{Command, Background}`. `CommandItem.UnmarshalYAML` accepts either a plain string (legacy) or an object — preserve this dual form when changing the schema. `~` in `project.path` is expanded at load time via `ExpandHome`. Validation in `Config.validate()` requires `execution_mode ∈ {"parallel", "sequential"}` and at least one project with name+path.

`internal/runner` is the execution core:

- `Run` dispatches to `runSequential` or `runParallel`. In parallel mode, project paths are validated up front, then each project runs in its own goroutine via `runProjectBuffered`; output is buffered per-project so logs don't interleave. In sequential mode, output is streamed live (PTY when stdout is a TTY, else stdio).
- Within a single project, commands always run **sequentially** in declaration order — this invariant is part of the spec (see `docs/OVERVIEW.md` §6).
- `expandProcKill` rewrites a literal `mdc proc kill` token in `commands.down` to `mdc proc kill -c <configName>` so downs can self-clean their tracked background processes.
- Background commands are launched detached via `StartBackgroundProcess`. When a log file is requested, the command is wrapped in `script(1)` (`internal/runner/script_unix.go` / `_windows.go`) so the child runs under a PTY, preserving ANSI colors in the log. PID, command, and dir are then appended to the config's pidfile.
- PTY support and platform-specific process attributes are split into `*_unix.go` / `*_windows.go` build-tagged files (`process_*`, `procattr_*`, `pty_*`, `script_*`). Keep both implementations in sync when touching these.

`internal/pidfile` owns persistent process state under `~/.config/mdc/`:

- `~/.config/mdc/pids/<config>.json` — array of `{pid, command, dir, project}` entries per config.
- `~/.config/mdc/proc/<config>/<project>/<pid>.log` — captured background process output.
- `BaseDir` package-level variable is overrideable for tests (mirrored by `procBaseDir` deriving from it). Tests that touch pidfiles must set/restore `pidfile.BaseDir`.

`internal/ps` resolves the live status of tracked PIDs (Running / Stopped) by probing the OS, used by `mdc proc list` / `mdc procs`.

`internal/logger` centralizes all user-facing output: `Start/Success/Error/Background/ProjectDone/Border/DryRun*/Output`. Emoji conventions (🚀 start, ✅ success, ❌ failure) and `[ProjectName]` prefixes come from spec §7 — keep them when adding new log sites.

`internal/version` holds `Version` injected via `-ldflags "-X mdc/internal/version.Version=..."` from `make build-v`.

## Tests

Unit tests live next to the code in `internal/*/`. Integration tests live in `test/integration/` and exercise the built binary end-to-end (run with `make test-integration`). When tests need a clean config/pid directory, override `pidfile.BaseDir` and the config dir helpers rather than touching the real `~/.config/mdc/`.

## Conventions

- Per `.cursor/rules/general.mdc`: read `docs/OVERVIEW.md` first when starting work; run `make check` (lint + tests) before considering a change done.
- Module path is `mdc` (see `go.mod`). Internal imports use `mdc/internal/...`.
- Go 1.25+. Cobra for CLI, `gopkg.in/yaml.v3` for YAML.
- Branching: `develop` is the working branch, `main` is release.
- **コミットメッセージは日本語で書くこと**。タイトル・本文ともに日本語が原則。型プレフィックス（`fix:` / `feat:` / `ci:` 等）は付けても付けなくてもよい（既存の履歴と揃える）。
