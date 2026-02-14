import { apiClient } from '../../shared/api/client'
import type { AuthResponse } from '../../shared/api/types'

type AuthPayload = {
  email: string
  password: string
}

export async function register(payload: AuthPayload): Promise<AuthResponse> {
  const { data } = await apiClient.post<AuthResponse>('/auth/register', payload)
  return data
}

export async function login(payload: AuthPayload): Promise<AuthResponse> {
  const { data } = await apiClient.post<AuthResponse>('/auth/login', payload)
  return data
}
