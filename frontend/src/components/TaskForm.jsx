import { useEffect, useRef, useState } from 'react'
import { MAX_DESCRIPTION_LENGTH, MAX_TITLE_LENGTH } from '../constants'

/**
 * Formulário usado tanto para criar quanto para editar uma tarefa.
 *
 * `onSubmit` devolve `{ ok, fields }`: quando a API recusa por validação,
 * os erros por campo voltam para cá e são exibidos abaixo do input.
 */
export function TaskForm({
  initialValues = { title: '', description: '' },
  submitLabel = 'Adicionar',
  busy = false,
  autoFocus = false,
  compact = false,
  onSubmit,
  onCancel,
}) {
  const [title, setTitle] = useState(initialValues.title)
  const [description, setDescription] = useState(initialValues.description)
  const [fieldErrors, setFieldErrors] = useState({})
  const titleRef = useRef(null)

  useEffect(() => {
    if (autoFocus) titleRef.current?.focus()
  }, [autoFocus])

  async function handleSubmit(event) {
    event.preventDefault()

    // Validação local espelha a do backend para dar resposta imediata.
    if (!title.trim()) {
      setFieldErrors({ title: 'Informe um título para a tarefa.' })
      titleRef.current?.focus()
      return
    }

    setFieldErrors({})
    const result = await onSubmit({ title: title.trim(), description: description.trim() })

    if (result?.ok) {
      if (!initialValues.title) {
        setTitle('')
        setDescription('')
      }
      return
    }
    if (result?.fields) setFieldErrors(result.fields)
  }

  function handleKeyDown(event) {
    if (event.key === 'Escape' && onCancel) onCancel()
    // Ctrl/Cmd + Enter envia a partir da descrição.
    if (event.key === 'Enter' && (event.metaKey || event.ctrlKey)) handleSubmit(event)
  }

  return (
    <form className={`task-form${compact ? ' task-form--compact' : ''}`} onSubmit={handleSubmit} onKeyDown={handleKeyDown}>
      <div className="field">
        <label className="field__label" htmlFor={`title-${submitLabel}`}>
          Título <span aria-hidden="true">*</span>
        </label>
        <input
          id={`title-${submitLabel}`}
          ref={titleRef}
          className={`field__input${fieldErrors.title ? ' field__input--invalid' : ''}`}
          value={title}
          maxLength={MAX_TITLE_LENGTH}
          placeholder="Ex.: Revisar contrato do cliente"
          onChange={(event) => setTitle(event.target.value)}
          aria-invalid={Boolean(fieldErrors.title)}
          aria-describedby={fieldErrors.title ? `title-error-${submitLabel}` : undefined}
          disabled={busy}
        />
        {fieldErrors.title && (
          <p className="field__error" id={`title-error-${submitLabel}`} role="alert">
            {fieldErrors.title}
          </p>
        )}
      </div>

      <div className="field">
        <label className="field__label" htmlFor={`description-${submitLabel}`}>
          Descrição <span className="field__hint">(opcional)</span>
        </label>
        <textarea
          id={`description-${submitLabel}`}
          className={`field__input field__input--textarea${fieldErrors.description ? ' field__input--invalid' : ''}`}
          value={description}
          maxLength={MAX_DESCRIPTION_LENGTH}
          rows={compact ? 2 : 3}
          placeholder="Detalhes, links ou critérios de conclusão"
          onChange={(event) => setDescription(event.target.value)}
          disabled={busy}
        />
        {fieldErrors.description && (
          <p className="field__error" role="alert">
            {fieldErrors.description}
          </p>
        )}
      </div>

      <div className="task-form__actions">
        {onCancel && (
          <button type="button" className="button button--ghost" onClick={onCancel} disabled={busy}>
            Cancelar
          </button>
        )}
        <button type="submit" className="button button--primary" disabled={busy}>
          {busy ? 'Salvando…' : submitLabel}
        </button>
      </div>
    </form>
  )
}
