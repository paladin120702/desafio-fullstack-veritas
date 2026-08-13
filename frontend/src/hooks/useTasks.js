import { useCallback, useEffect, useRef, useState } from 'react'
import { api, ApiError } from '../api/client'
import { COLUMN_IDS } from '../constants'

/**
 * Reordena a lista movendo `id` para `status` na posição `position`.
 *
 * A ordem dentro de uma coluna é a ordem de aparição no array — a mesma regra
 * usada pelo backend —, então mover é remover e reinserir.
 */
function reorder(tasks, id, status, position) {
  const moving = tasks.find((task) => task.id === id)
  if (!moving) return tasks

  const columns = new Map(COLUMN_IDS.map((columnId) => [columnId, []]))
  for (const task of tasks) {
    if (task.id === id) continue
    columns.get(task.status)?.push(task)
  }

  const target = columns.get(status) ?? []
  const index = position == null || position < 0 ? target.length : Math.min(position, target.length)
  target.splice(index, 0, { ...moving, status })

  return COLUMN_IDS.flatMap((columnId) =>
    (columns.get(columnId) ?? []).map((task, i) => ({ ...task, position: i })),
  )
}

/**
 * Estado do quadro e ações de CRUD.
 *
 * Mover, editar e excluir são otimistas: a UI muda na hora e volta ao estado
 * anterior se a API recusar. Criar espera a resposta, porque o ID vem do servidor.
 */
export function useTasks() {
  const [tasks, setTasks] = useState([])
  const [loading, setLoading] = useState(true)
  const [creating, setCreating] = useState(false)
  const [error, setError] = useState(null)
  const [busyIds, setBusyIds] = useState(() => new Set())

  // Guarda o estado atual para o rollback sem recriar os callbacks a cada render.
  const tasksRef = useRef(tasks)
  tasksRef.current = tasks

  const markBusy = useCallback((id, busy) => {
    setBusyIds((current) => {
      const next = new Set(current)
      if (busy) next.add(id)
      else next.delete(id)
      return next
    })
  }, [])

  const load = useCallback(async (signal) => {
    setLoading(true)
    try {
      const data = await api.listTasks(signal)
      setTasks(data)
      setError(null)
    } catch (err) {
      if (err.name === 'AbortError') return
      setError(err instanceof ApiError ? err.message : 'Erro inesperado ao carregar as tarefas.')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    const controller = new AbortController()
    load(controller.signal)
    return () => controller.abort()
  }, [load])

  const createTask = useCallback(async ({ title, description, status }) => {
    setCreating(true)
    try {
      const created = await api.createTask({ title, description, status })
      setTasks((current) => [...current, created])
      setError(null)
      return { ok: true, task: created }
    } catch (err) {
      const message = err instanceof ApiError ? err.message : 'Erro inesperado ao criar a tarefa.'
      const fields = err instanceof ApiError ? err.fields : null
      if (!fields) setError(message)
      return { ok: false, message, fields }
    } finally {
      setCreating(false)
    }
  }, [])

  const updateTask = useCallback(
    async (id, changes) => {
      const previous = tasksRef.current
      setTasks((current) => current.map((task) => (task.id === id ? { ...task, ...changes } : task)))
      markBusy(id, true)

      try {
        const updated = await api.updateTask(id, changes)
        setTasks((current) => current.map((task) => (task.id === id ? updated : task)))
        setError(null)
        return { ok: true, task: updated }
      } catch (err) {
        setTasks(previous)
        const message = err instanceof ApiError ? err.message : 'Erro inesperado ao salvar a tarefa.'
        const fields = err instanceof ApiError ? err.fields : null
        if (!fields) setError(message)
        return { ok: false, message, fields }
      } finally {
        markBusy(id, false)
      }
    },
    [markBusy],
  )

  const moveTask = useCallback(
    async (id, status, position) => {
      const previous = tasksRef.current
      const task = previous.find((item) => item.id === id)
      if (!task) return { ok: false }

      const targetSize = previous.filter((item) => item.status === status && item.id !== id).length
      const nextPosition = position == null ? targetSize : Math.max(0, Math.min(position, targetSize))
      if (task.status === status && task.position === nextPosition) return { ok: true }

      setTasks(reorder(previous, id, status, nextPosition))
      markBusy(id, true)

      try {
        await api.updateTask(id, { status, position: nextPosition })
        setError(null)
        return { ok: true }
      } catch (err) {
        setTasks(previous)
        setError(err instanceof ApiError ? err.message : 'Erro inesperado ao mover a tarefa.')
        return { ok: false }
      } finally {
        markBusy(id, false)
      }
    },
    [markBusy],
  )

  const deleteTask = useCallback(
    async (id) => {
      const previous = tasksRef.current
      setTasks((current) => current.filter((task) => task.id !== id))
      markBusy(id, true)

      try {
        await api.deleteTask(id)
        setError(null)
        return { ok: true }
      } catch (err) {
        setTasks(previous)
        setError(err instanceof ApiError ? err.message : 'Erro inesperado ao excluir a tarefa.')
        return { ok: false }
      } finally {
        markBusy(id, false)
      }
    },
    [markBusy],
  )

  return {
    tasks,
    loading,
    creating,
    error,
    busyIds,
    reload: () => load(),
    dismissError: () => setError(null),
    createTask,
    updateTask,
    moveTask,
    deleteTask,
  }
}
