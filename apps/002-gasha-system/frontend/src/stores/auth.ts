import { defineStore } from 'pinia'

const USER_KEY = 'gasha_user_token'
const ADMIN_KEY = 'gasha_admin_token'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    userToken: localStorage.getItem(USER_KEY) || '',
    adminToken: localStorage.getItem(ADMIN_KEY) || ''
  }),
  actions: {
    setUserToken(token: string) {
      this.userToken = token
      localStorage.setItem(USER_KEY, token)
    },
    clearUserToken() {
      this.userToken = ''
      localStorage.removeItem(USER_KEY)
    },
    setAdminToken(token: string) {
      this.adminToken = token
      localStorage.setItem(ADMIN_KEY, token)
    },
    clearAdminToken() {
      this.adminToken = ''
      localStorage.removeItem(ADMIN_KEY)
    }
  }
})
