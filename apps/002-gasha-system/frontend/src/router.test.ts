import { describe, expect, it } from 'vitest'
import { router } from './router'

describe('router', () => {
  it('has user and admin routes', () => {
    expect(router.resolve('/user').matched.length).toBeGreaterThan(0)
    expect(router.resolve('/admin').matched.length).toBeGreaterThan(0)
  })

  it('redirects root to /user', async () => {
    await router.push('/')
    await router.isReady()
    expect(router.currentRoute.value.fullPath).toBe('/user')
  })
})
