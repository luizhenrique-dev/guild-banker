import type { components } from '@/src/shared/api/schema'
export type TransactionType = components['schemas']['TransactionType']
export type TransactionCategory = components['schemas']['TransactionCategory']
export type TransactionStatus = components['schemas']['TransactionStatus']
export type TransactionSource = components['schemas']['TransactionSource']
export type TransactionVisibility = components['schemas']['TransactionVisibility']
export type TransactionResponse = components['schemas']['TransactionResponse']

export const TRANSACTION_CATEGORIES: TransactionCategory[] = [
  'GROCERY', 'HOUSING', 'UTILITIES', 'SUBSCRIPTIONS', 'INSURANCE',
  'EDUCATION', 'TRANSPORTATION', 'HEALTH', 'PERSONAL_CARE', 'TAXES',
  'OTHER', 'FOOD_AND_DINING', 'ENTERTAINMENT', 'SHOPPING', 'PETS',
  'TRAVEL', 'INVESTMENTS',
]
