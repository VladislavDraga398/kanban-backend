import { useDraggable } from '@dnd-kit/core'
import { CSS } from '@dnd-kit/utilities'
import type { Task, TaskPriority } from '../shared/api/types'

type DraggableTaskCardProps = {
  task: Task
  onEdit: (task: Task) => void
  onDelete: (task: Task) => void
  pending: boolean
}

function formatDate(value: string) {
  return new Date(value).toLocaleDateString('ru-RU', {
    day: '2-digit',
    month: 'short',
  })
}

const priorityLabel: Record<TaskPriority, string> = {
  low: 'Низкий',
  normal: 'Обычный',
  high: 'Высокий',
}

export function DraggableTaskCard({ task, onEdit, onDelete, pending }: DraggableTaskCardProps) {
  const { attributes, listeners, setNodeRef, transform, isDragging } = useDraggable({
    id: task.id,
    data: {
      taskId: task.id,
      columnId: task.column_id,
    },
  })

  const style = {
    transform: CSS.Translate.toString(transform),
  }

  return (
    <article
      ref={setNodeRef}
      style={style}
      className={`task-card ${isDragging ? 'is-dragging' : ''}`}
    >
      <header className="task-card__header" {...attributes} {...listeners}>
        <span className="task-card__grip" aria-hidden="true" />
        <div>
          <h4>{task.title}</h4>
          <div className="task-card__meta">
            <time dateTime={task.updated_at}>{formatDate(task.updated_at)}</time>
            <span>#{task.id.slice(0, 5)}</span>
          </div>
        </div>
      </header>
      {task.description && <p>{task.description}</p>}
      <div className="task-card__badges">
        <span className={`priority-badge priority-badge--${task.priority}`}>
          {priorityLabel[task.priority]}
        </span>
        {task.due_date && (
          <time className="due-badge" dateTime={task.due_date}>
            до {formatDate(task.due_date)}
          </time>
        )}
        {task.labels.map((label) => (
          <span className="label-badge" key={label}>
            {label}
          </span>
        ))}
      </div>
      <div className="task-card__actions">
        <button
          type="button"
          className="ghost-button"
          onClick={() => onEdit(task)}
          title="Редактировать задачу"
        >
          Править
        </button>
        <button
          type="button"
          className="danger-button"
          onClick={() => onDelete(task)}
          disabled={pending}
          title="Удалить задачу"
        >
          Удалить
        </button>
      </div>
    </article>
  )
}
