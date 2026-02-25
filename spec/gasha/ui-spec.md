# 002: Gasha System - UI Spec v1.1

前提:
- 本ファイルは UI/UX の仕様のみを定義する。
- API/DB/認証の仕様は `requirements.md` / `design.md` を正とする。

## 1. Goal
- ユーザーが「登録 -> ログイン -> ガシャ実行 -> 結果確認」まで迷わず操作できる。
- 管理者が「ログイン -> 一覧 -> 詳細確認」を最短導線で実行できる。

## 2. Scope
### In Scope
- ユーザー画面: 登録、ログイン、ガシャ、在庫、ログアウト
- 管理者画面: 管理者登録、管理者ログイン、アカウント一覧、詳細
- 共通レイアウト、配色、タイポ、状態表示（loading/error/empty）

### Out of Scope
- バックエンドAPIの仕様変更
- DBスキーマ変更
- A/Bテストや解析導入

## 3. Visual Direction
- キーワード: `clean` / `energetic` / `game-like`
- 避ける方向性:
  - テンプレ感の強い単調なカード並び
  - 紫基調
  - 単色ベタ背景
- 背景は薄いグラデーション + 装飾図形で奥行きを作る。

## 4. Design Tokens
- Color
  - `--color-primary: #0EA5E9`
  - `--color-accent: #F59E0B`
  - `--color-bg: #F8FAFC`
  - `--color-surface: #FFFFFF`
  - `--color-text: #0F172A`
  - `--color-muted: #475569`
  - `--color-danger: #DC2626`
- Radius
  - `--radius-sm: 8px`
  - `--radius-md: 12px`
  - `--radius-lg: 16px`
- Spacing
  - 8pxスケール（8/16/24/32）
- Motion
  - `150ms ~ 250ms`, `ease-out`
  - 画面遷移はフェード + 軽いY移動（過剰な動きは禁止）

## 5. Information Architecture
- User
  - `/` : ヒーロー + 登録/ログイン導線
  - `/gasha` : 残高表示、単発/10連アクション、結果表示
  - `/inventory` : 所持報酬一覧、空状態
- Admin
  - `/admin` : 管理者登録/ログイン
  - `/admin/accounts` : 一覧
  - `/admin/accounts/:id` : 詳細

## 6. Screen Requirements
### 6.1 User Top/Login
- 入力欄は `id`, `password`。
- 成功時の次導線を明示（「ガシャ画面へ」）。
- エラー時はコードと意味が分かる文言を表示する。

### 6.2 Gasha
- 現在ダイヤを最上段に固定表示。
- `単発ガシャ` と `10連ガシャ` を明確に分離。
- 実行結果はモーダルまたは結果カードで即時表示。
- 実行中はボタンをdisabledにし、二重送信を防止。

### 6.3 Inventory
- `name` 単位の所持数表示。
- 空状態では「まだ報酬がありません」など次アクション付きメッセージを表示。

### 6.4 Admin
- 一覧は `account_id`, `login_id`, `credit`, `createdAt` を表示。
- 詳細は報酬履歴を時系列で表示。
- 非管理者・未認証時は明確に `403` / `401` を表示。

## 7. State Design
- Loading: スケルトンまたはローディング表示
- Empty: 空状態専用メッセージ + 次導線
- Error:
  - `401` -> 「ログインが必要です」
  - `402` -> 「ダイヤが不足しています」
  - `403` -> 「管理者権限が必要です」
  - `5xx` -> 「時間をおいて再試行してください」

## 8. Responsive
- 対応幅
  - Mobile: 360px以上
  - Desktop: 1280px以上
- 最低条件
  - 横スクロールを発生させない
  - CTAボタンを親指で押せるサイズ（高さ40px以上）

## 9. Accessibility
- ボタン・入力はキーボード操作可能
- フォーカスリングを可視化
- テキストと背景のコントラストを確保
- 管理ページでアカウント一覧画面でIDごとにボタンを作り、そのユーザーの詳細を表示
- 管理ページにて詳細を表示するときはアカウント一覧を表示しない
- ユーザーがInventoryを確認するときガシャ結果は表示しない, 逆にガシャ結果が表示されているときはInventoryは表示しない

## 10. Acceptance Criteria
1. `npm run build` が成功する。
2. User主要導線（登録/ログイン/ガシャ/在庫/ログアウト）がUI崩れなく動作する。
3. Admin主要導線（登録/ログイン/一覧/詳細）がUI崩れなく動作する。
4. `401/402/403` のエラー表示が画面上で区別できる。
5. Mobile(360px) / Desktop(1280px) で横スクロールが発生しない。
6. 色・余白・タイポグラフィの一貫性がある。

## 11. Implementation Notes
- コンポーネント分割時は「表示責務」と「API呼び出し責務」を分離する。
- スタイルはトークン（CSS variables）経由で参照し、値をハードコードしない。
