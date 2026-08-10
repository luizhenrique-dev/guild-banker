// Types: Transactions — derivados da OpenAPI spec

export type TransactionType = 'EXPENSE' | 'INCOME'

export type TransactionCategory =
  | 'GROCERY'
  | 'HOUSING'
  | 'UTILITIES'
  | 'SUBSCRIPTIONS'
  | 'INSURANCE'
  | 'EDUCATION'
  | 'TRANSPORTATION'
  | 'HEALTH'
  | 'PERSONAL_CARE'
  | 'TAXES'
  | 'OTHER'
  | 'FOOD_AND_DINING'
  | 'ENTERTAINMENT'
  | 'SHOPPING'
  | 'PETS'
  | 'TRAVEL'
  | 'INVESTMENTS'

export const TRANSACTION_CATEGORIES: TransactionCategory[] = [
  'GROCERY', 'HOUSING', 'UTILITIES', 'SUBSCRIPTIONS', 'INSURANCE',
  'EDUCATION', 'TRANSPORTATION', 'HEALTH', 'PERSONAL_CARE', 'TAXES',
  'OTHER', 'FOOD_AND_DINING', 'ENTERTAINMENT', 'SHOPPING', 'PETS',
  'TRAVEL', 'INVESTMENTS',
]

export type TransactionStatus = 'ACTIVE' | 'CANCELLED'
export type TransactionSource = 'MANUAL' | 'IMPORT'
export type TransactionVisibility = 'PRIVATE' | 'PUBLIC'

export interface TransactionResponse {
  id: number
  type: TransactionType
  description: string
  amount: string
  category: TransactionCategory
  status: TransactionStatus
  source: TransactionSource
  visibility: TransactionVisibility
  occurredAt: string
  guildId: number
  userAccountId: number
  createdAt: string
  updatedAt?: string
}
