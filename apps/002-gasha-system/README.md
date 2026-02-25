# 002 Gasha System

## Backend (host run)

```bash
cd /Users/motiyama/spec-dev/apps/002-gasha-system
GOCACHE=/tmp/go-cache GOTMPDIR=/tmp go run ./cmd/api
```

## Backend infra (Docker)

```bash
cd /Users/motiyama/spec-dev/apps/002-gasha-system
docker compose up -d mysql memcached
```

API + infra all-in-one:

```bash
cd /Users/motiyama/spec-dev/apps/002-gasha-system
docker compose up -d --build
```

## Frontend (host run)

```bash
cd /Users/motiyama/spec-dev/apps/002-gasha-system/frontend
npm install
npm test
npm run dev
```

- Default UI: `http://127.0.0.1:5173`
- Default API: `http://127.0.0.1:8080` (`VITE_API_BASE_URL` で変更可能)

## Quick check

```bash
curl -s http://127.0.0.1:8080/health
```

## API smoke (minimum)

```bash
# user regist/login
curl -s -X POST http://127.0.0.1:8080/regist \
  -H 'Content-Type: application/json' \
  -d '{"id":"alice1","password":"pass123"}'

USER_TOKEN=$(curl -s "http://127.0.0.1:8080/llogin?id=alice1&password=pass123" \
  | sed -n 's/.*"sessionToken":"\([^"]*\)".*/\1/p')

# user APIs
curl -s -H "Authorization: Bearer ${USER_TOKEN}" http://127.0.0.1:8080/inventory
curl -s -X POST -H "Authorization: Bearer ${USER_TOKEN}" http://127.0.0.1:8080/gasha

# admin regist/login
curl -s -X POST http://127.0.0.1:8080/admin/regist \
  -H 'Content-Type: application/json' \
  -d '{"id":"admin1","password":"pass123"}'

ADMIN_TOKEN=$(curl -s "http://127.0.0.1:8080/admin/login?id=admin1&password=pass123" \
  | sed -n 's/.*"adminSessionToken":"\([^"]*\)".*/\1/p')

# admin APIs
curl -s -H "Authorization: Bearer ${ADMIN_TOKEN}" http://127.0.0.1:8080/account/list
curl -s -H "Authorization: Bearer ${ADMIN_TOKEN}" \
  "http://127.0.0.1:8080/account/detail?id=1"
```

## k6 (load test)

```bash
cd /Users/motiyama/spec-dev/apps/002-gasha-system
k6 run perf/k6/user_flow.js
k6 run perf/k6/admin_flow.js
k6 run perf/k6/load_mix.js
```

- Script details: `perf/README.md`
- Results output directory: `perf/results/`
- `k6` 未導入の場合は Docker でも実行可能（`perf/README.md` 参照）
