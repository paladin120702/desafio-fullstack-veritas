import { Board } from './components/Board'
import { ErrorBanner, LoadingState } from './components/Feedback'
import { TaskForm } from './components/TaskForm'
import { useTasks } from './hooks/useTasks'

export default function App() {
  const {
    tasks,
    loading,
    creating,
    error,
    busyIds,
    reload,
    dismissError,
    createTask,
    updateTask,
    moveTask,
    deleteTask,
  } = useTasks()

  const showBoard = !loading || tasks.length > 0

  return (
    <div className="app">
      <header className="app__header">
        <div>
          <h1 className="app__title">Mini Kanban</h1>
          <p className="app__subtitle">
            {tasks.length === 0
              ? 'Nenhuma tarefa cadastrada ainda.'
              : `${tasks.length} ${tasks.length === 1 ? 'tarefa' : 'tarefas'} no quadro.`}
          </p>
        </div>
        <button type="button" className="button button--ghost" onClick={reload} disabled={loading}>
          {loading ? 'Atualizando…' : 'Atualizar'}
        </button>
      </header>

      <section className="app__new-task" aria-label="Adicionar tarefa">
        <h2 className="section-title">Nova tarefa</h2>
        <TaskForm submitLabel="Adicionar" busy={creating} onSubmit={createTask} />
      </section>

      {error && <ErrorBanner message={error} onRetry={reload} onDismiss={dismissError} />}

      {loading && tasks.length === 0 && <LoadingState />}

      {showBoard && (
        <Board
          tasks={tasks}
          busyIds={busyIds}
          onEdit={updateTask}
          onDelete={deleteTask}
          onMove={moveTask}
        />
      )}

      <footer className="app__footer">
        Arraste os cartões entre as colunas ou use os botões ‹ › para mover.
      </footer>
    </div>
  )
}
