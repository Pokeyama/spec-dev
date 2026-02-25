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
})
