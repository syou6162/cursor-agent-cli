---
name: dev-workflow
description: cursor-agent-cli の開発ワークフロー。lint、テスト、ビルド、CIの構成を記載。コード変更時に参照する。
---

## 開発コマンド

```bash
# lint
go vet ./...

# テスト
go test ./...

# ビルド
go build -o cursor-agent-cli .

# インストール（PATH に配置）
go install github.com/syou6162/cursor-agent-cli@latest
```

## CI（GitHub Actions）

3ジョブ構成（`.github/workflows/ci.yml`）:
- **Lint**: `golangci-lint` v1.64.8
- **Test**: `go vet ./...` + `go test ./...`
- **Build**: `go build -o cursor-agent-cli .`

Go バージョン: 1.22

## プロジェクト構成

- `main.go`: エントリポイント。`cmd.NewRoot().Run(os.Args[1:])` を呼ぶだけ
- `internal/cmd/`: サブコマンド実装（root.go にルーティング、各コマンドは個別ファイル）
- `internal/cursor/`: APIクライアント、型定義、SSEパーサー
  - `client.go`: インターフェース定義 + `apiClient` 実装
  - `types.go`: リクエスト/レスポンス型
  - `sse.go`: SSEストリームリーダー

## テストパターン

- `internal/cmd/stub_client_test.go` に `stubClient` があり、各インターフェースをスタブ化
- テスト追加時は `stubClient` に新しいフィールドとメソッドを追加する
- `Root` 構造体の `clientFactory` にスタブを注入してテスト

## 認証

環境変数 `CURSOR_CLOUD_AGENT_API_KEY` を Basic 認証で使用。
