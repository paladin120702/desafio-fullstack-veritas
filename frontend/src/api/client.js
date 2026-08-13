const BASE_URL = (
  import.meta.env.VITE_API_URL || 'http://localhost:8080/api'
).replace(/\/$/, '')

/** Erro de API com a mensagem já pronta para exibição e os erros por campo. */
export class ApiError extends Error {
  constructor(message, { status = 0, code = 'unknown', fields = null } = {}) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
    this.fields = fields
  }
}

async function request(path, { method = 'GET', body, signal } = {}) {
  let response
  try {
    response = await fetch(`${BASE_URL}${path}`, {
      method,
      signal,
      headers: body ? { 'Content-Type': 'application/json' } : undefined,
      body: body ? JSON.stringify(body) : undefined,
    })
  } catch (err) {
    if (err.name === 'AbortError') throw err
    // fetch só rejeita por falha de rede/CORS — o servidor não respondeu.
    throw new ApiError(
      'Não foi possível falar com o servidor. Verifique se o backend está rodando.',
      { code: 'network_error' },
    )
  }

  if (response.status === 204) return null

  const payload = await response.json().catch(() => null)

  if (!response.ok) {
    const apiError = payload?.error
    throw new ApiError(apiError?.message || `Erro ${response.status} ao chamar a API.`, {
      status: response.status,
      code: apiError?.code || 'http_error',
      fields: apiError?.fields || null,
    })
  }

  return payload
}

export const api = {
  listTasks: (signal) => request('/tasks', { signal }).then((data) => data?.tasks ?? []),

  createTask: ({ title, description, status }) =>
    request('/tasks', { method: 'POST', body: { title, description, status } }),

  updateTask: (id, changes) =>
    request(`/tasks/${encodeURIComponent(id)}`, { method: 'PUT', body: changes }),

  deleteTask: (id) => request(`/tasks/${encodeURIComponent(id)}`, { method: 'DELETE' }),
}

export { BASE_URL }
