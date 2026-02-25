import http from 'k6/http'
import { check } from 'k6'

// k6実行時に環境変数があればそちらを優先。未指定時はローカル想定の既定値を使う。
const BASE_URL = __ENV.BASE_URL || 'http://127.0.0.1:8080'
const PASSWORD = __ENV.USER_PASSWORD || 'pass123'

export const options = {
  // 単体フロースクリプト: 1シナリオで一定負荷をかける。
  vus: Number(__ENV.VUS || 1),
  duration: __ENV.DURATION || '10s',
  thresholds: {
    // 要件のエラーレート基準（0.01%未満）と、暫定の応答時間基準。
    http_req_failed: ['rate<0.0001'],
    http_req_duration: ['p(95)<400']
  }
}

function json(data) {
  return JSON.stringify(data)
}

function auth(token) {
  return {
    headers: {
      Authorization: `Bearer ${token}`
    }
  }
}

export default function () {
  // VU/反復ごとに一意なIDを作り、衝突なく新規ユーザーを作成する。
  const suffix = `${__VU}_${__ITER}_${Date.now()}`
  const loginID = `k6_user_${suffix}`

  // 1) ユーザー登録
  const registRes = http.post(
    `${BASE_URL}/regist`,
    json({ id: loginID, password: PASSWORD }),
    { headers: { 'Content-Type': 'application/json' } }
  )
  check(registRes, {
    'regist status 201': (r) => r.status === 201
  })

  // 2) ログインしてセッショントークン取得
  const loginRes = http.get(
    `${BASE_URL}/llogin?id=${encodeURIComponent(loginID)}&password=${encodeURIComponent(PASSWORD)}`
  )
  check(loginRes, {
    'login status 200': (r) => r.status === 200,
    'session token exists': (r) => !!r.json('sessionToken')
  })
  if (loginRes.status !== 200) {
    return
  }

  const token = loginRes.json('sessionToken')

  // 3) 単発ガシャ実行
  const gashaRes = http.post(`${BASE_URL}/gasha`, null, auth(token))
  check(gashaRes, {
    'gasha status 200': (r) => r.status === 200,
    'gasha has rewards': (r) => Array.isArray(r.json('rewards'))
  })

  // 4) 所持一覧確認
  const invRes = http.get(`${BASE_URL}/inventory`, auth(token))
  check(invRes, {
    'inventory status 200': (r) => r.status === 200,
    'inventory has credit': (r) => typeof r.json('credit') === 'number'
  })

  // 5) ログアウト
  const logoutRes = http.get(`${BASE_URL}/logout`, auth(token))
  check(logoutRes, {
    'logout status 200': (r) => r.status === 200
  })
}
