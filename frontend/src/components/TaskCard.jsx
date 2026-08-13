import { useState } from 'react'
import { COLUMNS, COLUMN_IDS } from '../constants'
import { TaskForm } from './TaskForm'

function formatDate(value) {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  return date.toLocaleString('pt-BR', { day: '2-digit', month: '2-digit', hour: '2-digit', minute: '2-digit' })
}

/**
 * Cartão de uma tarefa: leitura, edição inline, mover entre colunas e excluir.
 *
 * O arrasto é HTML5 nativo; os botões ‹ › fazem o mesmo caminho por teclado/clique,
 * para quem não usa (ou não consegue usar) drag and drop.
 */
export function TaskCard({ task, busy, isDragging, onEdit, onDelete, onMove, onDragStart, onDragEnd, onDragOver }) {
  const [editing, setEditing] = useState(false)
  const [confirmingDelete, setConfirmingDelete] = useState(false)

  const columnIndex = COLUMN_IDS.indexOf(task.status)
  const previousColumn = COLUMNS[columnIndex - 1]
  const nextColumn = COLUMNS[columnIndex + 1]

  async function handleEdit(values) {
    const result = await onEdit(task.id, values)
    if (result.ok) setEditing(false)
    return result
  }

  if (editing) {
    return (
      <li className="card card--editing">
        <TaskForm
          initialValues={{ title: task.title, description: task.description || '' }}
          submitLabel="Salvar"
          busy={busy}
          autoFocus
          compact
          onSubmit={handleEdit}
          onCancel={() => setEditing(false)}
        />
      </li>
    )
  }

  return (
    <li
      className={`card${isDragging ? ' card--dragging' : ''}${busy ? ' card--busy' : ''}`}
      draggable={!busy}
      onDragStart={(event) => onDragStart(event, task)}
      onDragEnd={onDragEnd}
      onDragOver={onDragOver}
      aria-busy={busy}
    >
      <div className="card__body">
        <h3 className="card__title">{task.title}</h3>
        {task.description && <p className="card__description">{task.description}</p>}
        <p className="card__meta">Atualizada em {formatDate(task.updatedAt)}</p>
      </div>

      {confirmingDelete ? (
        <div className="card__confirm" role="alertdialog" aria-label={`Excluir ${task.title}?`}>
          <span>Excluir esta tarefa?</span>
          <div className="card__confirm-actions">
            <button type="button" className="button button--ghost" onClick={() => setConfirmingDelete(false)}>
              Não
            </button>
            <button type="button" className="button button--danger" onClick={() => onDelete(task.id)} disabled={busy}>
              Sim, excluir
            </button>
          </div>
        </div>
      ) : (
        <div className="card__actions">
          <div className="card__move">
            <button
              type="button"
              className="button button--icon"
              onClick={() => onMove(task.id, previousColumn.id)}
              disabled={!previousColumn || busy}
              title={previousColumn ? `Mover para ${previousColumn.title}` : 'Já está na primeira coluna'}
              aria-label={previousColumn ? `Mover ${task.title} para ${previousColumn.title}` : 'Já está na primeira coluna'}
            >
              ‹
            </button>
            <button
              type="button"
              className="button button--icon"
              onClick={() => onMove(task.id, nextColumn.id)}
              disabled={!nextColumn || busy}
              title={nextColumn ? `Mover para ${nextColumn.title}` : 'Já está na última coluna'}
              aria-label={nextColumn ? `Mover ${task.title} para ${nextColumn.title}` : 'Já está na última coluna'}
            >
              ›
            </button>
          </div>

          <div className="card__crud">
            <button type="button" className="button button--ghost" onClick={() => setEditing(true)} disabled={busy}>
              Editar
            </button>
            <button
              type="button"
              className="button button--ghost button--danger-text"
              onClick={() => setConfirmingDelete(true)}
              disabled={busy}
            >
              Excluir
            </button>
          </div>
        </div>
      )}
    </li>
  )
}
