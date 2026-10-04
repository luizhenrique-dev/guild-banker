import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { beforeEach, expect, it } from 'vitest'
import { http, HttpResponse } from 'msw'
import { server } from '@/src/mocks/server'
import { App } from '@/src/app/app'
import i18n from '@/src/shared/lib/i18n'

beforeEach(async () => { await i18n.changeLanguage('pt-BR') })
it('preserves file and review after upload/confirmation failures without retrying mutations', async () => {
  let uploads = 0
  server.use(http.post('*/api/v1/guilds/1/imports', () => { uploads++; return new HttpResponse(null, { status: 500 }) }))
  render(<MemoryRouter initialEntries={['/import?guild=1']}><App initialUser={{ id: 1, name: 'Ana', email: 'ana@familia.com' }} demo /></MemoryRouter>)
  const user = userEvent.setup()
  const input = await screen.findByLabelText('Selecionar arquivo')
  await user.upload(input, new File(['csv'], 'statement.csv', { type: 'text/csv' }))
  await user.click(screen.getByRole('button', { name: 'Enviar arquivo' }))
  await waitFor(() => expect(uploads).toBe(1))
  await waitFor(() => expect(screen.getByRole('button', { name: 'Enviar arquivo' })).toBeEnabled())
  expect(screen.getByText('statement.csv')).toBeVisible()
  server.resetHandlers()
  await user.click(screen.getByRole('button', { name: 'Enviar arquivo' }))
  await screen.findByRole('heading', { name: 'Revisar itens importados' })
  let confirms = 0
  server.use(http.post('*/api/v1/guilds/1/imports/1/confirm', () => { confirms++; return new HttpResponse(null, { status: 400 }) }))
  await user.click(screen.getByRole('button', { name: 'Confirmar importação' }))
  await waitFor(() => expect(confirms).toBe(1))
  expect(screen.getByRole('heading', { name: 'Revisar itens importados' })).toBeVisible()
})
