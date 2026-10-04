import { expect, it, vi } from 'vitest'
import { screen } from '@testing-library/react'

vi.mock('./start-mocks', () => ({ startMocks: vi.fn().mockRejectedValue(new Error('worker failed')) }))

it('renders a visible startup failure instead of financial pages', async () => {
  const element = document.createElement('div')
  element.id = 'root'
  document.body.append(element)
  await import('../main')
  expect(await screen.findByRole('alert')).toHaveTextContent('Falha ao iniciar a demonstração')
  expect(screen.queryByRole('navigation')).not.toBeInTheDocument()
  element.remove()
})
