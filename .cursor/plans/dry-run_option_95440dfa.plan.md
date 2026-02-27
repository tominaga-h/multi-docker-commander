---
name: dry-run option
overview: "`--dry-run` フラグを `up` / `down` コマンドに追加し、コマンドを実行せず実行計画（プロジェクト名・パス・コマンド順序）を表示する機能を実装する。"
todos:
  - id: flag-up-down
    content: cmd/up.go と cmd/down.go に --dry-run フラグを追加し、cmd/root.go の loadAndRun を更新
    status: completed
  - id: logger-dryrun
    content: internal/logger/logger.go に dry-run 用の出力関数を追加
    status: completed
  - id: runner-dryrun
    content: internal/runner/runner.go に DryRun 関数を実装
    status: completed
  - id: tests
    content: runner_test.go と logger_test.go に dry-run のテストを追加
    status: completed
  - id: make-check
    content: make check を実行してリント・テストを通すことを確認
    status: completed
isProject: false
---

# `--dry-run` オプションの実装

## 変更概要

`mdc up --dry-run <config>` および `mdc down --dry-run <config>` で、実際のコマンド実行をスキップし、実行計画のみを標準出力に表示する。

## 変更ファイルと内容

### 1. [cmd/root.go](cmd/root.go) -- `loadAndRun` に dryRun パラメータ追加

- `loadAndRun(configName, action string)` を `loadAndRun(configName, action string, dryRun bool)` に変更
- dryRun が true の場合 `runner.DryRun(cfg, action)` を呼び出し、false の場合は従来通り `runner.Run(...)` を実行

### 2. [cmd/up.go](cmd/up.go) -- `--dry-run` フラグ登録

- パッケージレベル変数 `upDryRun` を追加
- `init()` で `upCmd.Flags().BoolVar(&upDryRun, "dry-run", false, ...)` を登録
- `Run` 内で `loadAndRun(args[0], "up", upDryRun)` を呼び出す

### 3. [cmd/down.go](cmd/down.go) -- `--dry-run` フラグ登録

- パッケージレベル変数 `downDryRun` を追加
- `init()` で `downCmd.Flags().BoolVar(&downDryRun, "dry-run", false, ...)` を登録
- `Run` 内で `loadAndRun(configName, "down", downDryRun)` を呼び出す
- dryRun 時は `pidfile.KillAllWithCallback` をスキップ

### 4. [internal/runner/runner.go](internal/runner/runner.go) -- `DryRun` 関数の追加

新しいエクスポート関数 `DryRun(cfg *config.Config, action string) error` を追加:

- `commandsForAction()` を再利用してプロジェクトとコマンドを取得
- `logger.DryRun*` 系の関数を使って実行計画を出力
- パス存在チェックも行い、存在しないパスには警告を表示

### 5. [internal/logger/logger.go](internal/logger/logger.go) -- dry-run 用の出力関数追加

以下のログ関数を追加:

```go
func DryRunHeader(action, mode string)
func DryRunProject(projectName, path string, commands []string, backgrounds []bool)
```

出力イメージ:

```
📋 Dry-run: up (mode: sequential)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

[Frontend]
  📂 /path/to/frontend-repo
    1. make up
    2. make run

[Backend-API]
  📂 /path/to/backend-api-repo
    1. docker-compose up -d
    2. sleep 60 [background]
```

### 6. テストの追加

- [internal/runner/runner_test.go](internal/runner/runner_test.go): `TestDryRun` -- DryRun 関数が正しいフォーマットで出力されること、コマンドが実行されないことを検証
- [internal/logger/logger_test.go](internal/logger/logger_test.go): dry-run 出力関数のテスト

## 設計判断

- `--dry-run` は `up` / `down` のみに適用（`list` 等には不要）
- `runner.Run` のシグネチャは変更せず、別関数 `DryRun` として分離 -- 既存コードへの影響を最小化
- dry-run 時もパス検証を行い、存在しないパスには警告を出す

