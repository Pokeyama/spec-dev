# 002 Usage

## 実装コード
- `/Users/motiyama/spec-dev/apps/002-gasha-system`

## サーバー起動（ホスト）
```bash
cd /Users/motiyama/spec-dev/apps/002-gasha-system
GOCACHE=/tmp/go-cache GOTMPDIR=/tmp go run ./cmd/api
```

## インフラ起動（Docker: MySQL/memcached）
```bash
cd /Users/motiyama/spec-dev/apps/002-gasha-system
docker compose up -d mysql memcached
```

## フロント起動（ホスト）
```bash
cd /Users/motiyama/spec-dev/apps/002-gasha-system/frontend
npm install
npm test
npm run dev
```

## ヘルスチェック
```bash
curl -s http://127.0.0.1:8080/health
```

## API確認用curl（最小）
```bash
curl -s -X POST http://127.0.0.1:8080/regist \
  -H 'Content-Type: application/json' \
  -d '{"id":"alice1","password":"pass123"}'

USER_TOKEN=$(curl -s "http://127.0.0.1:8080/llogin?id=alice1&password=pass123" \
  | sed -n 's/.*"sessionToken":"\([^"]*\)".*/\1/p')

curl -s -H "Authorization: Bearer ${USER_TOKEN}" http://127.0.0.1:8080/inventory
curl -s -X POST -H "Authorization: Bearer ${USER_TOKEN}" http://127.0.0.1:8080/gasha

curl -s -X POST http://127.0.0.1:8080/admin/regist \
  -H 'Content-Type: application/json' \
  -d '{"id":"admin1","password":"pass123"}'

ADMIN_TOKEN=$(curl -s "http://127.0.0.1:8080/admin/login?id=admin1&password=pass123" \
  | sed -n 's/.*"adminSessionToken":"\([^"]*\)".*/\1/p')

curl -s -H "Authorization: Bearer ${ADMIN_TOKEN}" http://127.0.0.1:8080/account/list
```

## 負荷試験（k6）
```bash
cd /Users/motiyama/spec-dev/apps/002-gasha-system
k6 run perf/k6/load_mix.js
```

`k6` が未導入なら `perf/README.md` の Docker 実行例を利用。
