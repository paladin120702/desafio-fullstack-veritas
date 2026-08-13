/** Faixa de erro global, com ação de tentar novamente. */
export function ErrorBanner({ message, onRetry, onDismiss }) {
  return (
    <div className="banner banner--error" role="alert">
      <span className="banner__message">{message}</span>
      <div className="banner__actions">
        {onRetry && (
          <button type="button" className="button button--ghost" onClick={onRetry}>
            Tentar novamente
          </button>
        )}
        {onDismiss && (
          <button type="button" className="button button--icon" onClick={onDismiss} aria-label="Fechar aviso">
            ×
          </button>
        )}
      </div>
    </div>
  )
}

/** Estado de carregamento inicial do quadro. */
export function LoadingState({ label = 'Carregando tarefas…' }) {
  return (
    <div className="loading" role="status" aria-live="polite">
      <span className="loading__spinner" aria-hidden="true" />
      <span>{label}</span>
    </div>
  )
}
