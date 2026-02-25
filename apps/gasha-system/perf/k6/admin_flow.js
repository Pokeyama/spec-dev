import http from 'k6/http'
import { check } from 'k6'

// 管理者API向けフロー。環境変数がなければローカル既定値を使う。
const BASE_URL = __ENV.BASE_URL || 'http://127.0.0.1:8080'
const PASSWORD = __ENV.ADMIN_PASSWORD || 'pass123'

export const options = {
  // 単体フロースクリプト: 1シナリオで一定負荷。
  vus: Number(__ENV.VUS || 1),
  duration: __ENV.DURATION || '10s',
  thresholds: {
    // エラーレート基準 + 暫定の応答時間基準。
    http_req_failed: ['rate<0.0001'],
    http_req_duration: ['p(95)<500']
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
  // 重複を避けるため毎回一意な管理者IDを生成。
  const suffix = `${__VU}_${__ITER}_${Date.now()}`
  const adminID = `k6_admin_${suffix}`

  // 1) 管理者登録
  const registRes = http.post(
    `${BASE_URL}/admin/regist`,
    json({ id: adminID, password: PASSWORD }),
    { headers: { 'Content-Type': 'application/json' } }
  )
  check(registRes, {
    'admin regist status 201': (r) => r.status === 201
  })

  // 2) 管理者ログインしてトークン取得
  const loginRes = http.get(
    `${BASE_URL}/admin/login?id=${encodeURIComponent(adminID)}&password=${encodeURIComponent(PASSWORD)}`
  )
  check(loginRes, {
    'admin login status 200': (r) => r.status === 200,
    'admin token exists': (r) => !!r.json('adminSessionToken')
  })
  if (loginRes.status !== 200) {
    return
  }

  const token = loginRes.json('adminSessionToken')

  // 3) アカウント一覧取得
  const listRes = http.get(`${BASE_URL}/account/list`, auth(token))
  check(listRes, {
    'account list status 200': (r) => r.status === 200,
    'account list is array': (r) => Array.isArray(r.json('accounts'))
  })

  const firstAccountID = listRes.json('accounts.0.account_id')
  if (!firstAccountID) {
    return
  }

  // 4) 一覧の先頭ユーザー詳細を取得
  const detailRes = http.get(
    `${BASE_URL}/account/detail?id=${firstAccountID}`,
    auth(token)
  )
  check(detailRes, {
    'account detail status 200': (r) => r.status === 200,
    'account detail has rewards': (r) => Array.isArray(r.json('rewards'))
  })
}
