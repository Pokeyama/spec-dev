import { beforeEach, describe, expect, it } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useAuthStore } from './auth'

describe('auth store', () => {
  beforeEach(() => {
    localStorage.clear()
    setActivePinia(createPinia())
  })

  it('sets and clears user token', () => {
    const store = useAuthStore()

    store.setUserToken('user-token')
    expect(store.userToken).toBe('user-token')
    expect(localStorage.getItem('gasha_user_token')).toBe('user-token')

    store.clearUserToken()
    expect(store.userToken).toBe('')
    expect(localStorage.getItem('gasha_user_token')).toBeNull()
  })

  it('sets and clears admin token', () => {
    const store = useAuthStore()

    store.setAdminToken('admin-token')
    expect(store.adminToken).toBe('admin-token')
    expect(localStorage.getItem('gasha_admin_token')).toBe('admin-token')

    store.clearAdminToken()
    expect(store.adminToken).toBe('')
    expect(localStorage.getItem('gasha_admin_token')).toBeNull()
  })
})
