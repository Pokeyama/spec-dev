<template>
  <section class="page-grid">
    <article class="card card-accent-alt">
      <div class="card-head">
        <div>
          <p class="eyebrow">Admin Auth</p>
          <h2>管理者登録 / ログイン</h2>
        </div>
        <span class="token-badge" :class="auth.adminToken ? 'ok' : 'off'">
          admin token: {{ auth.adminToken ? 'set' : 'none' }}
        </span>
      </div>

      <p class="small">学習用途では `POST /admin/regist` を未認証で提供</p>

      <div class="form-grid">
        <label>
          <span>Admin ID</span>
          <input v-model="adminID" placeholder="admin id" />
        </label>
        <label>
          <span>Password</span>
          <input v-model="adminPass" placeholder="password" type="password" />
        </label>
      </div>
      <button :disabled="loading" @click="onAdminRegist">POST /admin/regist</button>
      <button class="secondary" :disabled="loading" @click="onAdminLogin">GET /admin/login</button>

      <p v-if="loading" class="small">loading...</p>
      <div v-if="error" class="error-box" role="alert">
        <strong>{{ error.code }}</strong>
        <span>{{ friendlyError(error.code) }}（{{ error.message }}）</span>
      </div>
    </article>

    <article class="card">
      <div class="card-head">
        <div>
          <p class="eyebrow">Admin APIs</p>
          <h2>アカウント参照</h2>
        </div>
      </div>

      <button :disabled="loading" @click="onList">GET /account/list</button>

      <template v-if="viewMode === 'detail' && detail">
        <div class="detail-head">
          <h3>Detail</h3>
          <button class="secondary back-button" :disabled="loading" @click="backToList">
            一覧に戻る
          </button>
        </div>
        <p class="small">account: {{ detail.login_id }} (#{{ detail.account_id }})</p>
        <ul v-if="detailRewards.length > 0" class="inventory-list" data-testid="admin-detail-panel">
          <li v-for="(r, i) in detailRewards" :key="i">
            <span>{{ r.name }}</span>
            <strong>{{ r.obtainedAt }}</strong>
          </li>
        </ul>
        <p v-else class="empty">詳細データはありません。</p>
      </template>

      <template v-else>
        <h3>Accounts</h3>
        <table v-if="accounts.length > 0" class="table" data-testid="admin-list-panel">
          <thead>
            <tr>
              <th>ID</th>
              <th>Login ID</th>
              <th>Credit</th>
              <th>Action</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="a in accounts" :key="a.account_id">
              <td>{{ a.account_id }}</td>
              <td>{{ a.login_id }}</td>
              <td>{{ a.credit }}</td>
              <td>
                <button
                  class="secondary inline-button"
                  :data-testid="`detail-btn-${a.account_id}`"
                  :disabled="loading"
                  @click="onDetail(a.account_id)"
                >
                  詳細 #{{ a.account_id }}
                </button>
              </td>
            </tr>
          </tbody>
        </table>
        <p v-else class="empty">一覧はまだ読み込まれていません。</p>
      </template>
    </article>
  </section>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import {
  accountDetail,
  accountList,
  adminLogin,
  adminRegist,
  mapError,
  type ApiError
} from '../api/client'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()

const adminID = ref('admin01')
const adminPass = ref('adminpass')

const loading = ref(false)
const error = ref<ApiError | null>(null)
const accounts = ref<Array<{ account_id: number; login_id: string; credit: number }>>([])
const detail = ref<{ account_id: number; login_id: string } | null>(null)
const detailRewards = ref<Array<{ name: string; obtainedAt: string }>>([])
const viewMode = ref<'list' | 'detail'>('list')

function setError(e: unknown) {
  error.value = mapError(e)
}

function friendlyError(code: string) {
  if (code === 'UNAUTHENTICATED') return 'ログインが必要です'
  if (code === 'FORBIDDEN') return '管理者権限が必要です'
  if (code === 'INSUFFICIENT_DIAMONDS') return 'ダイヤが不足しています'
  return 'エラーが発生しました'
}

async function onAdminRegist() {
  loading.value = true
  error.value = null
  try {
    await adminRegist(adminID.value, adminPass.value)
  } catch (e) {
    setError(e)
  } finally {
    loading.value = false
  }
}

async function onAdminLogin() {
  loading.value = true
  error.value = null
  try {
    const data = await adminLogin(adminID.value, adminPass.value)
    auth.setAdminToken(data.adminSessionToken)
  } catch (e) {
    setError(e)
  } finally {
    loading.value = false
  }
}

async function onList() {
  loading.value = true
  error.value = null
  if (!auth.adminToken) {
    loading.value = false
    return
  }
  try {
    const data = await accountList(auth.adminToken)
    accounts.value = data.accounts
    detail.value = null
    detailRewards.value = []
    viewMode.value = 'list'
  } catch (e) {
    setError(e)
  } finally {
    loading.value = false
  }
}

async function onDetail(accountID: number) {
  loading.value = true
  error.value = null
  if (!auth.adminToken) {
    loading.value = false
    return
  }
  try {
    const data = await accountDetail(auth.adminToken, accountID)
    detail.value = { account_id: data.account_id, login_id: data.login_id }
    detailRewards.value = data.rewards
    viewMode.value = 'detail'
  } catch (e) {
    setError(e)
  } finally {
    loading.value = false
  }
}

function backToList() {
  viewMode.value = 'list'
}
</script>
