# 001: Note API (Create + List) - Requirements v1.0

## 1. Goal
Goでの最小API開発を、cc-sddの流れ（Requirements -> Design -> Tasks -> Implementation）で一周する。

## 2. Problem
メモを1件登録して、登録済みメモを一覧表示したい。

## 3. Scope
### In Scope
- `POST /notes` でメモを1件作成
- `GET /notes` でメモ一覧取得
- バリデーション（title必須、1〜50文字）
- JSONレスポンス
- メモは最大50件まで保持
- 一覧は新しい順で返却
- 51件目登録時は最古メモを自動削除

### Out of Scope
- 認証
- DB永続化（今回はインメモリ）
- 更新・削除

## 4. Functional Requirements
- FR-1: 有効な `title` の `POST /notes` は `201 Created` を返す
- FR-2: `title` が空文字または51文字以上の場合は `400 Bad Request` を返す
- FR-3: `GET /notes` は保持中メモを新しい順で返す
- FR-4: 保持件数が50件の状態で新規作成した場合、最古メモを削除してから追加する
- FR-5: エラー応答は `error.code` / `error.message` を持つ

## 5. API Contract
### POST /notes
Request:
```json
{
  "title": "buy milk"
}
```

Success Response: `201 Created`
```json
{
  "id": 1,
  "title": "buy milk",
  "createdAt": "2026-02-25T13:00:00Z"
}
```

Validation Error: `400 Bad Request`
```json
{
  "error": {
    "code": "INVALID_ARGUMENT",
    "message": "title must be 1-50 characters"
  }
}
```

### GET /notes
Success Response: `200 OK`
```json
{
  "notes": [
    {
      "id": 1,
      "title": "buy milk",
      "createdAt": "2026-02-25T13:00:00Z"
    }
  ]
}
```

## 6. Acceptance Criteria
1. `POST /notes` に有効な `title` を送ると `201` と作成結果が返る
2. `POST /notes` で `title` が空文字、または51文字以上なら `400` が返る
3. `GET /notes` は作成済みメモを新しい順で返す
4. サーバー再起動でデータは消えてよい（インメモリ）
5. 51件目のメモを登録してもエラーにせず `201` を返し、保持件数は50件以下を維持する
6. 51件目登録後、最古の1件は削除される

## 7. Notes (C# 比較)
- C#でいう `Controller + Service + InMemoryRepository` の最小構成に近い
- Goでは最初は `handler` と `store` を分けるだけで十分
