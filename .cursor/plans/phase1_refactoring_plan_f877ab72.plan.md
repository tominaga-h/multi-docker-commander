---
name: Phase1 Refactoring Plan
overview: フェーズ1のコードを、単一責務原則（SRP）とDRY原則を中心に6つの観点からリファクタリングする。既存テストを壊さず、コードの保守性・可読性を向上させる。
todos:
  - id: r1-cmd-helper
    content: "R1: cmd パッケージに loadAndRun ヘルパーを追加し、up.go/down.go の重複を解消"
    status: completed
  - id: r2-base-dir
    content: "R2: ベースディレクトリ解決を config.BaseMDCDir() に共通化し、pidfile から参照"
    status: completed
  - id: r3-exec-split
    content: "R3: execCommand を execBackgroundCommand / execForegroundPTY / execForegroundStd に分割"
    status: completed
  - id: r4-cleanup
    content: "R4: バックグラウンドプロセスのクリーンアップロジックを pidfile パッケージに移動"
    status: completed
  - id: r5-parallel-slice
    content: "R5: commandsForAction の並列スライスを ProjectCommands 構造体ベースに変更"
    status: completed
  - id: r6-logger-helper
    content: "R6: Logger の共通書き込みヘルパー writef を抽出"
    status: completed
isProject: false
---

# Phase 1 リファクタリングプラン

## 現状の課題分析

### DRY違反

**1. `cmd/up.go` と `cmd/down.go` のコマンドハンドラ重複**

両ファイルが全く同じ「設定読み込み -> runner実行 -> エラー処理」パターンを繰り返している。

`cmd/up.go` (17-28行目):

```go
cfg, err := config.Load(configName)
if err != nil {
    fmt.Fprintln(os.Stderr, err)
    os.Exit(1)
}
if err := runner.Run(cfg, "up", configName); err != nil {
    fmt.Fprintln(os.Stderr, err)
    os.Exit(1)
}
```

`cmd/down.go` (20-28行目) もほぼ同一。

**2. ベースディレクトリ解決の重複**

`[internal/config/config.go](internal/config/config.go)` の `DefaultConfigDir()` と `[internal/pidfile/pidfile.go](internal/pidfile/pidfile.go)` の `baseDir()` が、どちらも `os.UserHomeDir()` + `~/.config/mdc/` というパスを独立に構築している。

```go
// config.go:58-64
func DefaultConfigDir() (string, error) {
    home, err := os.UserHomeDir()
    // ...
    return filepath.Join(home, ".config", "mdc"), nil
}

// pidfile.go:19-28
func baseDir() (string, error) {
    // ...
    home, err := os.UserHomeDir()
    // ...
    return filepath.Join(home, ".config", "mdc", "pids"), nil
}
```

**3. Logger の boilerplate 繰り返し**

`[internal/logger/logger.go](internal/logger/logger.go)` の公開関数（`Start`, `Success`, `Error`, `Background`, `Stop`, `ProjectDone`, `ProjectFailed`）がすべて同じ `mu.Lock(); defer mu.Unlock(); fmt.Fprintf(out, ...)` パターンを持つ。ヘルパー関数で共通化できる。

### SRP違反

**4. `execCommand` 関数の肥大化（72行）**

`[internal/runner/runner.go](internal/runner/runner.go)` の `execCommand`（107-178行目）が以下の複数責務を抱えている:

- バックグラウンドコマンドの起動とPIDファイル管理
- PTY検出と PTY 経由の実行
- バッファード/非バッファードの分岐
- ログ出力（Start/Success/Error/Border）

3つの実行パスが深いネストで分岐しており、可読性が低い。

**5. `cmd/down.go` にビジネスロジックが混在**

`[cmd/down.go](cmd/down.go)` の Run ハンドラ（30-39行目）で、バックグラウンドプロセスのクリーンアップ（PID読み込み -> ログ出力 -> Kill）を直接行っている。CLIハンドラがインフラ層の詳細に依存している。

### 暗黙的な結合

