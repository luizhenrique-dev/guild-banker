import { expect, test } from '@playwright/test'

test('new fixed expense fades in continuously from its first rendered frame', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' })
  await page.goto('/fixed-expenses')
  await expect(page.getByRole('button', { name: 'Nova despesa fixa', exact: true })).toBeVisible()
  const frames = await page.evaluate(async () => {
    const samples: { elapsed: number; opacity: number; overlayOpacity: number; x: number; y: number }[] = []
    const start = performance.now()
    const button = [...document.querySelectorAll('button')].find((element) => element.textContent?.includes('Nova despesa fixa'))!
    button.click()
    await new Promise<void>((resolve) => {
      const sample = () => {
        const dialog = document.querySelector('[role="dialog"]')
        if (dialog) {
          const box = dialog.getBoundingClientRect()
          const overlay = dialog.previousElementSibling!
          samples.push({ elapsed: performance.now() - start, opacity: Number(getComputedStyle(dialog).opacity), overlayOpacity: Number(getComputedStyle(overlay).opacity), x: box.x + box.width / 2, y: box.y + box.height / 2 })
        }
        if (performance.now() - start < 500) requestAnimationFrame(sample)
        else resolve()
      }
      requestAnimationFrame(sample)
    })
    return samples
  })
  expect(frames[0]!.opacity, JSON.stringify(frames)).toBeLessThan(0.1)
  expect(frames.some((frame) => frame.elapsed >= 100 && frame.opacity > 0 && frame.opacity < 0.9), JSON.stringify(frames)).toBe(true)
  for (let i = 1; i < frames.length; i++) {
    expect(frames[i]!.opacity).toBeGreaterThanOrEqual(frames[i - 1]!.opacity)
    expect(Math.abs(frames[i]!.opacity - frames[i]!.overlayOpacity)).toBeLessThan(0.1)
  }
  expect(frames.at(-1)!.opacity).toBe(1)
})

for (const width of [375, 1280]) {
  test(`new fixed expense stays centered throughout its fade-in at ${width}px`, async ({ page }) => {
    await page.setViewportSize({ width, height: 900 })
    await page.emulateMedia({ reducedMotion: 'no-preference' })
    await page.goto('/fixed-expenses')
    await page.getByRole('button', { name: 'Nova despesa fixa', exact: true }).click()
    const frames = await page.getByRole('dialog').evaluate((element) => {
      const animation = element.getAnimations()[0]
      if (!animation) throw new Error('Expected an opening animation')
      animation.pause()
      const duration = Number(animation.effect!.getTiming().duration)
      const samples = [0, duration / 2, duration].map((time) => {
        animation.currentTime = time
        const box = element.getBoundingClientRect()
        return { x: box.x + box.width / 2, y: box.y + box.height / 2, opacity: Number(getComputedStyle(element).opacity) }
      })
      animation.finish()
      return samples
    })
    for (const frame of frames) {
      expect(Math.abs(frame.x - width / 2)).toBeLessThan(1)
      expect(Math.abs(frame.y - 450)).toBeLessThan(1)
    }
    expect(frames[0]!.opacity).toBe(0)
    expect(frames[1]!.opacity).toBeGreaterThan(0)
    expect(frames[1]!.opacity).toBeLessThan(1)
    expect(frames[2]!.opacity).toBe(1)
  })
}

test('new fixed expense opens centered without animation when reduced motion is requested', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'reduce' })
  await page.goto('/fixed-expenses')
  await page.getByRole('button', { name: 'Nova despesa fixa', exact: true }).click()
  const state = await page.getByRole('dialog').evaluate((element) => {
    const box = element.getBoundingClientRect()
    return { x: box.x + box.width / 2, y: box.y + box.height / 2, width: innerWidth, height: innerHeight, animations: element.getAnimations().length }
  })
  expect(state.animations).toBe(0)
  expect(Math.abs(state.x - state.width / 2)).toBeLessThan(1)
  expect(Math.abs(state.y - state.height / 2)).toBeLessThan(1)
})

