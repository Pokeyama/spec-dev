# 001: Note API (Create + List) - Implementation Log v1.0

## 1. Status
- Current Phase: Completed
- Target: `requirements.md` の AC-1 〜 AC-6 を満たす

## 3. Implementation Notes
- Go module 初期化: `go mod init noteapi`
- 実装コード配置: `apps/001-note-api/`
- 追加ファイル:
  - `apps/001-note-api/main.go`（`:8080` でサーバー起動）
  - `apps/001-note-api/note.go`（`Note` モデル）
  - `apps/001-note-api/store.go`（インメモリ保持、ID採番、50件上限）
  - `apps/001-note-api/server.go`（`POST /notes`, `GET /notes`, エラー応答）
  - `apps/001-note-api/server_test.go`（要件対応テスト）
- 実装方針:
  - タイトルは `strings.TrimSpace` 後に 1〜50 文字（rune数）で検証
  - `GET /notes` は新しい順で返却
  - 51件目以降は最古1件を削除してから追加

## 4. Test Results
- Automated:
  - 実行コマンド: `GOCACHE=/tmp/go-cache GOTMPDIR=/tmp go test ./...`
  - 結果: `ok   noteapi  0.347s`
- Manual (`curl`):
  - `POST /notes` -> `201 Created` を確認
  - `GET /notes` -> `200 OK` と登録データ返却を確認

## 5. Open Issues
- なし
