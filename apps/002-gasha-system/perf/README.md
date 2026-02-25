# Performance Test (k6)

## Directory
- `k6/`: k6 scripts
- `results/`: output files (`.gitkeep` only by default)

## Prerequisites
- API up at `http://127.0.0.1:8080`
- MySQL / memcached up
- Either `k6` installed locally, or Docker available (`grafana/k6`)

## Scripts
- `k6/user_flow.js`: user E2E (regist -> login -> gasha -> inventory -> logout)
- `k6/admin_flow.js`: admin E2E (admin regist -> login -> list -> detail)
- `k6/load_mix.js`: mixed load (default user 45 VUs + admin 5 VUs)

## Run examples
```bash
cd /Users/motiyama/spec-dev/apps/002-gasha-system

# single user flow smoke
k6 run perf/k6/user_flow.js

# single admin flow smoke
k6 run perf/k6/admin_flow.js

# mixed load (50 concurrent total)
k6 run perf/k6/load_mix.js

# output summary json for record
k6 run --summary-export perf/results/load_mix_summary.json perf/k6/load_mix.js

# docker run (if local k6 is not installed)
docker run --rm \
  -v /Users/motiyama/spec-dev/apps/002-gasha-system/perf/k6:/scripts \
  -v /Users/motiyama/spec-dev/apps/002-gasha-system/perf/results:/results \
  grafana/k6 run \
  -e BASE_URL=http://host.docker.internal:8080 \
  --summary-export=/results/load_mix_summary.json \
  /scripts/load_mix.js
```

## Thresholds
- `http_req_failed < 0.01%` (`rate < 0.0001`)
- duration threshold is currently a temporary local value; tune after baseline measurement
