---
name: mdc Phase 1 実装
overview: "Go言語のマルチリポジトリ管理CLIツール「mdc」のプロジェクト初期化と Phase 1（MVP: mdc up / mdc down の基盤）を実装する。"
todos:
  - id: init-module
    content: go mod init mdc + cobra / yaml.v3 の依存追加
    status: completed
  - id: config-package
    content: "internal/config/config.go: YAML構造体定義 & Load関数の実装"
    status: completed
  - id: logger-package
    content: "internal/logger/logger.go: プレフィックス+絵文字ログ出力の実装"
    status: completed
  - id: runner-package
    content: "internal/runner/runner.go: os/exec実行基盤 + parallel/sequential制御の実装"
    status: completed
  - id: cmd-root
    content: "cmd/root.go: cobraルートコマンドの定義"
    status: completed
  - id: cmd-up-down
    content: "cmd/up.go, cmd/down.go: up/downコマンドの実装"
    status: completed
  - id: main-entry
    content: "main.go: エントリポイント作成"
    status: completed
  - id: build-verify
    content: go build による動作確認
    status: completed
isProject: false
---

# mdc Phase 1: コア機能の開通（MVP）実装計画

## ディレクトリ構成

Go CLIのベストプラクティス（`cmd/` にコマンド定義、`internal/` にビジネスロジック）に従い、以下の構成とする。

```
mdc/
├── main.go                    # エントリポイント（cmd.Execute() を呼ぶだけ）
├── go.mod
├── go.sum
├── cmd/
│   ├── root.go                # ルートコマンド定義
│   ├── up.go                  # mdc up コマンド
│   └── down.go                # mdc down コマンド
├── internal/
│   ├── config/
│   │   └── config.go          # YAML読み込み & 構造体定義
│   ├── runner/
│   │   └── runner.go          # os/exec によるコマンド実行、parallel/sequential 制御
│   └── logger/
│       └── logger.go          # プレフィックス＋絵文字ログ出力
└── docs/
    └── OVERVIEW.md            # 既存の要件定義書
```

- `internal/` を使うことで、パッケージが外部から import されることを防ぐ（Go の慣習）。
- `cmd/` には cobra のコマンド定義のみ配置し、ロジックは `internal/` に分離する。

## 依存ライブラリ

- `github.com/spf13/cobra` — CLI フレームワーク
- `gopkg.in/yaml.v3` — YAML パーサー

## 実装ファイル詳細

### 1. `main.go`（エントリポイント）

cobra の慣習に従い、`cmd.Execute()` を呼ぶだけのシンプルなエントリポイント。

### 2. `internal/config/config.go`（YAML 読み込み & 構造体）

Phase 2 を見越して `commands.up` / `commands.down` のネスト構造を最初から定義する。

```go
type Config struct {
    ExecutionMode string    `yaml:"execution_mode"`
    Projects      []Project `yaml:"projects"`
}

type Project struct {
    Name     string   `yaml:"name"`
    Path     string   `yaml:"path"`
    Commands Commands `yaml:"commands"`
}

type Commands struct {
    Up   []string `yaml:"up"`
    Down []string `yaml:"down"`
}
```

主要な関数:

- `ExpandHome(path string) (string, error)` — `~` で始まるパスを `os.UserHomeDir()` で絶対パスに展開するユーティリティ。Go は `~` を自動展開しないため、設定ディレクトリのパスとプロジェクトの `path` の両方でこの関数を使う。
- `DefaultConfigDir() string` — `os.UserHomeDir()` を使い `$HOME/.config/mdc/` の絶対パスを返す
- `Load(name string) (*Config, error)` — 設定名から YAML を読み込む（`.yml` 拡張子の自動補完付き）。読み込み後、各 `Project.Path` に対しても `ExpandHome()` を適用し `~/...` を解決する。
- バリデーション: `execution_mode` が `"parallel"` / `"sequential"` のいずれかであること、`projects` が空でないこと等

### 3. `internal/runner/runner.go`（コマンド実行基盤）

- `Run(cfg *config.Config, action string) error` — `action` は `"up"` or `"down"`
- `execution_mode` に応じて並列/直列を分岐:
  - **sequential**: `for` ループでプロジェクトを順番に処理
  - **parallel**: `sync.WaitGroup` + goroutine で全プロジェクトを同時処理
- 1プロジェクト内の複数コマンドは常に直列実行（`commands.up` の配列順）
- `os/exec.Command` に `Dir` フィールドを設定してディレクトリを変更
- コマンド文字列は `sh -c "..."` 経由で実行（パイプやシェル構文対応のため）
- 並列実行時のエラー収集: goroutine ごとにエラーをチャネルまたはスライスで回収し、最終的にまとめて報告

**標準出力・標準エラー出力の制御:**

- **sequential モード**: `cmd.Stdout` / `cmd.Stderr` を `os.Stdout` / `os.Stderr` に直接接続し、コマンドの出力をそのままターミナルにリアルタイム表示する。
- **parallel モード**: 出力が混ざって UI が崩れるのを防ぐため、以下の方式を採用する。
  1. 各コマンドの `cmd.Stdout` / `cmd.Stderr` を `bytes.Buffer` にリダイレクトしてバッファリングする。
  2. コマンドが**成功した場合**: バッファの中身は破棄し、ターミナルには logger の成功メッセージのみを出力する。
  3. コマンドが**失敗した場合**: logger のエラーメッセージに続けて、バッファに溜まった stderr（および stdout）の内容を出力し、何が起きたかを確認できるようにする。

### 4. `internal/logger/logger.go`（ログ出力）

要件定義書の UI/UX 仕様に対応:

- `Start(projectName, cmd string)` — `🚀 [Frontend] Executing: make up`
- `Success(projectName, cmd string)` — `✅ [Frontend] Completed: make up`
- `Error(projectName, cmd string, err error)` — `❌ [Frontend] Failed: make up — exit status 1`
- `Output(projectName, output string)` — エラー時にバッファされた出力を表示する
- 並列実行時は `sync.Mutex` で標準出力の競合を防ぐ（logger 経由の出力は全てロックで排他制御）

### 5. `cmd/root.go`（ルートコマンド）

- `mdc` のルートコマンドを定義（ヘルプテキスト、バージョン情報など）
- `Execute()` 関数を公開

### 6. `cmd/up.go` / `cmd/down.go`

- 引数 `[config-name]` を cobra の `Args` で受け取る
- `config.Load()` → `runner.Run()` の流れでコア処理を呼び出す
- エラー時は適切なメッセージを出力して非ゼロの終了コードで終了

## エラーハンドリング方針

- パスが存在しない場合: `runner` で実行前に `os.Stat` でチェックし、存在しなければプロジェクト名付きのエラーメッセージを出力
- コマンド失敗時: どのプロジェクトのどのコマンドで失敗したかを明示して処理を中断
- sequential モード: エラー発生時点で後続プロジェクトの処理を停止
- parallel モード: 各 goroutine は自プロジェクト内でエラー発生時に停止し、全 goroutine 完了後にエラーをまとめて報告

