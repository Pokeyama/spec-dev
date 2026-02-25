<template>
  <section class="page-grid">
    <article class="card">
      <h2>Admin Auth</h2>
      <p class="small">学習用途では /admin/regist は未認証</p>
      <input v-model="adminID" placeholder="admin id" />
      <input v-model="adminPass" placeholder="password" type="password" />
      <button @click="onAdminRegist">POST /admin/regist</button>
      <button class="secondary" @click="onAdminLogin">GET /admin/login</button>
      <p class="small">admin token: {{ auth.adminToken ? 'set' : 'none' }}</p>
      <p v-if="error" class="error">{{ error }}</p>
    </article>

    <article class="card">
      <h2>Admin APIs</h2>
      <button @click="onList">GET /account/list</button>
      <input v-model.number="detailID" placeholder="account_id" type="number" />
      <button class="secondary" @click="onDetail">GET /account/detail</button>

      <h3>Accounts</h3>
      <table class="table" v-if="accounts.length > 0">
        <thead>
          <tr>
            <th>ID</th>
            <th>Login ID</th>
            <th>Credit</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="a in accounts" :key="a.account_id">
            <td>{{ a.account_id }}</td>
            <td>{{ a.login_id }}</td>
            <td>{{ a.credit }}</td>
          </tr>
        </tbody>
      </table>

      <h3>Detail</h3>
      <p v-if="detail">account: {{ detail.login_id }} (#{{ detail.account_id }})</p>
      <ul>
        <li v-for="(r, i) in detailRewards" :key="i">{{ r.name }} / {{ r.obtainedAt }}</li>
      </ul>
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
  mapError
} from '../api/client'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()

const adminID = ref('admin01')
const adminPass = ref('adminpass')
const detailID = ref(1)
const error = ref('')
const accounts = ref<Array<{ account_id: number; login_id: string; credit: number }>>([])
const detail = ref<{ account_id: number; login_id: string } | null>(null)
const detailRewards = ref<Array<{ name: string; obtainedAt: string }>>([])

async function onAdminRegist() {
  error.value = ''
  try {
    await adminRegist(adminID.value, adminPass.value)
  } catch (e) {
    error.value = mapError(e).message
  }
}

async function onAdminLogin() {
  error.value = ''
  try {
    const data = await adminLogin(adminID.value, adminPass.value)
    auth.setAdminToken(data.adminSessionToken)
  } catch (e) {
    error.value = mapError(e).message
  }
}

async function onList() {
  error.value = ''
  if (!auth.adminToken) return
  try {
    const data = await accountList(auth.adminToken)
    accounts.value = data.accounts
  } catch (e) {
    error.value = mapError(e).message
  }
}

async function onDetail() {
  error.value = ''
  if (!auth.adminToken) return
  try {
    const data = await accountDetail(auth.adminToken, detailID.value)
    detail.value = { account_id: data.account_id, login_id: data.login_id }
    detailRewards.value = data.rewards
  } catch (e) {
    error.value = mapError(e).message
  }
}
</script>
