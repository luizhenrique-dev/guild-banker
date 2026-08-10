export type ImportBatchStatus = 'PENDING_REVIEW' | 'COMPLETED' | 'CANCELLED'
export type ImportItemStatus = 'READY' | 'DUPLICATE' | 'DISCARDED'

export interface ImportItemResponse {
  itemId: number
  occurredAt: string
  description: string
  amount: string
  type: 'EXPENSE' | 'INCOME'
  category: string
  bankCategory: string
  cardLast4: string
  installment: string
  status: ImportItemStatus
}

export interface UploadImportResponse {
  importId: number
  status: ImportBatchStatus
  fileName: string
  summary: {
    parsed: number
    candidates: number
    duplicates: number
    skippedZero: number
  }
  items: ImportItemResponse[]
}
