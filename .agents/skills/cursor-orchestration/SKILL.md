---
name: cursor-orchestration
description: cursor-agent-cli を使って Cursor Cloud Agent にタスクを委託・監視・キャンセル・フォローアップするオーケストレーション手順。Devin が Cursor にコーディング作業を任せるときに使用する。
---

## 概要

Devin がオーケストレーターとして Cursor Cloud Agent にタスクを委託し、監視・レビュー・フォローアップを行う。コーディング作業は Cursor に任せ、Devin は判断・監視・指示に特化する。

## 前提

- 認証: 環境変数 `CURSOR_CLOUD_AGENT_API_KEY`（Basic認証）
- CLI: `cursor-agent-cli`（org blueprint で `go install` 済み）

## ワークフロー

### 1. エージェント作成（タスク委託）

```bash
cursor-agent-cli create \
  --repo "https://github.com/<owner>/<repo>" \
  --prompt "<タスク指示>" \
  --branch main
```

- `autoCreatePR: true` は自動設定される
- レスポンスの `agent.url` をユーザーに提示する（Web UI一覧に出ないことがある）

### 2. リアルタイムストリーム監視（推奨）

```bash
cursor-agent-cli stream <agent_id> <run_id>
```

- SSE → NDJSON 出力。`assistant`, `thinking`, `tool_call`, `result`, `done` 等のイベントがリアルタイムで流れる
- ポーリングより推奨: Cursor が今何をしているか分かるため、判断精度が上がる
- 終了コード: `done` → 0、`error` → 2

### 2b. ステータスポーリング（フォールバック）

```bash
cursor-agent-cli status <agent_id> <run_id> --watch --interval 15
```

- 終了ステータス: `FINISHED`, `ERROR`, `CANCELLED`, `EXPIRED`
- `FINISHED` 時に `git.branches[0].prUrl` で PR URL 取得可能

### 3. 実行のキャンセル

```bash
cursor-agent-cli cancel <agent_id> <run_id>
```

- 応答しなくなった場合に使用。キャンセルは不可逆

### 4. PR 状態確認（FINISHED 後の必須ステップ）

`git_view_pr` で PR の状態（draft/open、CI、マージ可否）を確認する。仮定で報告しない。

### 5. 追加プロンプト（修正指示）

```bash
cursor-agent-cli run <agent_id> --prompt "<修正指示>"
```

- 前の実行が `RUNNING` 中は `409 agent_busy`。完了を待ってから投げる
- 修正プロンプトには具体的なコード例を含めると精度が上がる

### 6. エージェント/モデル一覧

```bash
cursor-agent-cli list
cursor-agent-cli models
```

## オーケストレーションパターン

### パターンA: タスク委託 → PR確認

create → stream で監視 → git_view_pr でPR確認 → 報告 → Linear に PR 紐づけ

### パターンB: CI失敗時の修正

git_pr_checks でCI確認 → 失敗内容読み取り → `run` で修正プロンプト → CI パスまで繰り返す

### パターンC: Copilot review 対応

git_view_pr でコメント検知 → trivial/obvious/ambiguous 分類 → 修正が必要なものを `run` で投入

### パターンD: Devin 自身がレビュアー

PR差分確認（テスト網羅性、コード一貫性、設計）→ 修正プロンプトを `run` で投入（1つずつシリアル）→ CI green 確認

## 既知の制約

- `autoCreatePR` は作成時のみ設定可能
- Cursor の PR はデフォルト draft になることが多い
- 1エージェント1実行。前の実行完了を待ってから次を投げる
- Copilot review 依頼は API 経由では不可（人間が GitHub UI で行う）
