import axios from 'axios'
import { getAuthToken } from '../auth/token'

const baseURL = import.meta.env.VITE_API_URL || '/api/v1'

export const apiClient = axios.create({
  baseURL,
  headers: {
    'Content-Type': 'application/json',
  },
})

apiClient.interceptors.request.use((config) => {
  const token = getAuthToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})
