# 001: Note API (Create + List) - Tasks v1.0

## 1. Requirements Fix
- [x] `requirements.md` の合意（API仕様、バリデーション、並び順、50件上限）

## 2. Design Fix
- [x] パッケージ構成を確定
- [x] エラーレスポンス形式を確定
- [x] 51件目登録時の挙動（最古自動削除）を確定

## 3. Implementation
- [x] Noteモデル実装
- [x] In-memory store実装
- [x] `POST /notes` 実装
- [x] `GET /notes` 実装
- [x] 50件超過時の最古削除ロジック実装

## 4. Tests
- [x] `POST /notes` 正常系
- [x] `POST /notes` 異常系（空、51文字）
- [x] `GET /notes` 並び順確認（新しい順）
- [x] 51件登録時に件数が50のまま維持されることを確認
- [x] 51件登録時に最古メモが削除されることを確認

## 5. Done Criteria
- [x] `go test ./...` が通る
- [x] `curl` で手動確認できる
- [x] `implementation.md` に結果を記録できている
