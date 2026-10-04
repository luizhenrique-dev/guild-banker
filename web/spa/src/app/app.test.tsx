import { beforeEach, expect, it, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { App } from './app'
import { useAuthStore } from '../shared/store/auth.store'
import i18n from '../shared/lib/i18n'

beforeEach(async () => {
  await i18n.changeLanguage('pt-BR')
  useAuthStore.getState().clearUser()
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
