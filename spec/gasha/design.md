# 002: Gasha System (Create + List) - Design v2.2

前提: [requirements.md](./requirements.md) の FR-1 〜 FR-15 / AC-1 〜 AC-11 を満たす。

## 1. Architecture
- 実行環境
    - Docker composeで API / MySQL / memcached を実行
    - フロントエンドは開発時にホストOS上で実行（Vue dev server）

- バックエンド
    - REST API
    - DDD/DI
    - Go
    - MySQL - master, user
    - memcached - session store
      - session:{token} -> {account_id, role, exp}

- フロントエンド
    - Vue.js　+ TypeScript + Vite
    - 開発時は `npm run dev`（HMR利用）

- 負荷テスト
    - k6

## 2. Data Model
- `accounts`
  - 用途: ユーザーアカウント
  - columns:
    - `account_id` (BIGINT, PK, AUTO_INCREMENT)
    - `login_id` (VARCHAR(64), NOT NULL, UNIQUE)
    - `password_hash` (VARCHAR(255), NOT NULL)
    - `role` (ENUM('user','admin'), NOT NULL, default 'user')
    - `credit` (INT, NOT NULL, default 1000)
    - `created_at` (DATETIME, NOT NULL)
    - `updated_at` (DATETIME, NOT NULL)
  - indexes:
    - `uk_accounts_login_id (login_id)` unique
    - `idx_accounts_role (role)`

- `rewards`
  - 用途: ガシャ景品マスタ（`pokemonList.csv` から投入）
  - columns:
    - `reward_id` (INT, PK)  # csv 1列目
    - `name` (VARCHAR(64), NOT NULL)  # csv 2列目
  - indexes:
    - `idx_rewards_name (name)`

- `reward_history`
  - 用途: アカウントごとの排出履歴
  - columns:
    - `reward_history_id` (BIGINT, PK, AUTO_INCREMENT)
    - `account_id` (BIGINT, NOT NULL)
    - `reward_id` (INT, NOT NULL)
    - `obtained_at` (DATETIME, NOT NULL)
  - indexes:
    - `idx_reward_history_account_id_obtained_at (account_id, obtained_at DESC)`
    - `idx_reward_history_reward_id (reward_id)`

- リレーション方針
  - 今回は外部キー制約は作成しない
  - `account_id` / `reward_id` の存在確認はアプリケーションロジックで担保する

## 3. Endpoints Design
- 共通
  - ユーザー認証必須APIは `Authorization: Bearer <sessionToken>` を要求
  - 管理者認証必須APIは `Authorization: Bearer <adminSessionToken>` を要求
  - レスポンスは `application/json`

- `POST /regist`
  - request body: `id`, `password`
  - flow:
    - 入力値バリデーション（必須、長さ）
    - `accounts` に新規INSERT（`credit=1000`, `role='user'`）
    - unique違反は `409 ALREADY_EXISTS`

- `GET /llogin`
  - request query: `id`, `password`
  - flow:
    - `accounts.login_id` で検索
    - パスワード照合成功で user用 token 発行
    - `memcached` に `session:{token}` で保存
    - token返却

- `GET /logout`
  - auth required
  - flow:
    - `memcached` から `session:{token}` を削除
    - `{"ok": true}` を返却

- `GET /inventory`
  - auth required
  - flow:
    - `reward_history.reward_id` と `rewards.reward_id` をJOIN
    - `name` 単位で集計して `items[]` を構築
    - `accounts.account_id` の残高（`credit`）を返却

- `POST /gasha`
  - auth required
  - cost: 10
  - transaction:
    - `accounts.account_id` を `SELECT ... FOR UPDATE`
    - credit不足なら `402 INSUFFICIENT_DIAMONDS`
    - `rewards` からランダム1件選択（`reward_id`）
    - `accounts.credit = credit - 10`
    - `reward_history(account_id, reward_id, obtained_at)` に1件INSERT
    - commit

- `POST /gasha/ten`
  - auth required
  - cost: 100
  - transaction:
    - `accounts.account_id` を `SELECT ... FOR UPDATE`
    - credit不足なら `402 INSUFFICIENT_DIAMONDS`
    - `rewards` からランダム10件選択（重複許可）
    - `accounts.credit = credit - 100`
    - `reward_history(account_id, reward_id, obtained_at)` に10件INSERT（bulk insert）
    - commit

- `POST /admin/regist`
  - request body: `id`, `password`
  - flow:
    - 入力値バリデーション（必須、長さ）
    - `accounts` に新規INSERT（`role='admin'`, `credit=0`）
    - unique違反は `409 ALREADY_EXISTS`

- `GET /admin/login`
  - request query: `id`, `password`
  - flow:
    - `accounts.login_id` で検索し `role='admin'` を確認
    - パスワード照合成功で admin用 token 発行
    - `memcached` に `session:{token}` で保存
    - `adminSessionToken` 返却

- `GET /account/list`
  - admin domain only（middlewareでhost判定）
  - `Authorization: Bearer <adminSessionToken>` で管理者認証（認証失敗は `401`, ロール不足は `403`）
  - flow:
    - `accounts` 一覧取得（account_id/login_id/credit/role='user'/created_at）

- `GET /account/detail?id=...`
  - admin domain only
  - `Authorization: Bearer <adminSessionToken>` で管理者認証（認証失敗は `401`, ロール不足は `403`）
  - flow:
    - queryの `id` から `accounts.account_id` を特定
    - role='user'であることを確認
    - 対象アカウントの `reward_history` を `rewards` JOINで取得
    - `obtained_at DESC` で返却

## 4. Error Handling
- error body
  - format:
    - `{"error":{"code":"...","message":"..."}}`

- status/code mapping
  - `400 INVALID_ARGUMENT`: 必須パラメータ不足、形式不正
  - `401 UNAUTHENTICATED`: ユーザーログイン/管理者ログイン失敗、セッション不正/期限切れ
  - `402 INSUFFICIENT_DIAMONDS`: ダイヤ不足（message: `insufficient diamonds`）
  - `403 FORBIDDEN`: 管理ドメイン以外から管理APIアクセス、または `role != admin`
  - `404 NOT_FOUND`: `account_id` または `reward_id` の参照先が未存在
  - `409 ALREADY_EXISTS`: `accounts.login_id` 重複
  - `500 INTERNAL`: 想定外エラー

- logging
  - 4xx: request_id, path, code をinfoで記録
  - 5xx: stack trace付きerrorで記録

- transaction rule
  - ガシャ系は1リクエスト1トランザクション
  - エラー時は `credit` 更新と `reward_history` INSERT を必ずrollback

## 5. Test Design
- 各ロジック層での単体テスト
- E2Eテスト（ユーザー系）: アカウント登録→ガシャ実行→結果確認→ログアウト→再ログイン
- E2Eテスト（管理者系）: 管理者登録→管理者ログイン→account/list→account/detail
- 50並列での負荷走行（目標エラーレート0.01％未満）
- ログインセッションの期限切れ時の挙動の確認
- `role=user`状態で管理APIを叩いたときの挙動確認

## 6. Note
- トークンをキーとして、valueにロールなどの情報を入れておく。
