import '@testing-library/jest-dom/vitest'
import { afterEach, beforeAll, afterAll } from 'vitest'
import { cleanup } from '@testing-library/react'
import { server } from '../mocks/server'
import { resetFixedExpensesMocks } from '../mocks/handlers/fixed-expenses'
import { resetGuildsMocks } from '../mocks/handlers/guilds'
import { resetImportsMocks } from '../mocks/handlers/imports'
import { resetTransactionsMocks } from '../mocks/handlers/transactions'

beforeAll(() => server.listen({ onUnhandledRequest: 'error' }))
afterEach(() => { cleanup(); server.resetHandlers(); resetFixedExpensesMocks(); resetGuildsMocks(); resetTransactionsMocks(); resetImportsMocks() })
afterAll(() => server.close())
