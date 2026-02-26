# MDC (Multi-Docker-Compose)

複数リポジトリにまたがる Docker 環境の起動・停止を、1つのコマンドで一括管理・実行するための CLI ツール。

## 特徴

- 複数リポジトリの Docker Compose を `mdc up` / `mdc down` で一括操作
- プロジェクト間の並列 (`parallel`) / 直列 (`sequential`) 実行モードを選択可能
- バックグラウンドプロセスの管理と状態確認 (`mdc procs`)
- プロジェクト名プレフィックス付きのログ出力で視認性を確保
- YAML ベースのシンプルな設定ファイル

## インストール

### GitHub Releases からダウンロード

[最新リリース](https://github.com/tominaga-h/multi-docker-compose/releases/latest)からビルド済みバイナリをダウンロードできます。

```bash
curl -L -o mdc https://github.com/tominaga-h/multi-docker-compose/releases/download/v0.1.0/mdc
chmod +x mdc
sudo mv mdc /usr/local/bin/
```

### ソースからビルド

```bash
git clone https://github.com/tominaga-h/multi-docker-compose.git
cd multi-docker-compose
make build
```

`./mdc` バイナリが生成されます。パスの通った場所にコピーしてください。

### バージョン埋め込みビルド

Git タグからバージョン情報を埋め込む場合:

```bash
make build-v
```

## クイックスタート

### 1. 設定ディレクトリの作成

```bash
mkdir -p ~/.config/mdc
```

### 2. 設定ファイルの作成

`~/.config/mdc/myproject.yml` を作成します:

```yaml
execution_mode: "parallel"
projects:
  - name: "Frontend"
    path: "/path/to/frontend-repo"
    commands:
      up:
        - "docker compose up -d"
      down:
        - "docker compose down"

  - name: "Backend-API"
    path: "/path/to/backend-api-repo"
    commands:
      up:
        - "docker compose up -d"
      down:
        - "docker compose down"
```

### 3. 起動と停止

```bash
mdc up myproject      # 全プロジェクトを起動
mdc down myproject    # 全プロジェクトを停止
```

拡張子 `.yml` は省略可能です。

## 設定ファイル

設定ファイルは `~/.config/mdc/` に YAML 形式で配置します。

### フィールド一覧

| フィールド | 必須 | 説明 |
|---|---|---|
| `execution_mode` | Yes | `"parallel"` (並列) または `"sequential"` (直列) |
| `projects` | Yes | プロジェクト定義のリスト (1つ以上) |
| `projects[].name` | Yes | プロジェクト名 (ログ出力のプレフィックスに使用) |
| `projects[].path` | Yes | プロジェクトのディレクトリパス (`~` 展開対応) |
| `projects[].commands.up` | No | 起動時に実行するコマンドのリスト |
| `projects[].commands.down` | No | 停止時に実行するコマンドのリスト |

### コマンドの記述形式

コマンドは文字列として記述できます:

```yaml
commands:
  up:
    - "docker compose up -d"
    - "echo done"
```

バックグラウンド実行が必要な場合はオブジェクト形式を使用します:

```yaml
commands:
  up:
    - command: "npm run dev"
      background: true
```

### 実行モード

- **parallel**: 全プロジェクトを Goroutine で同時に実行します。各プロジェクト内のコマンドは直列で実行されます。
- **sequential**: プロジェクトを定義順に1つずつ処理します。

## コマンドリファレンス

### `mdc up [config-name]`

指定した設定ファイルを読み込み、各プロジェクトの `commands.up` を実行します。

```bash
mdc up myproject
```

### `mdc down [config-name]`

指定した設定ファイルを読み込み、各プロジェクトの `commands.down` を実行します。`mdc up` で起動したバックグラウンドプロセスも自動的に停止します。

```bash
mdc down myproject
```

### `mdc list`

`~/.config/mdc/` 内の設定ファイル一覧を表示します。エイリアスとして `mdc ls` も使用できます。

```bash
mdc list
mdc ls
```

### `mdc procs [config-name]`

mdc が管理しているバックグラウンドプロセスの一覧を表示します。設定名を省略すると全設定のプロセスを表示します。

```bash
mdc procs              # 全設定のプロセスを表示
mdc procs myproject    # 特定の設定のプロセスのみ表示
```

### `mdc --version`

バージョン情報を表示します。

```bash
mdc --version
mdc -v
```

## 開発

### 必要な環境

- Go 1.25+

### ビルド

```bash
make build
```

### テスト

```bash
make test             # internal パッケージのテスト
make test-integration # 統合テスト
make test-all         # 全テスト
make test-cover       # カバレッジ付きテスト
```

## ライセンス

TBD
