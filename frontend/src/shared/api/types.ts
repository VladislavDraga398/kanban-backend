export type ApiErrorBody = {
  error: string
}

export type AuthResponse = {
  id: string
  email: string
  token: string
}

export type Board = {
  id: string
  owner_id: string
  name: string
  created_at: string
  updated_at: string
}

export type Column = {
  id: string
  board_id: string
  name: string
  position: number
  created_at: string
  updated_at: string
}

export type TaskPriority = 'low' | 'normal' | 'high'

export type Task = {
  id: string
  board_id: string
  column_id: string
  title: string
  description: string
  priority: TaskPriority
  labels: string[]
  due_date: string | null
  position: number
  created_at: string
  updated_at: string
}
