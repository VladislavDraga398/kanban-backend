import { apiClient } from '../../shared/api/client'
import type { Column, Task } from '../../shared/api/types'

type CreateColumnPayload = {
  boardId: string
  name: string
}

type UpdateColumnPayload = {
  boardId: string
  columnId: string
  name: string
}

type DeleteColumnPayload = {
  boardId: string
  columnId: string
}

type CreateTaskPayload = {
  boardId: string
  columnId: string
  title: string
  description: string
}

type UpdateTaskPayload = {
  boardId: string
  columnId: string
  taskId: string
  title: string
  description: string
}

type DeleteTaskPayload = {
  boardId: string
  columnId: string
  taskId: string
}

type MoveTaskPayload = {
  boardId: string
  taskId: string
  columnId: string
}

export async function listColumns(boardId: string): Promise<Column[]> {
  const { data } = await apiClient.get<Column[]>(`/boards/${boardId}/columns`)
  return data
}

export async function createColumn(payload: CreateColumnPayload): Promise<Column> {
  const { data } = await apiClient.post<Column>(`/boards/${payload.boardId}/columns`, {
    name: payload.name,
  })
  return data
}

export async function updateColumn(payload: UpdateColumnPayload): Promise<Column> {
  const { data } = await apiClient.put<Column>(
    `/boards/${payload.boardId}/columns/${payload.columnId}`,
    {
      name: payload.name,
    },
  )
  return data
}

export async function deleteColumn(payload: DeleteColumnPayload): Promise<void> {
  await apiClient.delete(`/boards/${payload.boardId}/columns/${payload.columnId}`)
}

export async function listTasks(boardId: string, columnId: string): Promise<Task[]> {
  const { data } = await apiClient.get<Task[]>(
    `/boards/${boardId}/columns/${columnId}/tasks`,
  )
  return data
}

export async function createTask(payload: CreateTaskPayload): Promise<Task> {
  const { data } = await apiClient.post<Task>(
    `/boards/${payload.boardId}/columns/${payload.columnId}/tasks`,
    {
      title: payload.title,
      description: payload.description,
    },
  )
  return data
}

export async function updateTask(payload: UpdateTaskPayload): Promise<Task> {
  const { data } = await apiClient.put<Task>(
    `/boards/${payload.boardId}/columns/${payload.columnId}/tasks/${payload.taskId}`,
    {
      title: payload.title,
      description: payload.description,
    },
  )
  return data
}

export async function deleteTask(payload: DeleteTaskPayload): Promise<void> {
  await apiClient.delete(
    `/boards/${payload.boardId}/columns/${payload.columnId}/tasks/${payload.taskId}`,
  )
}

export async function moveTask(payload: MoveTaskPayload): Promise<Task> {
  const { data } = await apiClient.patch<Task>(
    `/boards/${payload.boardId}/tasks/${payload.taskId}/move`,
    {
      column_id: payload.columnId,
    },
  )
  return data
}
