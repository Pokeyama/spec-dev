import http from 'k6/http'
import { check } from 'k6'

// 混合負荷シナリオ: user系とadmin系を同時に流して全体負荷を見る。
const BASE_URL = __ENV.BASE_URL || 'http://127.0.0.1:8080'
const USER_PASSWORD = __ENV.USER_PASSWORD || 'pass123'
const ADMIN_PASSWORD = __ENV.ADMIN_PASSWORD || 'pass123'

export const options = {
  // 2シナリオを同時実行。既定は user45 + admin5 = 合計50VU。
  scenarios: {
    user_flow: {
      executor: 'constant-vus',
      vus: Number(__ENV.USER_VUS || 45),
      duration: __ENV.DURATION || '60s',
      exec: 'userFlow'
    },
    admin_flow: {
      executor: 'constant-vus',
      vus: Number(__ENV.ADMIN_VUS || 5),
      duration: __ENV.DURATION || '60s',
      exec: 'adminFlow'
    }
  },
  thresholds: {
    // 要件基準: エラーレート0.01%未満。
    http_req_failed: ['rate<0.0001'],
    http_req_duration: ['p(95)<700']
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

function postJSON(path, body) {
  return http.post(`${BASE_URL}${path}`, json(body), {
    headers: { 'Content-Type': 'application/json' }
  })
}

export function userFlow() {
  // user系: regist -> login -> gasha -> inventory -> logout
  const suffix = `${__VU}_${__ITER}_${Date.now()}`
  const loginID = `k6_mix_user_${suffix}`

  const registRes = postJSON('/regist', { id: loginID, password: USER_PASSWORD })
  check(registRes, { 'mix user regist 201': (r) => r.status === 201 })
  if (registRes.status !== 201) {
    return
  }

  const loginRes = http.get(
    `${BASE_URL}/llogin?id=${encodeURIComponent(loginID)}&password=${encodeURIComponent(USER_PASSWORD)}`
  )
  check(loginRes, {
    'mix user login 200': (r) => r.status === 200,
    'mix user token': (r) => !!r.json('sessionToken')
  })
  if (loginRes.status !== 200) {
    return
  }

  const token = loginRes.json('sessionToken')

  const gashaRes = http.post(`${BASE_URL}/gasha`, null, auth(token))
  check(gashaRes, { 'mix user gasha 200': (r) => r.status === 200 })

  const invRes = http.get(`${BASE_URL}/inventory`, auth(token))
  check(invRes, { 'mix user inventory 200': (r) => r.status === 200 })

  const logoutRes = http.get(`${BASE_URL}/logout`, auth(token))
  check(logoutRes, { 'mix user logout 200': (r) => r.status === 200 })
}

export function adminFlow() {
  // admin系: admin/regist -> admin/login -> account/list -> account/detail
  const suffix = `${__VU}_${__ITER}_${Date.now()}`
  const adminID = `k6_mix_admin_${suffix}`

  const registRes = postJSON('/admin/regist', { id: adminID, password: ADMIN_PASSWORD })
  check(registRes, { 'mix admin regist 201': (r) => r.status === 201 })
  if (registRes.status !== 201) {
    return
  }

  const loginRes = http.get(
    `${BASE_URL}/admin/login?id=${encodeURIComponent(adminID)}&password=${encodeURIComponent(ADMIN_PASSWORD)}`
  )
  check(loginRes, {
    'mix admin login 200': (r) => r.status === 200,
    'mix admin token': (r) => !!r.json('adminSessionToken')
  })
  if (loginRes.status !== 200) {
    return
  }

  const token = loginRes.json('adminSessionToken')

  const listRes = http.get(`${BASE_URL}/account/list`, auth(token))
  check(listRes, { 'mix admin list 200': (r) => r.status === 200 })

  const firstAccountID = listRes.json('accounts.0.account_id')
  if (!firstAccountID) {
    return
  }

  const detailRes = http.get(`${BASE_URL}/account/detail?id=${firstAccountID}`, auth(token))
  check(detailRes, { 'mix admin detail 200': (r) => r.status === 200 })
}
