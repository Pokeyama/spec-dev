# 002: Gasha System (Create + List) - Tasks v1.4

## 1. Requirements / Design Fix
- [x] `requirements.md` と `design.md` の整合確認
- [x] `ui-spec.md` を追加し、UI仕様は `ui-spec.md` を正として参照する方針を明記
- [x] セッションキー方針を `session:{token}` に統一
- [x] `accounts.role` を `ENUM('user','admin')` 方針で確定

## 2. Project Setup
- [x] Goプロジェクト初期化（module, ディレクトリ構成）
- [x] Docker Composeで API / MySQL / memcached を起動可能にする
- [x] フロントエンドは開発時にホストOS上で起動する（非Docker）
- [x] 設定値（DB接続、memcached接続、token TTL）を環境変数化

## 3. DB Schema / Seed
- [x] `accounts` テーブル作成（`role ENUM('user','admin')` を含む）
- [x] `rewards` テーブル作成
- [x] `reward_history` テーブル作成
- [x] インデックス作成（`uk_accounts_login_id`, `idx_accounts_role`, 履歴系index）
- [x] `pokemonList.csv` を `rewards` に投入するシード処理を作成

## 4. Session / Auth
- [x] `session:{token}` の保存・取得・削除処理を実装
- [x] ユーザー認証ミドルウェア実装（`Authorization: Bearer <sessionToken>`）
- [x] 管理者認証ミドルウェア実装（`Authorization: Bearer <adminSessionToken>`）
- [x] `role != admin` の場合に `403 FORBIDDEN` を返す

## 5. User API Implementation
- [x] `POST /regist` 実装（`role='user'`, `credit=1000`）
- [x] `GET /llogin` 実装（token発行 + セッション保存）
- [x] `GET /logout` 実装（セッション削除）
- [x] `GET /inventory` 実装（所持一覧 + 残高）
- [x] `POST /gasha` 実装（10ダイヤ消費 + 1件排出）
- [x] `POST /gasha/ten` 実装（100ダイヤ消費 + 10件排出）
- [x] ダイヤ不足時 `402 INSUFFICIENT_DIAMONDS` を返す

## 6. Admin API Implementation
- [x] `POST /admin/regist` 実装（`role='admin'`, `credit=0`）
- [x] `GET /admin/login` 実装（`role='admin'` のみ成功）
- [x] `GET /account/list` 実装（`role='user'` のみ一覧）
- [x] `GET /account/detail?id={account_id}` 実装（`role='user'` 対象）

## 7. Frontend Setup
- [x] Vue 3 + TypeScript + Vite プロジェクト初期化
- [x] Vue Router セットアップ（user画面 / admin画面）
- [x] APIクライアント層作成（base URL, 共通ヘッダー, エラーハンドリング）
- [x] セッション管理（`sessionToken` / `adminSessionToken`）実装
- [x] 開発サーバーは `npm run dev` でホストOSから起動（HMR確認）

## 8. Frontend Implementation
- [x] ユーザー登録/ログイン画面実装（`POST /regist`, `GET /llogin`）
- [x] ユーザーガシャ画面実装（`POST /gasha`, `POST /gasha/ten`）
- [x] 所持一覧画面実装（`GET /inventory`）
- [x] ログアウト導線実装（`GET /logout`）
- [x] 管理者登録/ログイン画面実装（`POST /admin/regist`, `GET /admin/login`）
- [x] 管理者一覧画面実装（`GET /account/list`）
- [x] 管理者詳細画面実装（`GET /account/detail?id={account_id}`）
- [x] `401/402/403` エラー表示を画面に反映

## 8.1 Frontend Implementation (UI Spec v1.1 Delta)
- [x] 管理画面で `account/list` の各行にID別の詳細表示ボタンを追加
- [x] 管理画面で詳細表示中はアカウント一覧を非表示にする
- [x] ユーザー画面で `Last Rewards` と `Inventory` の表示を排他にする

## 9. Error / Transaction
- [x] エラーレスポンス形式を統一（`error.code`, `error.message`）
- [x] ステータスコード運用を実装（400/401/402/403/404/409/500）
- [x] ガシャ処理をトランザクション化（rollback保証）

## 10. Tests (Backend / API)
- [x] 単体テスト（バリデーション、認証、ロール判定）
- [x] APIテスト（正常系/異常系）
- [x] E2E（ユーザー系）: regist -> login -> gasha -> inventory -> logout
- [x] E2E（管理者系）: admin/regist -> admin/login -> account/list -> account/detail
- [x] 権限テスト: `role=user` で管理APIが `403`
- [x] セッション期限切れテスト: 保護APIが `401`
- [x] 負荷テスト（k6, 同時50アクセスでエラーレート0.01%未満）

## 11. Tests (Frontend)
- [x] 主要画面遷移テスト（ユーザー系/管理者系）
- [x] API失敗時の表示テスト（`401/402/403`）
- [x] セッション未保持時のガード挙動確認
- [x] レスポンシブ確認（最低: モバイル幅 / デスクトップ幅）
- [x] UI Spec v1.1 追加要件テスト（ユーザー表示排他、管理詳細表示）

## 12. Done Criteria
- [x] `go test ./...` が通る
- [x] 主要APIを `curl` で確認済み
- [x] フロントがローカル起動し主要導線を手動確認済み
- [x] `requirements.md` の AC-1 〜 AC-11 を満たす

## 13. Write Usage
- [x] サーバー起動方法を`spec/gasha`以下にREADME.mdとして書き込み
