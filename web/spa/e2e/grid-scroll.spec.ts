import { expect, test } from '@playwright/test'

for (const route of ['fixed-expenses', 'transactions', 'import']) {
  test(`${route} scrolls rows internally and adapts to the viewport`, async ({ page }) => {
    await page.addInitScript(() => {
      localStorage.setItem('guildbanker:last-guild:demo:1', '1')
      const original = window.fetch.bind(window)
      window.fetch = async (input, options) => {
        const response = await original(input, options)
        const path = new URL(input instanceof Request ? input.url : String(input), location.origin).pathname
        const method = options?.method ?? (input instanceof Request ? input.method : 'GET')
        if (!response.ok) return response
        const fixedExpenses = path.endsWith('/fixed-expenses') && method === 'GET'
        const transactions = path.endsWith('/transactions') && method === 'GET'
        const imports = path.endsWith('/imports') && method === 'POST'
        if (!fixedExpenses && !transactions && !imports) return response
        const data = await response.json()
        const items = fixedExpenses ? data : data.items
        const expanded = Array.from({ length: 32 }, (_, index) => ({
          ...items[index % items.length], id: index + 1000, itemId: index + 1000,
        }))
        return new Response(JSON.stringify(fixedExpenses ? expanded : { ...data, items: expanded }), {
          status: response.status, headers: response.headers,
        })
      }
    })
    await page.goto(`/${route}?guild=1`)
    if (route === 'import') {
      await page.locator('input[type=file]').setInputFiles({ name: 'statement.csv', mimeType: 'text/csv', buffer: Buffer.from('csv') })
      await page.getByRole('button', { name: 'Enviar arquivo' }).click()
    }
    const table = page.getByRole('table')
    await expect(table.locator('tbody tr')).toHaveCount(32)
    let previousHeight = 0
    for (const viewport of [{ width: 1280, height: 900 }, { width: 1280, height: 640 }, { width: 375, height: 740 }]) {
      await page.setViewportSize(viewport)
      const scroll = table.locator('..')
      await scroll.evaluate((element) => { element.scrollTop = 0 })
      const height = await scroll.evaluate((element) => element.clientHeight)
      expect(height).toBeGreaterThan(50)
      if (viewport.height === 640) expect(height).toBeLessThan(previousHeight)
      previousHeight = height
      expect(await page.evaluate(() => document.documentElement.scrollHeight <= innerHeight + 1)).toBe(true)
      expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth + 1)).toBe(true)
      expect(await page.locator('main').evaluate((element) => element.parentElement!.parentElement!.scrollHeight <= element.parentElement!.parentElement!.clientHeight + 1)).toBe(true)
      await scroll.evaluate((element) => { element.scrollTop = element.scrollHeight })
      expect(await scroll.evaluate((element) => element.scrollTop)).toBeGreaterThan(0)
      const header = await table.locator('thead').boundingBox()
      const box = await scroll.boundingBox()
      expect(Math.abs(header!.y - box!.y)).toBeLessThan(2)
      const last = await table.locator('tbody tr').last().boundingBox()
      expect(last!.y + last!.height).toBeLessThanOrEqual(box!.y + box!.height + 1)
    }
  })
}
