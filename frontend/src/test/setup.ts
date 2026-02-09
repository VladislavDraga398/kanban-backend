import '@testing-library/jest-dom/vitest'
import { cleanup } from '@testing-library/react'
import { afterEach } from 'vitest'

const storage = (() => {
  const data = new Map<string, string>()

  return {
    get length() {
      return data.size
    },
    clear() {
      data.clear()
    },
    getItem(key: string) {
      return data.has(key) ? data.get(key)! : null
    },
    key(index: number) {
      return Array.from(data.keys())[index] ?? null
    },
    removeItem(key: string) {
      data.delete(key)
    },
    setItem(key: string, value: string) {
      data.set(key, String(value))
    },
  } satisfies Storage
})()

Object.defineProperty(globalThis, 'localStorage', {
  configurable: true,
  value: storage,
})

afterEach(() => {
  cleanup()
  localStorage.clear()
})
