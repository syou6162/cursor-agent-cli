# cursor-agent-cli

[![CI](https://github.com/syou6162/cursor-agent-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/syou6162/cursor-agent-cli/actions/workflows/ci.yml)

Cursor Cloud Agent API を安全に操作するための Go 製 CLI ツールです。サブコマンド単位で API 操作を切り出し、JSON 形式で結果を出力します。

## 要件

- Go 1.22 以上

## インストール

最新版をインストールする:

```bash
go install github.com/syou6162/cursor-agent-cli@latest
```

`$GOPATH/bin`（未設定の場合は `~/go/bin`）に `cursor-agent-cli` が配置されます。PATH に含まれていることを確認してください。

## ビルド

```bash
go build -o cursor-agent-cli .
```

## 実行

引数なしで実行すると hello world 的な JSON レスポンスを返します。

```bash
./cursor-agent-cli
```

```json
{
  "message": "hello from cursor-agent-cli"
}
```

ヘルプを表示する:

```bash
./cursor-agent-cli help
```

## 開発

```bash
go vet ./...
go test ./...
go build .
```

## 認証

各サブコマンドは、環境変数 `CURSOR_CLOUD_AGENT_API_KEY` を使用して Cursor Cloud Agent API に認証します。

## サブコマンドと出力形式

| サブコマンド | 出力形式 |
|---|---|
| `models` | API レスポンスをそのまま JSON 出力 |
| `list` | API レスポンスをそのまま JSON 出力 |
| `create` | API レスポンスをそのまま JSON 出力 |
| `run` | API レスポンスをそのまま JSON 出力 |
| `status` | API レスポンス + `_cli` メタ情報（エージェント向け） |

`status` サブコマンドのみ、他のエージェントがプログラム的に扱いやすいよう CLI 独自の `_cli` フィールドを付与します。`status` / `status --watch` は常に stdout に 1 つの JSON を出力します（`status --help` を除く。ヘルプは stderr に usage を出力し、JSON は出力しません）。

### `status` の `_cli` フィールド

| フィールド | 型 | 説明 |
|---|---|---|
| `state` | string | `success`, `timeout`, `api_error`, `usage_error`, `config_error` |
| `exitCode` | int | プロセス終了コードと同じ値 |
| `pollingCount` | int | ポーリング回数（非 watch 時は 1） |
| `elapsedSeconds` | int | ポーリング開始から終了までの秒数 |
| `error` | string | エラー時のメッセージ（成功時は省略） |

### 終了コードとの対応

| `_cli.state` | 終了コード |
|---|---|
| `success` | 0 |
| `usage_error` | 1 |
| `api_error` | 2 |
| `timeout` | 3 |
| `config_error` | 3 |

### 出力例

以下は簡略化した例です。実際の出力には API レスポンス由来の `createdAt` / `updatedAt` などのフィールドも含まれます（値がある場合は `durationMs` なども含まれることがあります）。

正常終了（terminal status に到達）:

```json
{
  "id": "run-1",
  "agentId": "bc-1",
  "status": "FINISHED",
  "result": "Added README.md",
  "git": {
    "branches": [
      { "repoUrl": "https://github.com/org/repo", "prUrl": "https://github.com/org/repo/pull/123" }
    ]
  },
  "_cli": {
    "state": "success",
    "exitCode": 0,
    "pollingCount": 5,
    "elapsedSeconds": 45
  }
}
```

タイムアウト:

```json
{
  "id": "run-1",
  "agentId": "bc-1",
  "status": "RUNNING",
  "_cli": {
    "state": "timeout",
    "exitCode": 3,
    "pollingCount": 12,
    "elapsedSeconds": 180,
    "error": "timeout waiting for run to complete"
  }
}
```

API エラー（API レスポンスを取得できないため、`id` / `agentId` / `_cli` のみ出力）:

```json
{
  "id": "run-1",
  "agentId": "bc-1",
  "_cli": {
    "state": "api_error",
    "exitCode": 2,
    "pollingCount": 1,
    "elapsedSeconds": 0,
    "error": "Cursor API error (status=500): internal error"
  }
}
```

引数エラー:

```json
{
  "_cli": {
    "state": "usage_error",
    "exitCode": 1,
    "pollingCount": 0,
    "elapsedSeconds": 0,
    "error": "agent_id is required"
  }
}
```

## ライセンス

MIT License — 詳細は [LICENSE](LICENSE) を参照してください。

