import { apiClient } from '../../shared/api/client'
import type { Board } from '../../shared/api/types'

type CreateBoardPayload = {
  name: string
}

type UpdateBoardPayload = {
  id: string
  name: string
}

export async function listBoards(): Promise<Board[]> {
  const { data } = await apiClient.get<Board[]>('/boards')
  return data
}

export async function getBoard(id: string): Promise<Board> {
  const { data } = await apiClient.get<Board>(`/boards/${id}`)
  return data
}

export async function createBoard(payload: CreateBoardPayload): Promise<Board> {
  const { data } = await apiClient.post<Board>('/boards', payload)
  return data
}

export async function updateBoard(payload: UpdateBoardPayload): Promise<Board> {
  const { data } = await apiClient.put<Board>(`/boards/${payload.id}`, {
    name: payload.name,
  })
  return data
}

export async function deleteBoard(id: string): Promise<void> {
  await apiClient.delete(`/boards/${id}`)
}
