import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { it, expect } from 'vitest'
import i18n from '../lib/i18n'
import { AccessDeniedPage } from './access-denied-page'

it.each([
  ['pt-BR', 'Acesso negado', 'Voltar ao início'],
  ['en', 'Access denied', 'Back to Home'],
])('renders generic accessible content in %s', async (language, title, link) => {
  await i18n.changeLanguage(language)
  render(<MemoryRouter><AccessDeniedPage /></MemoryRouter>)
  expect(screen.getByRole('main', { name: title })).toBeVisible()
  expect(screen.getByRole('heading', { name: title, level: 1 })).toBeVisible()
  expect(screen.getByRole('link', { name: link })).toHaveAttribute('href', '/home')
  expect(screen.queryByRole('button')).toBeNull()
})
