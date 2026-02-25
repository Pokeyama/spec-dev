import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'

import AdminView from './AdminView.vue'
import { useAuthStore } from '../stores/auth'
import * as api from '../api/client'

vi.mock('../api/client', () => ({
  adminRegist: vi.fn(),
  adminLogin: vi.fn(),
  accountList: vi.fn(),
  accountDetail: vi.fn(),
  mapError: (error: any) => error?.response?.data?.error || { code: 'UNKNOWN', message: String(error) }
}))

function mountView() {
  const pinia = createPinia()
  setActivePinia(pinia)
  const wrapper = mount(AdminView, {
    global: {
      plugins: [pinia]
    }
  })
  return { wrapper, auth: useAuthStore() }
}

function clickButtonByText(wrapper: ReturnType<typeof mount>, label: string) {
  const button = wrapper
    .findAll('button')
    .find((b) => b.text().trim() === label)

  if (!button) {
    throw new Error(`button not found: ${label}`)
  }

  return button.trigger('click')
}

describe('AdminView', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.clearAllMocks()
  })

  it('does not call account list API when admin token is missing', async () => {
    const { wrapper, auth } = mountView()
    expect(auth.adminToken).toBe('')

    await clickButtonByText(wrapper, 'GET /account/list')

    expect(vi.mocked(api.accountList)).not.toHaveBeenCalled()
  })

  it('shows API error message (403) on account list failure', async () => {
    const { wrapper, auth } = mountView()
    auth.setAdminToken('admin-token')

    vi.mocked(api.accountList).mockRejectedValue({
      response: {
        data: {
          error: {
            code: 'FORBIDDEN',
            message: 'admin role required'
          }
        }
      }
    })

    await clickButtonByText(wrapper, 'GET /account/list')
    await flushPromises()

    expect(vi.mocked(api.accountList)).toHaveBeenCalledWith('admin-token')
    expect(wrapper.text()).toContain('admin role required')
  })

  it('opens detail from account list row button and hides list panel', async () => {
    const { wrapper, auth } = mountView()
    auth.setAdminToken('admin-token')

    vi.mocked(api.accountList).mockResolvedValue({
      accounts: [
        { account_id: 1, login_id: 'alice', credit: 990, createdAt: '2026-02-25T00:00:00Z' }
      ]
    })
    vi.mocked(api.accountDetail).mockResolvedValue({
      account_id: 1,
      login_id: 'alice',
      rewards: [{ name: 'Pikachu', obtainedAt: '2026-02-25T00:00:00Z' }]
    })

    await clickButtonByText(wrapper, 'GET /account/list')
    await flushPromises()
    expect(wrapper.find('[data-testid="admin-list-panel"]').exists()).toBe(true)

    await wrapper.find('[data-testid="detail-btn-1"]').trigger('click')
    await flushPromises()

    expect(vi.mocked(api.accountDetail)).toHaveBeenCalledWith('admin-token', 1)
    expect(wrapper.find('[data-testid="admin-list-panel"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="admin-detail-panel"]').exists()).toBe(true)
  })
})
