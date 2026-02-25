# 002: Gasha System (Create + List) - Requirements v2.2

## 1. Goal
単発ガシャと10連ガシャを実装する。

## 2. Problem
- アカウント作成にはIDとパスワードが必要
- アカウント毎に初期クレジット1000ダイヤが付与されている
- ガシャは1回10ダイヤ
- ガシャの景品マスタはpokemonList.csvの1列目をID、2列目をnameとしてDBに登録しておき、ランダムに排出
- アカウント毎に獲得したポケモン一覧を見れるページが存在
- 管理用ページが別ドメインにあり、アカウント一覧とアカウントごとの報酬獲得状況が見れる
- jsonレスポンス

## 3. Scope
### In Scope
- `POST /regist` 入力したIDとパスワードからアカウント作成
- `GET /llogin` アカウントのログイン
- `GET /logout` アカウントのログアウト
- `GET /inventory` 所持しているガシャ報酬
- `POST /gasha` 単発ガシャ
- `POST /gasha/ten` 10連ガシャ
- `POST /admin/regist` 管理用API: 管理者アカウント発行
- `GET /admin/login` 管理用API: 管理者用ログイン
- `GET /account/list` 管理用API: 発行済アカウント一覧が見れる
- `GET /account/detail` 管理用API: アカウントごとの報酬一覧が見れる

### Out of Scope
- パスワード再発行
- ダイヤ購入・課金
- 景品マスタ（pokemonList.csv）の更新API
- ガシャ提供割合の管理画面編集
- マルチテナント対応

## 4. Functional Requirements
- FR-1: `POST /regist` は新規アカウントを作成し、初期クレジット1000ダイヤを付与し、`role=user` を設定する
- FR-2: 既に存在するIDで `POST /regist` した場合はエラーを返す
- FR-3: `GET /llogin` はID/パスワード認証に成功した場合、セッショントークンを返す
- FR-4: 認証が必要なAPIに未ログインでアクセスした場合は `401 Unauthorized` を返す
- FR-5: `GET /logout` はログインセッションを無効化する
- FR-6: `GET /inventory` はログイン中アカウントの所持ポケモン一覧を返す
- FR-7: `POST /gasha` は10ダイヤ消費し、`pokemonList.csv` の2列目 `name` から1体をランダム排出する
- FR-8: `POST /gasha/ten` は100ダイヤ消費し、`name` から10体をランダム排出する
- FR-9: ダイヤ不足時は排出も消費も行わず、`402 Payment Required` を返す
- FR-10: ガシャで排出されたポケモンは当該アカウントの所持一覧へ即時反映される
- FR-11: `POST /admin/regist` は管理者アカウントを作成し、`role=admin` を設定する
- FR-12: `GET /admin/login` は `role=admin` アカウントのログイン成功時に管理用セッショントークンを返す
- FR-13: `GET /account/list` は管理者ログイン済みの場合のみ、アカウント一覧（`account_id`, `login_id`, 残高, 作成日時）を返す
- FR-14: `GET /account/detail` は管理者ログイン済みの場合のみ、`account_id` で指定したアカウントの獲得報酬履歴を返す
- FR-15: 全APIのレスポンスはJSONとする

## 5. API Contract
### POST /regist
Request:
```json
{
  "id": "alice",
  "password": "pass1234"
}
```
Success Response: `201 Created`
```json
{
  "id": "alice",
  "credit": 1000,
  "role": "user"
}
```
Duplicate ID Error: `409 Conflict`
```json
{
  "error": {
    "code": "ALREADY_EXISTS",
    "message": "id already exists"
  }
}
```

### GET /llogin
Request Query:
- `id`
- `password`

Success Response: `200 OK`
```json
{
  "sessionToken": "token-xxxx"
}
```
Auth Error: `401 Unauthorized`
```json
{
  "error": {
    "code": "UNAUTHENTICATED",
    "message": "invalid id or password"
  }
}
```

### GET /logout
Request Header:
- `Authorization: Bearer <sessionToken>`

Success Response: `200 OK`
```json
{
  "ok": true
}
```

