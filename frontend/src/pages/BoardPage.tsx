import { DndContext, type DragEndEvent, useDroppable } from '@dnd-kit/core'
import {
  useMutation,
  useQueries,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query'
import { useMemo, useState, type FormEvent } from 'react'
import { Link, useParams } from 'react-router-dom'
import { DraggableTaskCard } from '../components/DraggableTaskCard'
import {
  createColumn,
  createTask,
  deleteColumn,
  deleteTask,
  listColumns,
  listTasks,
  moveTask,
  updateColumn,
  updateTask,
} from '../features/board/api'
import { getBoard } from '../features/boards/api'
import { getErrorMessage } from '../shared/api/errors'
import type { Column, Task } from '../shared/api/types'

type TaskDraft = {
  title: string
  description: string
}

type ColumnCardProps = {
  boardId: string
  column: Column
  tasks: Task[]
  taskLoading: boolean
  draft: TaskDraft
  busy: boolean
  onDraftChange: (columnId: string, patch: Partial<TaskDraft>) => void
  onCreateTask: (columnId: string) => void
  onEditColumn: (column: Column) => void
  onDeleteColumn: (column: Column) => void
  onEditTask: (task: Task) => void
  onDeleteTask: (task: Task) => void
}

function ColumnCard({
  boardId,
  column,
  tasks,
  taskLoading,
  draft,
  busy,
  onDraftChange,
  onCreateTask,
  onEditColumn,
  onDeleteColumn,
  onEditTask,
  onDeleteTask,
}: ColumnCardProps) {
  const { setNodeRef, isOver } = useDroppable({ id: column.id })

  return (
    <section className={`column-card ${isOver ? 'is-over' : ''}`} ref={setNodeRef}>
      <header className="column-card__header">
        <div>
          <h3>{column.name}</h3>
          <p>{tasks.length} задач</p>
        </div>
        <div className="column-card__header-actions">
          <button type="button" className="ghost-button" onClick={() => onEditColumn(column)}>
            Имя
          </button>
          <button
            type="button"
            className="danger-button"
            onClick={() => onDeleteColumn(column)}
            disabled={busy}
          >
            Удалить
          </button>
        </div>
      </header>

      <form
        className="task-create-form"
        onSubmit={(event) => {
          event.preventDefault()
          onCreateTask(column.id)
        }}
      >
        <input
          type="text"
          placeholder="Новая задача"
          value={draft.title}
          onChange={(event) => onDraftChange(column.id, { title: event.target.value })}
        />
        <textarea
          placeholder="Описание (опционально)"
          value={draft.description}
          onChange={(event) => onDraftChange(column.id, { description: event.target.value })}
          rows={2}
        />
        <button type="submit" className="primary-button" disabled={busy}>
          Добавить задачу
        </button>
      </form>

      <div className="task-list">
        {taskLoading && <p className="hint-text">Загружаю задачи...</p>}
        {!taskLoading && tasks.length === 0 && <p className="hint-text">Пусто</p>}
        {tasks.map((task) => (
          <DraggableTaskCard
            key={task.id}
            task={task}
            pending={busy}
            onEdit={onEditTask}
            onDelete={onDeleteTask}
          />
        ))}
      </div>

      <p className="column-id">#{boardId.slice(0, 4)} · {column.id.slice(0, 6)}</p>
    </section>
  )
}

export function BoardPage() {
  const { boardId = '' } = useParams<{ boardId: string }>()
  const queryClient = useQueryClient()

  const [newColumnName, setNewColumnName] = useState('')
  const [taskDrafts, setTaskDrafts] = useState<Record<string, TaskDraft>>({})
  const [editingColumn, setEditingColumn] = useState<Column | null>(null)
  const [editingColumnName, setEditingColumnName] = useState('')
  const [editingTask, setEditingTask] = useState<Task | null>(null)
  const [editingTaskTitle, setEditingTaskTitle] = useState('')
  const [editingTaskDescription, setEditingTaskDescription] = useState('')
  const [actionError, setActionError] = useState<string | null>(null)

  const boardQuery = useQuery({
    queryKey: ['board', boardId, 'meta'],
    queryFn: () => getBoard(boardId),
    enabled: Boolean(boardId),
  })

  const columnsQuery = useQuery({
    queryKey: ['board', boardId, 'columns'],
    queryFn: () => listColumns(boardId),
    enabled: Boolean(boardId),
  })

  const columns = useMemo(() => columnsQuery.data ?? [], [columnsQuery.data])

  const taskQueries = useQueries({
    queries: columns.map((column) => ({
      queryKey: ['board', boardId, 'tasks', column.id],
      queryFn: () => listTasks(boardId, column.id),
      enabled: Boolean(boardId),
    })),
  })

  const tasksByColumn = useMemo(() => {
    const map = new Map<string, Task[]>()
    columns.forEach((column, index) => {
      map.set(column.id, taskQueries[index]?.data ?? [])
    })
    return map
  }, [columns, taskQueries])

  const createColumnMutation = useMutation({
    mutationFn: createColumn,
    onSuccess: () => {
      setNewColumnName('')
      setActionError(null)
      queryClient.invalidateQueries({ queryKey: ['board', boardId, 'columns'] })
    },
    onError: (error) => setActionError(getErrorMessage(error)),
  })

  const updateColumnMutation = useMutation({
    mutationFn: updateColumn,
    onSuccess: () => {
      setEditingColumn(null)
      setEditingColumnName('')
      setActionError(null)
      queryClient.invalidateQueries({ queryKey: ['board', boardId, 'columns'] })
    },
    onError: (error) => setActionError(getErrorMessage(error)),
  })

  const deleteColumnMutation = useMutation({
    mutationFn: deleteColumn,
    onSuccess: () => {
      setActionError(null)
      queryClient.invalidateQueries({ queryKey: ['board', boardId, 'columns'] })
      queryClient.invalidateQueries({ queryKey: ['board', boardId, 'tasks'] })
    },
    onError: (error) => setActionError(getErrorMessage(error)),
  })

  const createTaskMutation = useMutation({
    mutationFn: createTask,
    onSuccess: (task) => {
      setActionError(null)
      setTaskDrafts((current) => ({
        ...current,
        [task.column_id]: { title: '', description: '' },
      }))
      queryClient.invalidateQueries({
        queryKey: ['board', boardId, 'tasks', task.column_id],
      })
    },
    onError: (error) => setActionError(getErrorMessage(error)),
  })

  const updateTaskMutation = useMutation({
    mutationFn: updateTask,
    onSuccess: (task) => {
      setActionError(null)
      setEditingTask(null)
      setEditingTaskTitle('')
      setEditingTaskDescription('')
      queryClient.invalidateQueries({
        queryKey: ['board', boardId, 'tasks', task.column_id],
      })
    },
    onError: (error) => setActionError(getErrorMessage(error)),
  })

  const deleteTaskMutation = useMutation({
    mutationFn: deleteTask,
    onSuccess: () => {
      setActionError(null)
      queryClient.invalidateQueries({ queryKey: ['board', boardId, 'tasks'] })
    },
    onError: (error) => setActionError(getErrorMessage(error)),
  })

  const moveTaskMutation = useMutation({
    mutationFn: moveTask,
    onSuccess: () => {
      setActionError(null)
      queryClient.invalidateQueries({ queryKey: ['board', boardId, 'tasks'] })
    },
    onError: (error) => setActionError(getErrorMessage(error)),
  })

  function updateDraft(columnId: string, patch: Partial<TaskDraft>) {
    setTaskDrafts((current) => ({
      ...current,
      [columnId]: {
        title: current[columnId]?.title ?? '',
        description: current[columnId]?.description ?? '',
        ...patch,
      },
    }))
  }

  function submitCreateColumn(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const name = newColumnName.trim()
    if (!name || !boardId) {
      return
    }
    createColumnMutation.mutate({ boardId, name })
  }

  function submitCreateTask(columnId: string) {
    const draft = taskDrafts[columnId] ?? { title: '', description: '' }
    const title = draft.title.trim()
    if (!title || !boardId) {
      return
    }
    createTaskMutation.mutate({
      boardId,
      columnId,
      title,
      description: draft.description.trim(),
    })
  }

  function openColumnEdit(column: Column) {
    setEditingColumn(column)
    setEditingColumnName(column.name)
  }

  function submitColumnEdit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!editingColumn || !boardId) {
      return
    }
    const name = editingColumnName.trim()
    if (!name) {
      return
    }
    updateColumnMutation.mutate({
      boardId,
      columnId: editingColumn.id,
      name,
    })
  }

  function openTaskEdit(task: Task) {
    setEditingTask(task)
    setEditingTaskTitle(task.title)
    setEditingTaskDescription(task.description)
  }

  function submitTaskEdit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!editingTask || !boardId) {
      return
    }
    const title = editingTaskTitle.trim()
    if (!title) {
      return
    }
    updateTaskMutation.mutate({
      boardId,
      columnId: editingTask.column_id,
      taskId: editingTask.id,
      title,
      description: editingTaskDescription.trim(),
    })
  }

  function onDragEnd(event: DragEndEvent) {
    const { active, over } = event
    const fromColumnId = active.data.current?.columnId as string | undefined
    const targetColumnId = over?.id ? String(over.id) : undefined
    if (!boardId || !fromColumnId || !targetColumnId || fromColumnId === targetColumnId) {
      return
    }
    moveTaskMutation.mutate({
      boardId,
      taskId: String(active.id),
      columnId: targetColumnId,
    })
  }

  const isBusy =
    createColumnMutation.isPending ||
    updateColumnMutation.isPending ||
    deleteColumnMutation.isPending ||
    createTaskMutation.isPending ||
    updateTaskMutation.isPending ||
    deleteTaskMutation.isPending ||
    moveTaskMutation.isPending

  if (!boardId) {
    return (
      <main className="page-shell">
        <p className="error-text panel">Board id отсутствует в URL.</p>
      </main>
    )
  }

  return (
    <main className="page-shell">
      <header className="topbar">
        <div>
          <p className="badge">BOARD VIEW</p>
          <h1>{boardQuery.data?.name || 'Загрузка доски...'}</h1>
        </div>
        <Link className="ghost-button" to="/boards">
          Назад к доскам
        </Link>
      </header>

      <section className="panel">
        <h2>Колонки</h2>
        <form className="inline-form" onSubmit={submitCreateColumn}>
          <input
            type="text"
            placeholder="Название колонки"
            value={newColumnName}
            onChange={(event) => setNewColumnName(event.target.value)}
            required
          />
          <button
            type="submit"
            className="primary-button"
            disabled={createColumnMutation.isPending}
          >
            Добавить колонку
          </button>
        </form>
      </section>

      {actionError && <p className="error-text panel">{actionError}</p>}
      {boardQuery.isError && <p className="error-text panel">{getErrorMessage(boardQuery.error)}</p>}
      {columnsQuery.isError && (
        <p className="error-text panel">{getErrorMessage(columnsQuery.error)}</p>
      )}

      <DndContext onDragEnd={onDragEnd}>
        <section className="kanban-grid">
          {columnsQuery.isLoading && <p className="panel">Загружаю колонки...</p>}
          {!columnsQuery.isLoading && columns.length === 0 && (
            <p className="panel">Пока нет колонок. Добавь первую.</p>
          )}
          {columns.map((column, index) => (
            <ColumnCard
              key={column.id}
              boardId={boardId}
              column={column}
              tasks={tasksByColumn.get(column.id) ?? []}
              taskLoading={Boolean(taskQueries[index]?.isLoading)}
              draft={taskDrafts[column.id] ?? { title: '', description: '' }}
              busy={isBusy}
              onDraftChange={updateDraft}
              onCreateTask={submitCreateTask}
              onEditColumn={openColumnEdit}
              onDeleteColumn={(target) =>
                deleteColumnMutation.mutate({
                  boardId,
                  columnId: target.id,
                })
              }
              onEditTask={openTaskEdit}
              onDeleteTask={(task) =>
                deleteTaskMutation.mutate({
                  boardId,
                  columnId: task.column_id,
                  taskId: task.id,
                })
              }
            />
          ))}
        </section>
      </DndContext>

      {editingColumn && (
        <div className="modal-backdrop" onClick={() => setEditingColumn(null)}>
          <section
            className="modal"
            onClick={(event) => event.stopPropagation()}
            role="dialog"
            aria-modal="true"
          >
            <h3>Переименовать колонку</h3>
            <form className="modal-form" onSubmit={submitColumnEdit}>
              <input
                type="text"
                value={editingColumnName}
                onChange={(event) => setEditingColumnName(event.target.value)}
                required
              />
              <div className="modal-actions">
                <button
                  type="button"
                  className="ghost-button"
                  onClick={() => setEditingColumn(null)}
                >
                  Отмена
                </button>
                <button type="submit" className="primary-button" disabled={isBusy}>
                  Сохранить
                </button>
              </div>
            </form>
          </section>
        </div>
      )}

      {editingTask && (
        <div className="modal-backdrop" onClick={() => setEditingTask(null)}>
          <section
            className="modal"
            onClick={(event) => event.stopPropagation()}
            role="dialog"
            aria-modal="true"
          >
            <h3>Редактировать задачу</h3>
            <form className="modal-form" onSubmit={submitTaskEdit}>
              <input
                type="text"
                value={editingTaskTitle}
                onChange={(event) => setEditingTaskTitle(event.target.value)}
                required
              />
              <textarea
                rows={4}
                value={editingTaskDescription}
                onChange={(event) => setEditingTaskDescription(event.target.value)}
              />
              <div className="modal-actions">
                <button
                  type="button"
                  className="ghost-button"
                  onClick={() => setEditingTask(null)}
                >
                  Отмена
                </button>
                <button type="submit" className="primary-button" disabled={isBusy}>
                  Сохранить
                </button>
              </div>
            </form>
          </section>
        </div>
      )}
    </main>
  )
}
