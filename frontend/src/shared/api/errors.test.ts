import { describe, expect, it } from 'vitest'
import { getErrorMessage } from './errors'

describe('getErrorMessage', () => {
  it('returns api error message from axios response body', () => {
    const error = {
      isAxiosError: true,
      message: 'Request failed',
      response: {
        data: {
          error: 'invalid credentials',
        },
      },
    }

    expect(getErrorMessage(error)).toBe('invalid credentials')
  })

  it('falls back to axios message when response body does not contain error', () => {
    const error = {
      isAxiosError: true,
      message: 'Network Error',
      response: {
        data: {},
      },
    }

    expect(getErrorMessage(error)).toBe('Network Error')
  })

  it('returns native error message', () => {
    const error = new Error('native error')
    expect(getErrorMessage(error)).toBe('native error')
  })

  it('returns unknown fallback for non-error values', () => {
    expect(getErrorMessage('fail')).toBe('Unknown error')
  })
})
