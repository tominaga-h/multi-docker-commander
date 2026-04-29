---
name: release
description: mdc プロジェクトのリリース作業を実行するスキル。「リリースして」「v0.x.x をリリース」「release」「タグを切ってリリース」「mainにマージしてリリース」「GitHub Releaseを作って」等の依頼で起動する。develop ブランチで `make check` を実行 → リリースノート生成 → CHANGELOG 更新 → README/version.go のバージョン書き換え → develop コミット → main マージ → タグ打ち → push → `gh release create` → リリースバイナリのビルド & アップロード までを一気通貫で行う。ユーザー指定タグ（例 v0.2.0）を引数として受け取る。
---

# /release スキル

ユーザーが指定したタグをもとに mdc プロジェクトのリリース準備を実行する。

## 注意事項

- ユーザーが指定したタグはこれから作成するものなので、存在確認は不要。
- コミットメッセージ・リリースノートはすべて日本語で書く（プロジェクト規約）。

## 手順

### 0. 事前チェック

- `develop` ブランチにいることを確認する。異なるブランチの場合はユーザーに警告して確認を取る。
- `make check`（lint + 全テスト）を実行し、パスすることを確認する。失敗した場合はリリース手順を中断する。

### 1. リリースノートの作成

- Git コマンド（`git describe --tags --abbrev=0` など）で直前のタグを特定する。
- `git log <直前のタグ>..HEAD --oneline` で差分コミットを確認する。
- `docs/release` フォルダ配下に `<タグ>.md`（例: タグが `v0.1.0` なら `docs/release/v0.1.0.md`）を作成する。
- 差分コミットの内容をもとにリリースノートを生成する。
- 既存のリリースノート（`docs/release/` 内）のフォーマットに合わせること。セクション構成: 主な機能 / バグ修正 / 内部改善 / 対応プラットフォーム / 既知の制限。

### 2. チェンジログに追記

- `docs/CHANGELOG.md` に新しいバージョンのエントリを追記する。
- [Keep a Changelog](https://keepachangelog.com/ja/1.1.0/) 形式に従うこと。
  - セクション名: `Added` / `Changed` / `Fixed` / `Removed` 等
  - 日付フォーマット: `YYYY-MM-DD`
  - 見出し例: `## [0.2.0] - 2026-02-28`
- 既存エントリの**上**に新しいバージョンを追加する（最新が上）。

### 3. バージョンの変更

以下の既存ドキュメント等に記載されたバージョンを、指定されたタグに変更する。

- `README.md` および `docs/README_ja.md`
  - version バッヂの値、リンク先のリリース URL（タグそのまま。例: `v0.2.0`）
  - curl コマンドのアセット URL（タグそのまま。例: `v0.2.0`）
- `internal/version/version.go`
  - バージョンの値（`v` prefix を除いた値を設定する。例: タグが `v0.2.0` なら `"0.2.0"`）

### 4. コミット・マージ・タグ打ち・プッシュ・CI 待機・リリースノート上書き

リリース本体（GitHub Release ページ作成 + マルチプラットフォームバイナリのビルド & 添付）は `.github/workflows/release.yml` がタグ push を契機に自動実行する。
このスキルがやることは、コミット → タグ push → CI 完了待機 → 手書きリリースノートで上書き、まで。

事前にユーザーに以下の操作を自動実行してよいか確認すること。
確認してユーザーが NG を返したら、**フォールバック手順**（後述）を提示して処理を一時停止すること。

#### 自動実行する操作

1. **develop でのコミット**
   - 変更したファイルを `git add` する。
   - `git commit -m "Release <タグ>"` でコミットする。
2. **main ブランチへマージしてタグを打つ**
   - `git checkout main`
   - `git merge --no-ff develop -m "Merge branch 'develop' for release <タグ>"`
   - `git tag <タグ>`
3. **リモートへのプッシュ**
   - `git push origin main --tags`（タグ push でリリースワークフローが起動する）
   - `git checkout develop`（作業ブランチに戻る）
   - `git push origin develop`
4. **CI（Release ワークフロー）の完了待機**
   - `gh run watch --exit-status $(gh run list --workflow=release.yml --branch=main --limit=1 --json databaseId --jq '.[0].databaseId')`
   - ワークフローが失敗した場合は中断してユーザーに報告する（リリースノート上書きには進まない）。
5. **GitHub Release のリリースノートを上書き**
   - CI が `softprops/action-gh-release` で生成した自動リリースノートを、手書きの `docs/release/<タグ>.md` で上書きする。
   - `gh release edit <タグ> --notes-file docs/release/<タグ>.md`

#### フォールバック手順（ユーザーが NG を返した場合）

以下のように実行すべきコマンド一覧を `<変更ファイル>` と `<タグ>` を実際の値に置換した状態で表示し、処理を一時停止する。

```bash
以下のコマンドを手動で実行してください:

# develop でコミット
git add <変更ファイル一覧>
git commit -m "Release <タグ>"

# main へマージしてタグ打ち
git checkout main
git merge develop
git tag <タグ>

# プッシュと作業ブランチへの復帰（タグ push で CI のリリースワークフローが起動）
git push origin main
git push origin <タグ>
git checkout develop
git push origin develop

# CI（Release ワークフロー）の完了を待機
gh run watch --exit-status $(gh run list --workflow=release.yml --branch=main --limit=1 --json databaseId --jq '.[0].databaseId')

# 自動生成されたリリースノートを手書きノートで上書き
gh release edit <タグ> --notes-file docs/release/<タグ>.md
```
