import { beforeEach, expect, it, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { App } from './app'
import i18n from '../shared/lib/i18n'

beforeEach(async () => {
  await i18n.changeLanguage('pt-BR')
})
it.each(['/', '/fixed-expenses', '/transactions', '/import', '/guild'])('gates %s without requests when session is absent', async (path) => {
  const fetchSpy = vi.spyOn(globalThis, 'fetch')
  render(<MemoryRouter initialEntries={[path]}><App /></MemoryRouter>)
  expect(await screen.findByRole('heading', { name: 'Sessão não integrada' })).toBeVisible()
  expect(screen.queryByText('Aluguel')).not.toBeInTheDocument()
  expect(fetchSpy).not.toHaveBeenCalled()
  fetchSpy.mockRestore()
})
it('handles unknown routes', async () => {
  render(<MemoryRouter initialEntries={['/missing']}><App /></MemoryRouter>)
  expect(await screen.findByRole('heading', { name: 'Página não encontrada' })).toBeVisible()
})

it('opens Home from root and menu, preserving query without requiring a guild', async () => {
  render(<MemoryRouter initialEntries={['/?guild=999']}><App initialUser={{ id: 99, name: '<b>Ana</b>', email: 'ana@example.com' }} demo /></MemoryRouter>)
  expect(await screen.findByRole('heading', { name: 'Bem-vindo(a) <b>Ana</b>' })).toBeVisible()
  const home = screen.getByRole('link', { name: 'Início' })
  expect(home).toHaveAttribute('href', '/home?guild=999')
  fireEvent.click(home)
  expect(screen.getByRole('heading', { name: 'Bem-vindo(a) <b>Ana</b>' }).querySelector('b')).toBeNull()
})
