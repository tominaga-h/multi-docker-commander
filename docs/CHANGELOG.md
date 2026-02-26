# Changelog

このプロジェクトのすべての注目すべき変更はこのファイルに記録されます。

フォーマットは [Keep a Changelog](https://keepachangelog.com/ja/1.1.0/) に準拠し、
バージョニングは [Semantic Versioning](https://semver.org/lang/ja/) に従います。

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
