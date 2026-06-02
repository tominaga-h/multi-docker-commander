# Changelog

このプロジェクトのすべての注目すべき変更はこのファイルに記録されます。

フォーマットは [Keep a Changelog](https://keepachangelog.com/ja/1.1.0/) に準拠し、
バージョニングは [Semantic Versioning](https://semver.org/lang/ja/) に従います。

## [2.0.3] - 2026-06-02

### Fixed

- `mdc down` がコマンド失敗で途中終了しても、追跡中のバックグラウンドプロセスのクリーンアップが必ず実行されるよう修正 (#177)

### Added

- `mdc proc kill --dead` オプションで、既に終了済みのプロセスの PID ファイルエントリのみを一括削除（実行中のプロセスは対象外）(#191)

## [2.0.2] - 2026-04-29

### Added

- リリースワークフローに darwin/amd64 ビルドを追加
- バイナリのダウンロード・インストール手順を README に追記

### Changed

- CLAUDE.md にコミットメッセージは日本語で書く規約を明記

## [2.0.1] - 2026-04-29

### Added

- `mdc proc kill --all` オプションで追跡中の全バックグラウンドプロセスを一括終了
- マルチプラットフォーム対応の自動リリースワークフロー (GitHub Actions) を追加

### Changed

- `mdc proc kill` のドキュメントを README に追記

## [2.0.0] - 2026-03-10

### Added

- `mdc proc kill <PID>` コマンドでバックグラウンドプロセスとその子プロセスツリーを強制終了 (#21)
- README に Homebrew からのインストール手順を追加

## [1.0.0] - 2026-03-05

### Added

- `mdc init <config-name>` コマンドで設定ファイルのテンプレートを新規作成 (#17)
- `mdc edit <config-name>` コマンドで設定ファイルをエディタで直接編集 (#19)
- `mdc rm <config-name>` コマンドで設定ファイルを削除 (#20)
- README を英語化し、日本語版を `docs/README_ja.md` に分離 (#18)
- デモ GIF を README に追加

### Changed

- README 全体の見直しとコマンドリファレンスの充実

## [0.2.1] - 2026-02-28

### Added

- リリース手順を自動化する Cursor Command (`/release`) を追加

### Changed

- プロジェクト名を Multi-Docker-Commander に変更（バクロニム）

## [0.2.0] - 2026-02-28

### Added

- `--dry-run` オプションで実行計画を事前確認可能に (#9)
- `mdc ps [config-name]` による Docker コンテナ状態の一覧表示 (#16)
- `mdc proc attach <PID>` によるバックグラウンドプロセスのログストリーム (#14)
- `mdc proc stop <PID>` / `mdc proc restart <PID>` によるバックグラウンドプロセスの停止・再起動 (#13)
- GitHub Actions による自動テスト CI 環境 (#10)
- pre-push フックによるコミット前チェック (#12)

### Changed

- golangci-lint をツールチェーンビルド方式に変更
- README の全面的な見直し・更新 (#15)

### Fixed

- GitHub Actions の build バッヂが正しく表示されない問題を修正

## [0.1.0] - 2026-02-27

### Added

- `mdc up [config-name]` / `mdc down [config-name]` による複数プロジェクトの一括起動・停止
- プロジェクト間の実行モード制御 (`parallel` / `sequential`)
- `mdc list` (`mdc ls`) による設定ファイル一覧表示 (#7)
- `mdc procs [config-name]` によるバックグラウンドプロセス一覧表示 (#5)
- `mdc --version` / `mdc -v` によるバージョン表示 (#6)
- プロジェクト名プレフィックスと絵文字によるリッチなログ出力 (#4)
- コマンド単位のバックグラウンド実行 (`background: true`) と PID ファイルによるプロセス管理 (#5)
- YAML ベースの設定ファイル (`~/.config/mdc/`)。拡張子省略・`~` 展開に対応
- Windows / Unix 両プラットフォーム対応 (PTY フォールバック)
- `internal/config`, `internal/runner`, `internal/logger`, `internal/pidfile` のユニットテスト
- 統合テスト (`test/integration/`)

### Fixed

- YAML ファイルが存在しても設定ファイルの読み込みに失敗するバグを修正 (#1)
- PTY 未対応の環境で TTY エラーが発生する問題を修正 (#3)

### Changed

- Phase1 コードベースのリファクタリング (パッケージ構成の整理)
