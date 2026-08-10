import { fixedExpensesHandlers } from './fixed-expenses'
import { guildsHandlers } from './guilds'
import { transactionsHandlers } from './transactions'
import { importsHandlers } from './imports'

export const handlers = [
  ...fixedExpensesHandlers,
  ...guildsHandlers,
  ...transactionsHandlers,
  ...importsHandlers,
]
