import { useMemo, useState } from 'react'
import { COLUMNS } from '../constants'
import { Column } from './Column'

/** Agrupa as tarefas por coluna preservando a ordem devolvida pela API. */
function groupByColumn(tasks) {
  const groups = Object.fromEntries(COLUMNS.map((column) => [column.id, []]))
  for (const task of tasks) {
    if (groups[task.status]) groups[task.status].push(task)
  }
  return groups
}

/** O quadro: três colunas fixas e a coordenação do drag and drop entre elas. */
export function Board({ tasks, busyIds, onEdit, onDelete, onMove }) {
  const [draggedTask, setDraggedTask] = useState(null)
  const grouped = useMemo(() => groupByColumn(tasks), [tasks])

  function handleDragStart(event, task) {
    setDraggedTask(task)
    event.dataTransfer.effectAllowed = 'move'
    // Alguns navegadores só iniciam o arrasto se houver payload.
    event.dataTransfer.setData('text/plain', task.id)
  }

  function handleDragEnd() {
    setDraggedTask(null)
  }

  function handleDrop(columnId, index) {
    const task = draggedTask
    setDraggedTask(null)
    if (!task) return

    // O índice foi calculado com o cartão arrastado ainda na lista; ao mover
    // dentro da mesma coluna, tudo depois dele desloca uma posição para cima.
    const position = task.status === columnId && index > task.position ? index - 1 : index
    onMove(task.id, columnId, position)
  }

  return (
    <div className="board">
      {COLUMNS.map((column) => (
        <Column
          key={column.id}
          column={column}
          tasks={grouped[column.id]}
          busyIds={busyIds}
          draggedTask={draggedTask}
          onDragStart={handleDragStart}
          onDragEnd={handleDragEnd}
          onDrop={handleDrop}
          onEdit={onEdit}
          onDelete={onDelete}
          onMove={onMove}
        />
      ))}
    </div>
  )
}
