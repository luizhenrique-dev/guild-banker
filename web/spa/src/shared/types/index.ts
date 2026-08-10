// ============================================================
// 📌 PADRÃO: Tipos compartilhados derivados da OpenAPI spec
// Centralizamos aqui os tipos que são usados por múltiplas features.
// ============================================================

/** Envelope padrão de erro da API (gin.H{"error": ...}) */
export interface ApiError {
  error: string
}

/** Resposta paginada por cursor (usada em listTransactions) */
export interface CursorPaginatedResponse<T> {
  items: T[]
  nextCursor?: string
}

/** Parâmetros de paginação por cursor */
export interface CursorPaginationParams {
  cursor?: string
  limit?: number
}
