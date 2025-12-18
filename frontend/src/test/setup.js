import '@testing-library/jest-dom'
import { expect, afterEach } from 'vitest'
import { cleanup } from '@testing-library/react'

// Cleanup after each test
afterEach(() => {
  cleanup()
})

// Mock fetch globally for tests
global.fetch = vi.fn()

// Reset mocks after each test
afterEach(() => {
  vi.resetAllMocks()
})
