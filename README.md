# cursor-agent-cli

[![CI](https://github.com/syou6162/cursor-agent-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/syou6162/cursor-agent-cli/actions/workflows/ci.yml)

Cursor Cloud Agent API を安全に操作するための Go 製 CLI ツールです。サブコマンド単位で API 操作を切り出し、JSON 形式で結果を出力します。

## 要件

- Go 1.22 以上

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

今後追加されるサブコマンドでは、環境変数 `CURSOR_CLOUD_AGENT_API_KEY` を使用して Cursor Cloud Agent API に認証します。

## ライセンス

MIT License — 詳細は [LICENSE](LICENSE) を参照してください。
