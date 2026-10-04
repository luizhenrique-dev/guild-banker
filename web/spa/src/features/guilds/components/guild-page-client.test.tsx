import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { beforeEach, expect, it } from 'vitest'
import { http, HttpResponse } from 'msw'
import { server } from '@/src/mocks/server'
import { App } from '@/src/app/app'
import i18n from '@/src/shared/lib/i18n'

beforeEach(async () => { await i18n.changeLanguage('pt-BR') })
it('keeps the invitation draft after failure and clears it only on success', async () => {
  let calls = 0
  server.use(http.post('*/api/v1/guilds/1/invites', () => { calls++; return new HttpResponse(null, { status: 500 }) }))
  render(<MemoryRouter initialEntries={['/guild?guild=1']}><App initialUser={{ id: 1, name: 'Ana', email: 'ana@familia.com' }} demo /></MemoryRouter>)
  const user = userEvent.setup()
  const input = await screen.findByRole('textbox', { name: 'Convidar membro' })
  await user.type(input, 'test@example.com')
  await user.click(screen.getByRole('button', { name: 'Enviar convite' }))
  await waitFor(() => expect(calls).toBe(1))
  await waitFor(() => expect(screen.getByRole('button', { name: 'Enviar convite' })).toBeEnabled())
  expect(input).toHaveValue('test@example.com')
  server.use(http.post('*/api/v1/guilds/1/invites', () => new HttpResponse(null, { status: 204 })))
  await user.click(screen.getByRole('button', { name: 'Enviar convite' }))
  await waitFor(() => expect(input).toHaveValue(''))
})
it('offers retry when guild listing fails', async () => {
  server.use(http.get('*/api/v1/guilds', () => new HttpResponse(null, { status: 400 })))
  render(<MemoryRouter initialEntries={['/guild']}><App initialUser={{ id: 1, name: 'Ana', email: 'ana@familia.com' }} demo /></MemoryRouter>)
  const retry = await screen.findByRole('button', { name: 'Tentar novamente' })
  server.resetHandlers()
  await userEvent.click(retry)
  expect(await screen.findByRole('textbox', { name: 'Convidar membro' })).toBeVisible()
})
it('shows the empty guild state without requesting financial data', async () => {
  let financialRequests = 0
  server.use(
    http.get('*/api/v1/guilds', () => HttpResponse.json([])),
    http.get('*/api/v1/guilds/:id/transactions', () => { financialRequests++; return HttpResponse.json([]) }),
  )
  render(<MemoryRouter initialEntries={['/transactions']}><App initialUser={{ id: 1, name: 'Ana', email: 'ana@familia.com' }} demo /></MemoryRouter>)
  expect(await screen.findByText('Nenhuma guild disponível.')).toBeVisible()
  expect(financialRequests).toBe(0)
})