**6. `commandsForAction` の並列スライス**

`[internal/runner/runner.go](internal/runner/runner.go)` の `commandsForAction` が `[][]config.CommandItem` を返し、`projects[i]` と `commands[i]` がインデックスで暗黙的に対応している。これは読み手に認知負荷をかけ、バグの温床になり得る。

---

## リファクタリング項目

### R1: cmd パッケージのヘルパー関数抽出（DRY）

`cmd/root.go` に共通ヘルパーを追加し、`up.go`/`down.go` の重複を解消する。

- `cmd/root.go` に `loadAndRun(configName, action string)` 関数を追加
- `up.go` と `down.go` から共通パターンを置き換え
- 対象ファイル: `[cmd/root.go](cmd/root.go)`, `[cmd/up.go](cmd/up.go)`, `[cmd/down.go](cmd/down.go)`

### R2: ベースディレクトリ解決の共通化（DRY）

`internal/config/config.go` に `BaseMDCDir()` を定義し、`pidfile` パッケージからも参照する。

- `config.BaseMDCDir()` を新設（`~/.config/mdc` を返す）
- `config.DefaultConfigDir()` は `BaseMDCDir()` を使うように変更
- `pidfile.baseDir()` は `config.BaseMDCDir()` + `/pids` を使うように変更
- 対象ファイル: `[internal/config/config.go](internal/config/config.go)`, `[internal/pidfile/pidfile.go](internal/pidfile/pidfile.go)`

### R3: `execCommand` の分割（SRP）

72行の `execCommand` を、実行パスごとに分割する。

- `execBackgroundCommand(p, item, configName)` を抽出: バックグラウンド起動 + PID保存
- `execForegroundPTY(p, item, buffered)` を抽出: PTY経由のフォアグラウンド実行
- `execForegroundStd(p, item, buffered)` を抽出: 標準のフォアグラウンド実行
- `execCommand` はロギングとディスパッチだけの薄いファサードにする
- 対象ファイル: `[internal/runner/runner.go](internal/runner/runner.go)`

### R4: バックグラウンドプロセスのクリーンアップをパッケージに移動（SRP）

`cmd/down.go` のクリーンアップロジックを `pidfile` パッケージの関数に移動する。

- `pidfile.CleanupWithLog(configName string)` を新設: ログ出力 + KillAll をまとめる
- ただし `logger` への依存を避けるため、コールバック関数を受け取る設計にする: `pidfile.KillAllWithCallback(configName string, onStop func(projectName, command string, pid int)) error`
- `cmd/down.go` は1行で呼び出すだけにする
- 対象ファイル: `[internal/pidfile/pidfile.go](internal/pidfile/pidfile.go)`, `[cmd/down.go](cmd/down.go)`

### R5: 並列スライスの解消（可読性・安全性）

`commandsForAction` の返り値を構造体ベースに変更する。

- `type ProjectCommands struct { Project config.Project; Commands []config.CommandItem }` を定義
- `commandsForAction` が `[]ProjectCommands` を返すようにする
- `runSequential`/`runParallel`/`runProjectBuffered` をそれに合わせて調整
- 対象ファイル: `[internal/runner/runner.go](internal/runner/runner.go)`

### R6: Logger の共通書き込みヘルパー抽出（DRY）

`logger.go` のすべての公開関数が同じ mu.Lock/Unlock + fmt.Fprintf パターンを持つのを共通化する。

- `writef(format string, args ...any)` プライベート関数を追加（Lock/Unlock + Fprintf をまとめる）
- 各公開関数（`Start`, `Success`, `Error` 等）を `writef` 呼び出しに置き換え
- 対象ファイル: `[internal/logger/logger.go](internal/logger/logger.go)`

---

## 実施方針

- 各項目を独立したステップで実施し、ステップごとに `make build` と既存テスト通過を確認する
- テストコードは原則そのまま維持（内部関数名の変更に伴うテスト修正のみ行う）
- 外部公開APIの変更は最小限にとどめる

