import { describe, expect, it } from 'vitest'
import styles from './styles.css?raw'

describe('styles', () => {
  it('defines responsive breakpoint for desktop layout', () => {
    expect(styles).toContain('@media (min-width: 900px)')
    expect(styles).toContain('grid-template-columns: 1fr 1fr')
  })
})
