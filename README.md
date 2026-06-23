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

引数なし、または `help` を指定すると使い方を表示します。

```bash
./cursor-agent-cli
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
| `stream` | SSE イベントを NDJSON（1行1JSON）で出力 |
| `cancel` | API レスポンスをそのまま JSON 出力 |

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

### `stream` — SSE ストリーミング

`GET /v1/agents/{id}/runs/{run_id}/stream` に接続し、Server-Sent Events をリアルタイムで NDJSON 出力します。

```bash
./cursor-agent-cli stream <agent_id> <run_id>
```

ポーリング（`status --watch`）では `status` しか分からないのに対し、`stream` では以下のイベントがリアルタイムで流れます:

| イベント | 内容 |
|---|---|
| `status` | ステータス変更（`RUNNING` → `FINISHED` 等） |
| `assistant` | アシスタントのテキストデルタ（トークン単位） |
| `thinking` | 思考プロセスのテキストデルタ |
| `tool_call` | ツール呼び出し（`name`, `status`, `args`, `result`） |
| `result` | 最終結果（`text`, `durationMs`, `git`） |
| `interaction_update` | ステップ進捗、トークン消費、思考完了等の詳細イベント |
| `heartbeat` | キープアライブ |
| `error` | エラー |
| `done` | ストリーム終了 |

出力例:

```jsonl
{"event":"status","data":{"runId":"run-1","status":"RUNNING"}}
{"event":"assistant","data":{"text":"Checking"},"id":"1713033006000-0"}
{"event":"tool_call","data":{"callId":"tc-1","name":"read_file","status":"started"},"id":"1713033007000-0"}
{"event":"result","data":{"runId":"run-1","status":"FINISHED","text":"Done.","durationMs":12357}}
{"event":"done","data":{}}
```

終了コード: `done` イベント受信時は 0、`error` イベント受信時は 2。

### `cancel` — 実行のキャンセル

`POST /v1/agents/{id}/runs/{run_id}/cancel` で実行中のランをキャンセルします。

```bash
./cursor-agent-cli cancel <agent_id> <run_id>
```

出力例:

```json
{
  "id": "run-00000000-0000-0000-0000-000000000001"
}
```

キャンセルは不可逆です。会話を継続する場合は `run` で新しい実行を開始してください。既に終了した実行のキャンセルは `409` エラーになります。

## ライセンス

MIT License — 詳細は [LICENSE](LICENSE) を参照してください。
