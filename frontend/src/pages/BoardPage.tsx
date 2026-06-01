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
import type { Column, Task, TaskPriority } from '../shared/api/types'

type TaskDraft = {
  title: string
  description: string
  priority: TaskPriority
  labels: string
  dueDate: string
}

const emptyTaskDraft: TaskDraft = {
  title: '',
  description: '',
  priority: 'normal',
  labels: '',
  dueDate: '',
}

function labelsToText(labels: string[]) {
  return labels.join(', ')
}

function parseLabels(value: string) {
  return value
    .split(',')
    .map((label) => label.trim())
    .filter(Boolean)
    .slice(0, 5)
}

type ColumnCardProps = {
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

function formatBoardDate(value?: string) {
  if (!value) {
    return '...'
  }
  return new Date(value).toLocaleDateString('ru-RU', {
    day: '2-digit',
    month: 'long',
  })
}

function ColumnSkeleton() {
  return (
    <section className="column-card column-card--skeleton" aria-label="Загрузка колонки">
      <div className="skeleton-line skeleton-line--wide" />
      <div className="skeleton-line" />
      <div className="skeleton-card" />
      <div className="skeleton-card skeleton-card--short" />
    </section>
  )
}

function ColumnCard({
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
  const hasTasks = tasks.length > 0

  return (
    <section
      className={`column-card ${isOver ? 'is-over' : ''} ${hasTasks ? '' : 'is-empty'}`}
      ref={setNodeRef}
    >
      <header className="column-card__header">
        <div className="column-card__title">
          <h3>{column.name}</h3>
          <span>{tasks.length} задач</span>
        </div>
        <div className="column-card__header-actions">
          <button
            type="button"
            className="ghost-button"
            onClick={() => onEditColumn(column)}
            title="Переименовать колонку"
          >
            Править
          </button>
          <button
            type="button"
            className="danger-button"
            onClick={() => onDeleteColumn(column)}
            disabled={busy}
            title="Удалить колонку"
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
        <div className="task-create-form__top">
          <span>Быстрая задача</span>
          <span>{draft.title.trim().length}/120</span>
        </div>
        <input
          type="text"
          placeholder="Новая задача"
          value={draft.title}
          maxLength={120}
          onChange={(event) => onDraftChange(column.id, { title: event.target.value })}
        />
        <textarea
          placeholder="Описание (опционально)"
          value={draft.description}
          onChange={(event) => onDraftChange(column.id, { description: event.target.value })}
          rows={2}
        />
        <div className="task-create-form__meta">
          <select
            value={draft.priority}
            onChange={(event) =>
              onDraftChange(column.id, { priority: event.target.value as TaskPriority })
            }
            aria-label="Приоритет задачи"
          >
            <option value="low">Низкий</option>
            <option value="normal">Обычный</option>
            <option value="high">Высокий</option>
          </select>
          <input
            type="date"
            value={draft.dueDate}
            onChange={(event) => onDraftChange(column.id, { dueDate: event.target.value })}
            aria-label="Срок задачи"
          />
        </div>
        <input
          type="text"
          placeholder="Метки через запятую"
          value={draft.labels}
          onChange={(event) => onDraftChange(column.id, { labels: event.target.value })}
        />
        <button type="submit" className="primary-button" disabled={busy}>
          Добавить задачу
        </button>
      </form>

      <div className="task-list">
        {taskLoading && (
          <div className="task-list__loading" aria-label="Загружаю задачи">
            <div className="skeleton-card skeleton-card--short" />
            <div className="skeleton-card" />
          </div>
        )}
        {!taskLoading && tasks.length === 0 && (
          <div className="empty-column">
            <strong>Пока пусто</strong>
            <span>Создай первую карточку в этом потоке.</span>
          </div>
        )}
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

      <footer className="column-card__footer">
        <span>Обновлена {formatBoardDate(column.updated_at)}</span>
        <span>#{column.id.slice(0, 6)}</span>
      </footer>
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
  const [editingTaskPriority, setEditingTaskPriority] = useState<TaskPriority>('normal')
  const [editingTaskLabels, setEditingTaskLabels] = useState('')
  const [editingTaskDueDate, setEditingTaskDueDate] = useState('')
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

  const totalTasks = useMemo(
    () => Array.from(tasksByColumn.values()).reduce((total, tasks) => total + tasks.length, 0),
    [tasksByColumn],
  )

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
        [task.column_id]: emptyTaskDraft,
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
      setEditingTaskPriority('normal')
      setEditingTaskLabels('')
      setEditingTaskDueDate('')
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
        ...(current[columnId] ?? emptyTaskDraft),
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
      priority: draft.priority,
      labels: parseLabels(draft.labels),
      dueDate: draft.dueDate || undefined,
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
    setEditingTaskPriority(task.priority)
    setEditingTaskLabels(labelsToText(task.labels))
    setEditingTaskDueDate(task.due_date ?? '')
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
      priority: editingTaskPriority,
      labels: parseLabels(editingTaskLabels),
      dueDate: editingTaskDueDate,
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
    <main className="page-shell board-page">
      <header className="board-hero">
        <div className="board-hero__content">
          <Link className="back-link" to="/boards">
            Назад к доскам
          </Link>
          <p className="badge">PREMIUM BOARD</p>
          <h1>{boardQuery.data?.name || 'Загрузка доски...'}</h1>
          <p className="board-hero__subtitle">
            Обновлена {formatBoardDate(boardQuery.data?.updated_at)}
          </p>
          <div className="board-hero__stats" aria-label="Статистика доски">
            <div>
              <strong>{columns.length}</strong>
              <span>колонки</span>
            </div>
            <div>
              <strong>{totalTasks}</strong>
              <span>задачи</span>
            </div>
            <div>
              <strong>{isBusy ? 'Идёт' : 'Готово'}</strong>
              <span>статус</span>
            </div>
          </div>
        </div>
        <form className="board-quick-add" onSubmit={submitCreateColumn}>
          <label htmlFor="new-column-name">Новая колонка</label>
          <input
            id="new-column-name"
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
      </header>

      {actionError && <p className="error-text panel">{actionError}</p>}
      {boardQuery.isError && <p className="error-text panel">{getErrorMessage(boardQuery.error)}</p>}
      {columnsQuery.isError && (
        <p className="error-text panel">{getErrorMessage(columnsQuery.error)}</p>
      )}

      <DndContext onDragEnd={onDragEnd}>
        <section className="kanban-grid">
          {columnsQuery.isLoading && (
            <>
              <ColumnSkeleton />
              <ColumnSkeleton />
              <ColumnSkeleton />
            </>
          )}
          {!columnsQuery.isLoading && columns.length === 0 && (
            <section className="board-empty-state">
              <p className="badge">EMPTY BOARD</p>
              <h2>Здесь пока чистый лист</h2>
              <p>Добавь первую колонку в панели выше.</p>
            </section>
          )}
          {columns.map((column, index) => (
            <ColumnCard
              key={column.id}
              column={column}
              tasks={tasksByColumn.get(column.id) ?? []}
              taskLoading={Boolean(taskQueries[index]?.isLoading)}
              draft={taskDrafts[column.id] ?? emptyTaskDraft}
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
              <div className="modal-form__meta">
                <label>
                  <span>Приоритет</span>
                  <select
                    value={editingTaskPriority}
                    onChange={(event) => setEditingTaskPriority(event.target.value as TaskPriority)}
                  >
                    <option value="low">Низкий</option>
                    <option value="normal">Обычный</option>
                    <option value="high">Высокий</option>
                  </select>
                </label>
                <label>
                  <span>Срок</span>
                  <input
                    type="date"
                    value={editingTaskDueDate}
                    onChange={(event) => setEditingTaskDueDate(event.target.value)}
                  />
                </label>
              </div>
              <label className="modal-form__label">
                <span>Метки</span>
                <input
                  type="text"
                  value={editingTaskLabels}
                  onChange={(event) => setEditingTaskLabels(event.target.value)}
                  placeholder="frontend, urgent"
                />
              </label>
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
