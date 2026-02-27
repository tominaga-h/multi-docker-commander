---
name: mdc ps command
overview: "`mdc ps` コマンドを新規実装する。各プロジェクトの `path` で `docker compose ps --format json` を実行し、結果をパースして PROJECT / CONTAINER ID / NAME / STATUS / PORTS の統一テーブルで表示する。"
todos:
  - id: create-runner-ps
    content: "`internal/runner/ps.go` を作成: docker compose ps --format json の実行・パース・並列集約ロジック"
    status: completed
  - id: create-cmd-ps
    content: "`cmd/ps.go` を作成: Cobra コマンド定義、config 読み込み、runner 呼び出し、go-pretty テーブル表示"
    status: completed
  - id: add-parse-tests
    content: "`internal/runner/ps_test.go` を作成: JSON パースのユニットテスト"
    status: completed
  - id: run-check
    content: "`make check` を実行して lint + テストが通ることを確認"
    status: completed
isProject: false
---

# `mdc ps` コマンドの実装

## 方針

各プロジェクトの `path` ディレクトリで `docker compose ps --format json` を実行し、JSON 出力をパースして go-pretty テーブルで統一表示する。Docker API (SDK) は使わず、既存の `os/exec` パターンを踏襲する。

## 実装対象ファイル

### 1. `internal/runner/ps.go` (新規作成)

`docker compose ps` の実行と JSON パースのロジックを担当する。

- `docker compose ps --format json` を各プロジェクトの path で実行
- JSON 出力をパース (NDJSON 形式: 1行1コンテナの JSON)
- 全プロジェクトを goroutine で並列実行し結果を集約 (config の `execution_mode` に関係なく ps は常に並列)
- `docker compose` が見つからない場合は `docker-compose` にフォールバック

主な型定義:

```go
type ContainerInfo struct {
    ID      string
    Name    string
    State   string
    Status  string
    Ports   string
}

type ProjectContainers struct {
    ProjectName string
    Containers  []ContainerInfo
    Err         error
}
```

### 2. `cmd/ps.go` (新規作成)

Cobra コマンドの定義。既存の [cmd/proc.go](cmd/proc.go) のテーブル出力パターンを踏襲する。

```go
var psCmd = &cobra.Command{
    Use:   "ps [config-name]",
    Short: "Show container status for all projects",
    Args:  cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        // 1. config.Load(args[0]) で設定読み込み
        // 2. runner.CollectPS(cfg) で全プロジェクトのコンテナ情報を収集
        // 3. go-pretty テーブルで統一表示
    },
}
```

テーブルのカラム構成:

- PROJECT: YAML の project name
- CONTAINER ID: コンテナ ID (短縮12文字)
- NAME: コンテナ名
- STATUS: 実行状態 (running, exited 等)
- PORTS: ポートマッピング

### 3. `internal/runner/ps_test.go` (新規作成)

JSON パース部分のユニットテスト。

## 処理フロー

```mermaid
flowchart TD
    A["mdc ps config-name"] --> B["config.Load()"]
    B --> C["各プロジェクトで並列実行"]
    C --> D1["Project A: docker compose ps --format json"]
    C --> D2["Project B: docker compose ps --format json"]
    C --> D3["Project N: ..."]
    D1 --> E["JSON パース"]
    D2 --> E
    D3 --> E
    E --> F["go-pretty テーブルで統一表示"]
```



## 詳細設計

### docker compose ps --format json の出力形式

`docker compose ps --format json` は NDJSON (1行1オブジェクト) を返す:

```json
{"ID":"abc123","Name":"frontend-web-1","State":"running","Status":"Up 2 hours","Ports":"0.0.0.0:3000->3000/tcp"}
```

Go の struct にマッピングしてパースする。

### エラーハンドリング

- docker compose が未インストールの場合: エラーメッセージを表示して終了
- あるプロジェクトでコンテナが0件の場合: そのプロジェクトの行は出力しない
- あるプロジェクトで実行エラーの場合: そのプロジェクト名 + エラーメッセージをテーブル下に警告表示し、他のプロジェクトは正常表示する
- パスが存在しない場合: 警告表示して他のプロジェクトは続行

### STATUS の色分け (go-pretty/text)

- `running` -> 緑
- `exited` / その他 -> 赤

## 既存コードへの影響

- 他の既存ファイル: 変更なし (各コマンドは自身の `init()` で `rootCmd.AddCommand` するパターン)

