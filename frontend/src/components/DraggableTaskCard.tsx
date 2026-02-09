import { useDraggable } from '@dnd-kit/core'
import { CSS } from '@dnd-kit/utilities'
import type { Task } from '../shared/api/types'

type DraggableTaskCardProps = {
  task: Task
  onEdit: (task: Task) => void
  onDelete: (task: Task) => void
  pending: boolean
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
      {...attributes}
      {...listeners}
    >
      <header>
        <h4>{task.title}</h4>
      </header>
      {task.description && <p>{task.description}</p>}
      <div className="task-card__actions">
        <button type="button" className="ghost-button" onClick={() => onEdit(task)}>
          Редактировать
        </button>
        <button
          type="button"
          className="danger-button"
          onClick={() => onDelete(task)}
          disabled={pending}
        >
          Удалить
        </button>
      </div>
    </article>
  )
}