for (const width of [375, 1280]) {
  test(`fixed expense CRUD, keyboard focus and local table scroll at ${width}px`, async ({ page }) => {
    await page.setViewportSize({ width, height: 900 })
    await page.goto('/fixed-expenses')
    await expect(page.getByRole('row').filter({ hasText: 'Aluguel' })).toBeVisible()
    const newButton = page.getByRole('button', { name: 'Nova despesa fixa', exact: true })
    await newButton.focus()
    await page.keyboard.press('Enter')
    const dialog = page.getByRole('dialog')
    await expect(dialog.getByRole('textbox', { name: 'Nome', exact: true })).toBeFocused()
    await page.keyboard.press('Escape')
    await expect(dialog).toHaveCount(0)
    await expect(newButton).toBeFocused()
    await page.keyboard.press('Enter')
    await dialog.getByRole('textbox', { name: 'Nome', exact: true }).fill('Internet teste')
    await dialog.getByRole('textbox', { name: 'Valor (R$)' }).fill('123,45')
    await dialog.getByRole('spinbutton', { name: 'Dia de vencimento' }).fill('31')
    await dialog.getByRole('combobox', { name: 'Categoria' }).selectOption('SUBSCRIPTIONS')
    const create = page.waitForRequest((request) => request.url().includes('/api/v1/guilds/1/fixed-expenses') && request.method() === 'POST')
    await dialog.getByRole('button', { name: 'Salvar', exact: true }).click()
    expect((await create).postDataJSON()).toEqual({ name: 'Internet teste', amount: '123.45', due_day: 31, category: 'SUBSCRIPTIONS' })
    const row = page.getByRole('row').filter({ hasText: 'Internet teste' })
    await expect(row).toContainText('123,45')
    const edit = row.getByRole('button', { name: 'Editar Internet teste' })
    await edit.focus()
    await page.keyboard.press('Enter')
    await expect(dialog.getByRole('textbox', { name: 'Nome', exact: true })).toHaveAttribute('readonly', '')
    await dialog.getByRole('textbox', { name: 'Valor (R$)' }).fill('200.01')
    await dialog.getByRole('spinbutton', { name: 'Dia de vencimento' }).fill('1')
    await dialog.getByRole('combobox', { name: 'Categoria' }).selectOption('OTHER')
    const update = page.waitForRequest((request) => /fixed-expenses\/\d+$/.test(request.url()) && request.method() === 'PATCH')
    await dialog.getByRole('button', { name: 'Salvar', exact: true }).click()
    expect((await update).postDataJSON()).toEqual({ amount: '200.01', due_day: 1, category: 'OTHER' })
    await expect(row).toContainText('200,01')
    await expect(edit).toBeFocused()
    const deactivate = row.getByRole('button', { name: 'Desativar Internet teste' })
    await deactivate.focus()
    await page.keyboard.press('Enter')
    const confirmation = page.getByRole('alertdialog')
    await expect(confirmation.getByRole('button', { name: 'Cancelar' })).toBeFocused()
    await confirmation.evaluate((element) => Promise.all(element.getAnimations().map((animation) => animation.finished)))
    await page.keyboard.press('Escape')
    await expect(deactivate).toBeFocused()
    await page.keyboard.press('Enter')
    const pause = page.waitForRequest((request) => request.url().endsWith('/deactivate') && request.method() === 'PATCH')
    await confirmation.getByRole('button', { name: 'Desativar', exact: true }).click()
    expect((await pause).postDataJSON()).toEqual({ status: 'PAUSED' })
    await expect(confirmation).toHaveCount(0)
    await expect(row).toHaveCount(0)
    await expect(newButton).toBeFocused()
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
    if (width === 375) {
      const table = page.getByRole('table', { name: 'Despesas fixas da guild' })
      await table.evaluate((element) => element.parentElement!.focus())
      await page.keyboard.press('ArrowRight')
      await expect.poll(() => table.evaluate((element) => element.parentElement!.scrollLeft)).toBeGreaterThan(0)
      expect(await table.evaluate((element) => element.parentElement!.scrollWidth > element.parentElement!.clientWidth)).toBe(true)
    }
    await page.screenshot({ path: `test-results/fixed-expenses-${width}.png`, fullPage: true })
  })
}

test('forbidden expense action preserves the draft and session', async ({ page }) => {
  await page.addInitScript(() => {
    const original = window.fetch.bind(window)
    window.fetch = (input, options) => {
      const path = new URL(input instanceof Request ? input.url : String(input), location.origin).pathname
      const method = options?.method ?? (input instanceof Request ? input.method : 'GET')
      if (path === '/api/v1/guilds/1/fixed-expenses' && method === 'POST') return Promise.resolve(new Response('{"error":"forbidden"}', { status: 403, headers: { 'Content-Type': 'application/json' } }))
      return original(input, options)
    }
  })
  await page.goto('/fixed-expenses')
  await page.getByRole('button', { name: 'Nova despesa fixa' }).click()
  const dialog = page.getByRole('dialog')
  await dialog.getByRole('textbox', { name: 'Nome', exact: true }).fill('Draft preserved')
  await dialog.getByRole('textbox', { name: 'Valor (R$)' }).fill('123,45')
  await dialog.getByRole('button', { name: 'Salvar', exact: true }).click()
  await expect(dialog.getByRole('alert')).toHaveText('Você não tem permissão para realizar esta ação.')
  await expect(dialog.getByRole('textbox', { name: 'Nome', exact: true })).toHaveValue('Draft preserved')
  await expect(dialog.getByRole('textbox', { name: 'Valor (R$)' })).toHaveValue('123,45')
  await expect(page).toHaveURL(/\/fixed-expenses/)
  await expect(page.getByRole('heading', { name: 'Acesso negado' })).toHaveCount(0)
  await dialog.getByRole('button', { name: 'Cancelar', exact: true }).click()
  await expect(page.getByRole('button', { name: 'Sair' })).toBeEnabled()
})

// Existing business journeys begin with an explicit saved guild preference.
test.beforeEach(async ({ page }) => { await page.addInitScript(() => localStorage.setItem('guildbanker:last-guild:demo:1', '1')) })
