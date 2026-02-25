import axios, { AxiosError } from 'axios'

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://127.0.0.1:8080'
})

export type ApiError = { code: string; message: string }

function mapError(error: unknown): ApiError {
  const e = error as AxiosError<{ error?: ApiError }>
  return e.response?.data?.error || { code: 'UNKNOWN', message: e.message }
}

export async function regist(id: string, password: string) {
  const { data } = await api.post('/regist', { id, password })
  return data
}

export async function login(id: string, password: string) {
  const { data } = await api.get('/llogin', { params: { id, password } })
  return data as { sessionToken: string }
}

export async function logout(token: string) {
  const { data } = await api.get('/logout', {
    headers: { Authorization: `Bearer ${token}` }
  })
  return data
}

export async function inventory(token: string) {
  const { data } = await api.get('/inventory', {
    headers: { Authorization: `Bearer ${token}` }
  })
  return data as { items: Array<{ name: string; count: number }>; credit: number }
}

export async function gasha(token: string) {
  const { data } = await api.post('/gasha', null, {
    headers: { Authorization: `Bearer ${token}` }
  })
  return data as {
    consumedCredit: number
    remainingCredit: number
    rewards: Array<{ name: string }>
  }
}

export async function gashaTen(token: string) {
  const { data } = await api.post('/gasha/ten', null, {
    headers: { Authorization: `Bearer ${token}` }
  })
  return data as {
    consumedCredit: number
    remainingCredit: number
    rewards: Array<{ name: string }>
  }
}

export async function adminRegist(id: string, password: string) {
  const { data } = await api.post('/admin/regist', { id, password })
  return data
}

export async function adminLogin(id: string, password: string) {
  const { data } = await api.get('/admin/login', { params: { id, password } })
  return data as { adminSessionToken: string }
}

export async function accountList(token: string) {
  const { data } = await api.get('/account/list', {
    headers: { Authorization: `Bearer ${token}` }
  })
  return data as {
    accounts: Array<{
      account_id: number
      login_id: string
      credit: number
      createdAt: string
    }>
  }
}

export async function accountDetail(token: string, accountID: number) {
  const { data } = await api.get('/account/detail', {
    headers: { Authorization: `Bearer ${token}` },
    params: { id: accountID }
  })
  return data as {
    account_id: number
    login_id: string
    rewards: Array<{ name: string; obtainedAt: string }>
  }
}

export { mapError }
