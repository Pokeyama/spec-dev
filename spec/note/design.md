# 001: Note API (Create + List) - Design v1.0

前提: [requirements.md](./requirements.md) の FR-1 〜 FR-5 / AC-1 〜 AC-6 を満たす。

## 1. Architecture
- HTTP Handler
- Application Logic (薄くても可)
- In-memory Store

## 2. Data Model
- `Note`
  - `id` (int64)
  - `title` (string)
  - `createdAt` (time.Time)
- `NoteStore`
  - `notes` (`[]Note`)
  - `nextID` (`int64`)
  - `maxNotes` (`int`, 固定値 50)

## 3. Endpoints Design
- `POST /notes`
  - JSON decode
  - validate
  - store save
  - 50件超過時は最古1件を削除してから追加
  - `201` return
- `GET /notes`
  - store list
  - 新しい順に並べて返却（降順）
  - `200` return

## 4. Error Handling
- リクエストJSON不正: `400`
- バリデーションエラー: `400`
- 想定外エラー: `500`

## 5. Test Design
- handler単体テスト（`httptest`）
- store単体テスト
- 主要E2E風テスト（`POST -> GET`）
- 51件登録時の上限維持テスト（件数50、最古削除）

## 6. Decisions
- ID採番は単純インクリメント（1始まり）
- 時刻は `time.Now().UTC()` を使用