### GET /inventory
Request Header:
- `Authorization: Bearer <sessionToken>`

Success Response: `200 OK`
```json
{
  "items": [
    {
      "name": "Pikachu",
      "count": 2
    }
  ],
  "credit": 890
}
```

### POST /gasha
Request Header:
- `Authorization: Bearer <sessionToken>`

Success Response: `200 OK`
```json
{
  "consumedCredit": 10,
  "remainingCredit": 990,
  "rewards": [
    {
      "name": "Bulbasaur"
    }
  ]
}
```
Insufficient Diamonds Error: `402 Payment Required`
```json
{
  "error": {
    "code": "INSUFFICIENT_DIAMONDS",
    "message": "insufficient diamonds"
  }
}
```

### POST /gasha/ten
Request Header:
- `Authorization: Bearer <sessionToken>`

Success Response: `200 OK`
```json
{
  "consumedCredit": 100,
  "remainingCredit": 900,
  "rewards": [
    {
      "name": "Bulbasaur"
    }
  ]
}
```
Insufficient Diamonds Error: `402 Payment Required`
```json
{
  "error": {
    "code": "INSUFFICIENT_DIAMONDS",
    "message": "insufficient diamonds"
  }
}
```

### POST /admin/regist
Request:
```json
{
  "id": "admin01",
  "password": "adminpass"
}
```
Success Response: `201 Created`
```json
{
  "id": "admin01",
  "role": "admin"
}
```
Duplicate ID Error: `409 Conflict`
```json
{
  "error": {
    "code": "ALREADY_EXISTS",
    "message": "id already exists"
  }
}
```

### GET /admin/login
Request Query:
- `id`
- `password`

Success Response: `200 OK`
```json
{
  "adminSessionToken": "admin-token-xxxx"
}
```
Auth Error: `401 Unauthorized`
```json
{
  "error": {
    "code": "UNAUTHENTICATED",
    "message": "invalid id or password"
  }
}
```
Role Error: `403 Forbidden`
```json
{
  "error": {
    "code": "FORBIDDEN",
    "message": "admin role required"
  }
}
```

### GET /account/list
Request Header:
- `Authorization: Bearer <adminSessionToken>`

Success Response: `200 OK`
```json
{
  "accounts": [
    {
      "account_id": 1,
      "login_id": "alice",
      "credit": 900,
      "createdAt": "2026-02-25T15:00:00Z"
    }
  ]
}
```

### GET /account/detail
Request Query:
- `id` (`account_id`)
Request Header:
- `Authorization: Bearer <adminSessionToken>`

Success Response: `200 OK`
```json
{
  "account_id": 1,
  "login_id": "alice",
  "rewards": [
    {
      "name": "Bulbasaur",
      "obtainedAt": "2026-02-25T15:10:00Z"
    }
  ]
}
```

## 6. Acceptance Criteria
1. 新規登録時にアカウントが作成され、残高が1000ダイヤである
2. 重複ID登録は `409` で失敗する
3. ログイン成功時にセッショントークンを取得でき、未ログイン時は保護APIが `401` になる
4. 単発ガシャ実行で残高が10減り、報酬が1件増える
5. 10連ガシャ実行で残高が100減り、報酬が10件増える
6. ダイヤ不足でガシャ実行した場合、残高と報酬は変化せず、`402` と `insufficient diamonds` が返る
7. `GET /inventory` でユーザーの所持報酬と残高が確認できる
8. `POST /admin/regist` で管理者アカウントを作成できる
9. `GET /admin/login` で管理用セッショントークンを取得できる
10. 管理用APIは管理者ログイン済みトークンでのみ利用でき、`account_id` を基準にアカウント一覧と個別報酬履歴が確認できる
11. 負荷走行を行い同時50アクセスでもエラー率が0.01%未満であること

## 7. Notes
- ここでは技術選定（MySQLなど）は記載しない
- 実装方針や構成、認証の持ち方は `design.md` で決める
- `/admin/regist` は学習用途のため未認証で提供。本番ではIP制限または管理者認証必須。