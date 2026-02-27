---
name: proc stop/restart commands
overview: 既存の `mdc procs` を `mdc proc` にリネームし、`proc list` / `proc stop <PID>` / `proc restart <PID>` のサブコマンド構造に再構成する。グレースフルシャットダウン（SIGTERM → 10秒待機 → SIGKILL）も実装する。
todos:
  - id: pidfile-find-remove
    content: pidfile パッケージに FindByPID / RemoveEntry を実装
    status: completed
  - id: graceful-kill
    content: process_unix.go / process_windows.go に GracefulKill を実装し、KillAllWithCallback も更新
    status: completed
  - id: runner-start-bg
    content: runner パッケージに StartBackgroundProcess を抽出・実装
    status: completed
  - id: cmd-proc-restructure
    content: procs.go を proc.go にリネームし、proc 親コマンド + list サブコマンドに再構成
    status: completed
  - id: cmd-proc-stop
    content: cmd/proc_stop.go に mdc proc stop <PID> を実装
    status: completed
  - id: cmd-proc-restart
    content: cmd/proc_restart.go に mdc proc restart <PID> を実装
    status: completed
  - id: logger-update
    content: logger に Restart 等の必要なログ関数を追加
    status: completed
  - id: build-test
    content: make build && make test-all で動作確認
    status: completed
isProject: false
---

# mdc proc stop / restart コマンドの実装

## 現状の課題

- `mdc down` は config 内の全バックグラウンドプロセスを SIGKILL で即座に強制終了する
- 個別プロセスの停止・再起動手段がない
- グレースフルシャットダウンが未実装

## 実装方針

### 1. pidfile パッケージの拡張

**[internal/pidfile/pidfile.go](internal/pidfile/pidfile.go)** に以下を追加:

- `FindByPID(pid int) (configName, projectName string, entry Entry, err error)` — 全 PID ファイルを走査し、指定 PID のエントリを特定
- `RemoveEntry(configName, projectName string, pid int) error` — 特定エントリを PID ファイルから削除（エントリが空になったらファイルも削除）

### 2. グレースフルシャットダウンの実装

**[internal/pidfile/process_unix.go](internal/pidfile/process_unix.go)** に追加:

- `GracefulKill(pid int, timeout time.Duration) error` — SIGTERM 送信 → timeout 待機 → まだ生きていれば SIGKILL

`**internal/pidfile/process_windows.go`** にも同名関数を追加（Windows は SIGTERM 非対応のため `Kill()` で代替）

`**KillAllWithCallback`** も `GracefulKill` を使うよう更新し、`mdc down` もグレースフル化

### 3. runner パッケージの拡張

**[internal/runner/runner.go](internal/runner/runner.go)** に追加:

- `StartBackgroundProcess(command, dir string) (int, error)` — `execBackgroundCommand` からプロセス起動ロジックを抽出。restart 時に再利用

### 4. コマンド構造の再編成

**[cmd/procs.go](cmd/procs.go)** を `**cmd/proc.go`** にリネームし、`proc` を親コマンドに変更:

- `mdc proc` — サブコマンドなしで実行した場合は `proc list` と同等の動作
- `mdc proc list [config-name]` — 既存の procs 機能をそのまま移行

新規ファイル:

- `**cmd/proc_stop.go`** — `mdc proc stop <PID>`
  - `FindByPID` で PID エントリを特定
  - `GracefulKill` でプロセスを停止
  - `RemoveEntry` で PID ファイルから削除
  - `logger.Stop` でログ出力
- `**cmd/proc_restart.go`** — `mdc proc restart <PID>`
  - `FindByPID` で PID エントリを特定（command と dir を取得）
  - `GracefulKill` でプロセスを停止
  - `runner.StartBackgroundProcess` で同じコマンドを再実行
  - `RemoveEntry` で旧エントリを削除し、`pidfile.Append` で新エントリを追加
  - `logger.Stop` + `logger.Background` でログ出力

### 5. ログ関数の追加

**[internal/logger/logger.go](internal/logger/logger.go)** に必要に応じてログ関数を追加（`Restart` など）

## コマンド使用例

```
$ mdc proc list myproject
CONFIG     PROJECT    COMMAND             DIR         PID    STATUS
myproject  Frontend   npm run dev         ~/frontend  12345  Running

$ mdc proc stop 12345
🛑 [Frontend] Stopping: npm run dev (PID: 12345)
✅ [Frontend] Stopped successfully

$ mdc proc restart 12345
🛑 [Frontend] Stopping: npm run dev (PID: 12345)
🔄 [Frontend] Background: npm run dev (PID: 67890)
```

