<template>
  <section class="page-grid">
    <article class="card card-accent">
      <div class="card-head">
        <div>
          <p class="eyebrow">User Auth</p>
          <h2>ユーザー登録 / ログイン</h2>
        </div>
        <span class="token-badge" :class="auth.userToken ? 'ok' : 'off'">
          token: {{ auth.userToken ? 'set' : 'none' }}
        </span>
      </div>

      <div class="form-grid">
        <label>
          <span>Register ID</span>
          <input v-model="registID" placeholder="id" />
        </label>
        <label>
          <span>Register Password</span>
          <input v-model="registPass" placeholder="password" type="password" />
        </label>
      </div>
      <button :disabled="loading" @click="onRegist">POST /regist</button>

      <div class="form-grid gap-top">
        <label>
          <span>Login ID</span>
          <input v-model="loginID" placeholder="id" />
        </label>
        <label>
          <span>Login Password</span>
          <input v-model="loginPass" placeholder="password" type="password" />
        </label>
      </div>
      <button class="secondary" :disabled="loading" @click="onLogin">GET /llogin</button>

      <p v-if="loading" class="small">loading...</p>
      <div v-if="error" class="error-box" role="alert">
        <strong>{{ error.code }}</strong>
        <span>{{ friendlyError(error.code) }}（{{ error.message }}）</span>
      </div>
    </article>

    <article class="card">
      <div class="card-head">
        <div>
          <p class="eyebrow">Gasha</p>
          <h2>実行コントロール</h2>
        </div>
        <p class="credit-chip">Credit: <strong>{{ credit }}</strong></p>
      </div>

      <div class="button-row">
        <button :disabled="loading" @click="onGasha">POST /gasha</button>
        <button :disabled="loading" @click="onGashaTen">POST /gasha/ten</button>
        <button class="warn" :disabled="loading" @click="onLogout">GET /logout</button>
      </div>
      <button class="secondary" :disabled="loading" @click="onInventory">GET /inventory</button>

      <template v-if="viewMode === 'rewards'">
        <h3>Last Rewards</h3>
        <ul v-if="lastRewards.length > 0" class="pill-list" data-testid="rewards-panel">
          <li v-for="(r, i) in lastRewards" :key="i">{{ r.name }}</li>
        </ul>
        <p v-else class="empty">まだガシャ結果がありません。`POST /gasha` を実行してください。</p>
      </template>

      <template v-else-if="viewMode === 'inventory'">
        <h3>Inventory</h3>
        <ul v-if="items.length > 0" class="inventory-list" data-testid="inventory-panel">
          <li v-for="item in items" :key="item.name">
            <span>{{ item.name }}</span>
            <strong>x{{ item.count }}</strong>
          </li>
        </ul>
        <p v-else class="empty">まだ報酬がありません。</p>
      </template>

      <p v-else class="empty">`POST /gasha` または `GET /inventory` を実行すると結果を表示します。</p>
    </article>
  </section>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import {
  gasha,
  gashaTen,
  inventory,
  login,
  logout,
  mapError,
  regist,
  type ApiError
} from '../api/client'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()

const registID = ref('alice')
const registPass = ref('pass1234')
const loginID = ref('alice')
const loginPass = ref('pass1234')

const loading = ref(false)
const error = ref<ApiError | null>(null)
const credit = ref(0)
const lastRewards = ref<Array<{ name: string }>>([])
const items = ref<Array<{ name: string; count: number }>>([])
const viewMode = ref<'none' | 'rewards' | 'inventory'>('none')

function setError(e: unknown) {
  error.value = mapError(e)
}

function friendlyError(code: string) {
  if (code === 'UNAUTHENTICATED') return 'ログインが必要です'
  if (code === 'INSUFFICIENT_DIAMONDS') return 'ダイヤが不足しています'
  if (code === 'FORBIDDEN') return '権限が不足しています'
  return 'エラーが発生しました'
}

async function onRegist() {
  loading.value = true
  error.value = null
  try {
    const data = await regist(registID.value, registPass.value)
    credit.value = data.credit
  } catch (e) {
    setError(e)
  } finally {
    loading.value = false
  }
}

async function onLogin() {
  loading.value = true
  error.value = null
  try {
    const data = await login(loginID.value, loginPass.value)
    auth.setUserToken(data.sessionToken)
  } catch (e) {
    setError(e)
  } finally {
    loading.value = false
  }
}

async function onLogout() {
  loading.value = true
  error.value = null
  if (!auth.userToken) {
    loading.value = false
    return
  }
  try {
    await logout(auth.userToken)
    auth.clearUserToken()
    items.value = []
    lastRewards.value = []
    viewMode.value = 'none'
  } catch (e) {
    setError(e)
  } finally {
    loading.value = false
  }
}

async function onInventory() {
  loading.value = true
  error.value = null
  if (!auth.userToken) {
    loading.value = false
    return
  }
  try {
    const data = await inventory(auth.userToken)
    items.value = data.items
    credit.value = data.credit
    viewMode.value = 'inventory'
  } catch (e) {
    setError(e)
  } finally {
    loading.value = false
  }
}

async function onGasha() {
  loading.value = true
  error.value = null
  if (!auth.userToken) {
    loading.value = false
    return
  }
  try {
    const data = await gasha(auth.userToken)
    lastRewards.value = data.rewards
    credit.value = data.remainingCredit
    viewMode.value = 'rewards'
  } catch (e) {
    setError(e)
  } finally {
    loading.value = false
  }
}

async function onGashaTen() {
  loading.value = true
  error.value = null
  if (!auth.userToken) {
    loading.value = false
    return
  }
  try {
    const data = await gashaTen(auth.userToken)
    lastRewards.value = data.rewards
    credit.value = data.remainingCredit
    viewMode.value = 'rewards'
  } catch (e) {
    setError(e)
  } finally {
    loading.value = false
  }
}
</script>
