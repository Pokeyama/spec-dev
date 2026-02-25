<template>
  <section class="page-grid">
    <article class="card">
      <h2>User Auth</h2>
      <p class="small">登録とログイン</p>
      <input v-model="registID" placeholder="id" />
      <input v-model="registPass" placeholder="password" type="password" />
      <button @click="onRegist">POST /regist</button>

      <input v-model="loginID" placeholder="id" />
      <input v-model="loginPass" placeholder="password" type="password" />
      <button class="secondary" @click="onLogin">GET /llogin</button>

      <p class="small">token: {{ auth.userToken ? 'set' : 'none' }}</p>
      <p v-if="error" class="error">{{ error }}</p>
    </article>

    <article class="card">
      <h2>User Actions</h2>
      <p class="small">ガシャと所持一覧</p>
      <div class="button-row">
        <button @click="onGasha">POST /gasha</button>
        <button @click="onGashaTen">POST /gasha/ten</button>
        <button class="warn" @click="onLogout">GET /logout</button>
      </div>
      <button class="secondary" @click="onInventory">GET /inventory</button>

      <p>Credit: {{ credit }}</p>
      <h3>Last Rewards</h3>
      <ul>
        <li v-for="(r, i) in lastRewards" :key="i">{{ r.name }}</li>
      </ul>

      <h3>Inventory</h3>
      <ul>
        <li v-for="item in items" :key="item.name">{{ item.name }} x{{ item.count }}</li>
      </ul>
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
  regist
} from '../api/client'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()

const registID = ref('alice')
const registPass = ref('pass1234')
const loginID = ref('alice')
const loginPass = ref('pass1234')

const error = ref('')
const credit = ref(0)
const lastRewards = ref<Array<{ name: string }>>([])
const items = ref<Array<{ name: string; count: number }>>([])

async function onRegist() {
  error.value = ''
  try {
    const data = await regist(registID.value, registPass.value)
    credit.value = data.credit
  } catch (e) {
    error.value = mapError(e).message
  }
}

async function onLogin() {
  error.value = ''
  try {
    const data = await login(loginID.value, loginPass.value)
    auth.setUserToken(data.sessionToken)
  } catch (e) {
    error.value = mapError(e).message
  }
}

async function onLogout() {
  error.value = ''
  if (!auth.userToken) return
  try {
    await logout(auth.userToken)
    auth.clearUserToken()
    items.value = []
    lastRewards.value = []
  } catch (e) {
    error.value = mapError(e).message
  }
}

async function onInventory() {
  error.value = ''
  if (!auth.userToken) return
  try {
    const data = await inventory(auth.userToken)
    items.value = data.items
    credit.value = data.credit
  } catch (e) {
    error.value = mapError(e).message
  }
}

async function onGasha() {
  error.value = ''
  if (!auth.userToken) return
  try {
    const data = await gasha(auth.userToken)
    lastRewards.value = data.rewards
    credit.value = data.remainingCredit
  } catch (e) {
    error.value = mapError(e).message
  }
}

async function onGashaTen() {
  error.value = ''
  if (!auth.userToken) return
  try {
    const data = await gashaTen(auth.userToken)
    lastRewards.value = data.rewards
    credit.value = data.remainingCredit
  } catch (e) {
    error.value = mapError(e).message
  }
}
</script>
