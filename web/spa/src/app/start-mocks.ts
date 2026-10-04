export async function startMocks(dev: boolean, optIn: string | undefined): Promise<boolean> {
  if (!dev || optIn !== 'true') return false
  const { worker } = await import('../mocks/browser')
  await worker.start({
    serviceWorker: { url: '/mockServiceWorker.js' },
    onUnhandledRequest(request, print) {
      if (new URL(request.url).pathname.startsWith('/api/')) print.error()
    },
  })
  return true
}
