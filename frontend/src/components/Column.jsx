import { Fragment, useState } from 'react'
import { TaskCard } from './TaskCard'

/**
 * Uma das três colunas fixas: cabeçalho com contador, lista de cartões e a
 * área de soltar do drag and drop.
 */
export function Column({
  column,
  tasks,
  busyIds,
  draggedTask,
  onDragStart,
  onDragEnd,
  onDrop,
  onEdit,
  onDelete,
  onMove,
}) {
  const [dropIndex, setDropIndex] = useState(null)
  const isDragActive = Boolean(draggedTask)

  function handleDragOver(event) {
    if (!isDragActive) return
    event.preventDefault()
    event.dataTransfer.dropEffect = 'move'
    // Fora de qualquer cartão: solta no fim da coluna.
    setDropIndex((current) => (current === null ? tasks.length : current))
  }

  function handleCardDragOver(event, index) {
    if (!isDragActive) return
    event.preventDefault()
    event.stopPropagation()
    const rect = event.currentTarget.getBoundingClientRect()
    const isBelowMiddle = event.clientY > rect.top + rect.height / 2
    setDropIndex(isBelowMiddle ? index + 1 : index)
  }

  function handleDragLeave(event) {
    // dragleave dispara ao passar entre filhos; só limpa ao sair da coluna.
    if (!event.currentTarget.contains(event.relatedTarget)) setDropIndex(null)
  }

  function handleDrop(event) {
    if (!isDragActive) return
    event.preventDefault()
    const index = dropIndex ?? tasks.length
    setDropIndex(null)
    onDrop(column.id, index)
  }

  return (
    <section
      className={`column${isDragActive ? ' column--drop-target' : ''}`}
      style={{ '--column-accent': column.accent }}
      aria-label={column.title}
      onDragOver={handleDragOver}
      onDragLeave={handleDragLeave}
      onDrop={handleDrop}
    >
      <header className="column__header">
        <h2 className="column__title">{column.title}</h2>
        <span className="column__count" aria-label={`${tasks.length} tarefas`}>
          {tasks.length}
        </span>
      </header>

      <ul className="column__list">
        {tasks.map((task, index) => (
          <Fragment key={task.id}>
            {dropIndex === index && <li className="drop-marker" aria-hidden="true" />}
            <TaskCard
              task={task}
              busy={busyIds.has(task.id)}
              isDragging={draggedTask?.id === task.id}
              onEdit={onEdit}
              onDelete={onDelete}
              onMove={onMove}
              onDragStart={onDragStart}
              onDragEnd={onDragEnd}
              onDragOver={(event) => handleCardDragOver(event, index)}
            />
          </Fragment>
        ))}
        {dropIndex === tasks.length && <li className="drop-marker" aria-hidden="true" />}
      </ul>

      {tasks.length === 0 && (
        <p className="column__empty">
          {isDragActive ? 'Solte aqui' : 'Nenhuma tarefa nesta coluna.'}
        </p>
      )}
    </section>
  )
}
