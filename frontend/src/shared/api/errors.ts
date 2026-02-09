import axios from 'axios'
import type { ApiErrorBody } from './types'

export function getErrorMessage(error: unknown): string {
  if (axios.isAxiosError<ApiErrorBody>(error)) {
    return error.response?.data?.error || error.message || 'Request failed'
  }
  if (error instanceof Error) {
    return error.message
  }
  return 'Unknown error'
}
