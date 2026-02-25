import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'

import UserView from './UserView.vue'
import { useAuthStore } from '../stores/auth'
import * as api from '../api/client'

vi.mock('../api/client', () => ({
  regist: vi.fn(),
  login: vi.fn(),
  logout: vi.fn(),
  inventory: vi.fn(),
  gasha: vi.fn(),
  gashaTen: vi.fn(),
  mapError: (error: any) => error?.response?.data?.error || { code: 'UNKNOWN', message: String(error) }
}))

function mountView() {
  const pinia = createPinia()
  setActivePinia(pinia)
  const wrapper = mount(UserView, {
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

describe('UserView', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.clearAllMocks()
  })

  it('does not call gasha API when session token is missing', async () => {
    const { wrapper, auth } = mountView()
    expect(auth.userToken).toBe('')

    await clickButtonByText(wrapper, 'POST /gasha')

    expect(vi.mocked(api.gasha)).not.toHaveBeenCalled()
  })

  it('shows API error message (402) on gasha failure', async () => {
    const { wrapper, auth } = mountView()
    auth.setUserToken('user-token')

    vi.mocked(api.gasha).mockRejectedValue({
      response: {
        data: {
          error: {
            code: 'INSUFFICIENT_DIAMONDS',
            message: 'insufficient diamonds'
          }
        }
      }
    })

    await clickButtonByText(wrapper, 'POST /gasha')
    await flushPromises()

    expect(vi.mocked(api.gasha)).toHaveBeenCalledWith('user-token')
    expect(wrapper.text()).toContain('insufficient diamonds')
  })
})
