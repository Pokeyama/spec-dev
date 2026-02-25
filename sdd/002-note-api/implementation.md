# 002: Gasha System (Create + List) - Implementation Log v1.7

## 1. Status
- Current Phase: Completed
- Code Directory: `/Users/motiyama/spec-dev/apps/002-gasha-system`
- Source Specs:
  - `requirements.md`
  - `design.md`
  - `tasks.md`

## 2. Progress
- [x] 実装ディレクトリ作成
- [x] Go module 初期化（`go mod init gashasystem`）
- [x] バックエンドAPI実装（ユーザー/管理者API一式）
- [x] セッション実装（`session:{token}`）
- [x] `accounts.role ENUM('user','admin')` 方針で実装
- [x] `pokemonList.csv` を読み込む報酬シード実装
- [x] Docker起動ファイル作成（`docker-compose.yml`, `Dockerfile`）
- [x] MySQL初期化SQL作成（`accounts`, `rewards`, `reward_history` + index）
- [x] MySQLシード処理作成（`pokemonList.csv` -> `rewards`）
- [x] フロント初期実装（Vue 3 + TypeScript + Vite）
- [x] 起動手順ドキュメント作成
- [x] フロントの実行検証（`npm install`, `npm run build`, `npm run dev`）
- [x] バックエンドのMySQL/memcached実接続版への切替
- [x] k6スクリプト雛形作成（`perf/k6`）
- [x] バックエンドAPIテスト追加（正常系/異常系/期限切れ）
- [x] k6負荷テスト実測（50並列）
- [x] フロントテスト追加（画面遷移/API失敗表示/セッションガード/レスポンシブ確認）

## 3. Decisions
- バックエンド実装は `apps/002-gasha-system` に集約する
- 先に「起動できる最小API」を作ってから機能を積み上げる
- `tasks.md` の順序に沿って進める

## 4. Implementation Notes
- 実装配置:
  - backend: `apps/002-gasha-system/cmd`, `apps/002-gasha-system/internal`
  - frontend: `apps/002-gasha-system/frontend`
  - reward seed: `apps/002-gasha-system/seed/pokemonList.csv`
  - mysql init: `apps/002-gasha-system/sql/10_schema.sql`, `apps/002-gasha-system/sql/20_seed_rewards.sh`
  - k6: `apps/002-gasha-system/perf/k6`
- 実装済みエンドポイント:
  - `POST /regist`, `GET /llogin`, `GET /logout`, `GET /inventory`
  - `POST /gasha`, `POST /gasha/ten`
  - `POST /admin/regist`, `GET /admin/login`, `GET /account/list`, `GET /account/detail`
- エラーフォーマットを `{\"error\":{\"code\",\"message\"}}` に統一
- 永続化は MySQL、セッションは memcached を利用する実装へ切り替え済み
- `Draw` はDBトランザクションで `credit` 更新と `reward_history` 追加を一括処理

## 5. Verification Log
- `GOCACHE=/tmp/go-cache GOTMPDIR=/tmp go test ./...` 実行
  - 結果: `gashasystem/internal/server` のAPIテストを含めてPASS
- `docker compose config` 実行
  - 結果: compose定義の構文解決OK（`api/mysql/memcached`）
- `docker compose up -d mysql` 実行（`apps/002-gasha-system`）
  - 結果: `accounts`, `rewards`, `reward_history` が作成されることを確認
  - 結果: `SELECT COUNT(*) FROM rewards` が `1025`（CSVシード投入成功）
- 文字化け対策
  - `20_seed_rewards.sh` に `mysql --default-character-set=utf8mb4` を指定
  - APIレスポンスで日本語景品名が正しく返ることを確認（例: `タッツー`）
- `go run ./cmd/api` + `curl` で疎通確認
  - `GET /health` -> `200`
  - `POST /regist` -> `201`
  - `GET /llogin` -> `200` (`sessionToken` 取得)
  - `POST /gasha` -> `200`
  - `GET /inventory` -> `200`
  - `POST /admin/regist` -> `201`
  - `GET /admin/login` -> `200` (`adminSessionToken` 取得)
  - `GET /account/list` -> `200`
  - `GET /account/detail` -> `200`
  - `role=user` で `GET /account/list` -> `403`
  - `register` 後に MySQL上で対象 `login_id` が作成されることを確認
- frontend:
  - `npm install` 完了（依存導入済み）
  - `npm run test` -> `5 files / 9 tests` PASS
  - `npm run build` -> 成功
  - `npm run dev -- --host 127.0.0.1 --port 5173` 起動確認（`/` のHTML取得OK）
- k6:
  - スクリプト配置完了（`perf/k6/*.js`）
  - Docker経由で短時間実測（`DURATION=5s`, `USER_VUS=5`, `ADMIN_VUS=1`）
    - `http_req_failed`: `0.00%`
  - Docker経由で50並列実測（`DURATION=10s`, `USER_VUS=45`, `ADMIN_VUS=5`）
    - `http_req_failed`: `0.00%`（閾値 `rate<0.0001` を満たす）
    - `p(95)` latency: `39.01ms`
    - summary: `apps/002-gasha-system/perf/results/load_mix_50vu_20260225_181439.json`
- 手動確認:
  - user / admin の主要導線をブラウザで確認済み（2026-02-25）
  - 要件を満たすことを確認

## 6. Next Steps
1. 必要なら k6 を 60s 以上で再計測し、結果を `perf/results` に追記
