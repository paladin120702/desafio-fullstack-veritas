/** As três colunas fixas do quadro. O `id` é o mesmo `status` da API. */
export const COLUMNS = [
  { id: 'todo', title: 'A Fazer', accent: '#c99a3f' },
  { id: 'doing', title: 'Em Progresso', accent: '#3f7bc9' },
  { id: 'done', title: 'Concluídas', accent: '#3fa06a' },
]

export const COLUMN_IDS = COLUMNS.map((column) => column.id)

export const MAX_TITLE_LENGTH = 120
export const MAX_DESCRIPTION_LENGTH = 1000
