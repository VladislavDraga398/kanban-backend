import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useMemo, useState, type FormEvent } from 'react'
import { Link } from 'react-router-dom'
import { useAuth } from '../auth/use-auth'
import {
  createBoard,
  deleteBoard,
  listBoards,
  updateBoard,
} from '../features/boards/api'
import { getErrorMessage } from '../shared/api/errors'
import type { Board } from '../shared/api/types'

function BoardCard({
  board,
  onRename,
  onDelete,
  pending,
}: {
  board: Board
  onRename: (board: Board) => void
  onDelete: (board: Board) => void
  pending: boolean
}) {
  return (
    <article className="board-card">
      <div className="board-card__content">
        <h3>{board.name}</h3>
        <p>
          Обновлена: {new Date(board.updated_at).toLocaleString()}
        </p>
      </div>
      <div className="board-card__actions">
        <Link to={`/boards/${board.id}`} className="ghost-button">
          Открыть
        </Link>
        <button type="button" className="ghost-button" onClick={() => onRename(board)}>
          Переименовать
        </button>
        <button
          type="button"
          className="danger-button"
          onClick={() => onDelete(board)}
          disabled={pending}
        >
          Удалить
        </button>
      </div>
    </article>
  )
}

export function BoardsPage() {
  const queryClient = useQueryClient()
  const { logout } = useAuth()

  const [newBoardName, setNewBoardName] = useState('')
  const [editingBoard, setEditingBoard] = useState<Board | null>(null)
  const [editName, setEditName] = useState('')
  const [actionError, setActionError] = useState<string | null>(null)

  const boardsQuery = useQuery({
    queryKey: ['boards'],
    queryFn: listBoards,
  })

  const createMutation = useMutation({
    mutationFn: createBoard,
    onSuccess: () => {
      setNewBoardName('')
      setActionError(null)
      queryClient.invalidateQueries({ queryKey: ['boards'] })
    },
    onError: (error) => {
      setActionError(getErrorMessage(error))
    },
  })

  const renameMutation = useMutation({
    mutationFn: updateBoard,
    onSuccess: () => {
      setEditingBoard(null)
      setEditName('')
      setActionError(null)
      queryClient.invalidateQueries({ queryKey: ['boards'] })
    },
    onError: (error) => {
      setActionError(getErrorMessage(error))
    },
  })

  const deleteMutation = useMutation({
    mutationFn: deleteBoard,
    onSuccess: () => {
      setActionError(null)
      queryClient.invalidateQueries({ queryKey: ['boards'] })
    },
    onError: (error) => {
      setActionError(getErrorMessage(error))
    },
  })

  function onCreateBoard(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const name = newBoardName.trim()
    if (!name) {
      return
    }
    createMutation.mutate({ name })
  }

  function openRenameModal(board: Board) {
    setEditingBoard(board)
    setEditName(board.name)
  }

  function submitRename(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!editingBoard) {
      return
    }
    const name = editName.trim()
    if (!name) {
      return
    }
    renameMutation.mutate({ id: editingBoard.id, name })
  }

  const isBusy = useMemo(
    () => createMutation.isPending || renameMutation.isPending || deleteMutation.isPending,
    [createMutation.isPending, renameMutation.isPending, deleteMutation.isPending],
  )

  return (
    <main className="page-shell">
      <header className="topbar">
        <div>
          <p className="badge">KANBAN CONTROL</p>
          <h1>Доски</h1>
        </div>
        <button type="button" className="ghost-button" onClick={logout}>
          Выйти
        </button>
      </header>

      <section className="panel">
        <h2>Создать доску</h2>
        <form className="inline-form" onSubmit={onCreateBoard}>
          <input
            type="text"
            value={newBoardName}
            onChange={(event) => setNewBoardName(event.target.value)}
            placeholder="Например: Product Roadmap"
            required
          />
          <button type="submit" className="primary-button" disabled={createMutation.isPending}>
            Добавить
          </button>
        </form>
      </section>

      {actionError && <p className="error-text panel">{actionError}</p>}
      {boardsQuery.isError && (
        <p className="error-text panel">{getErrorMessage(boardsQuery.error)}</p>
      )}

      <section className="board-grid">
        {boardsQuery.isLoading && <p className="panel">Загружаю доски...</p>}
        {boardsQuery.data?.length === 0 && (
          <p className="panel">Пока пусто. Создай первую доску.</p>
        )}
        {boardsQuery.data?.map((board) => (
          <BoardCard
            key={board.id}
            board={board}
            onRename={openRenameModal}
            onDelete={(item) => deleteMutation.mutate(item.id)}
            pending={deleteMutation.isPending}
          />
        ))}
      </section>

      {editingBoard && (
        <div className="modal-backdrop" onClick={() => setEditingBoard(null)}>
          <section
            className="modal"
            onClick={(event) => event.stopPropagation()}
            role="dialog"
            aria-modal="true"
          >
            <h3>Переименовать доску</h3>
            <form className="modal-form" onSubmit={submitRename}>
              <input
                type="text"
                value={editName}
                onChange={(event) => setEditName(event.target.value)}
                required
              />
              <div className="modal-actions">
                <button
                  type="button"
                  className="ghost-button"
                  onClick={() => setEditingBoard(null)}
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
