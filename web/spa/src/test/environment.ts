import { builtinEnvironments, type Environment } from 'vitest/environments'

// jsdom supplies the DOM; keep fetch and its Web API types in the Node realm.
export default {
  name: 'jsdom-fetch',
  transformMode: 'web',
  async setup(global, options) {
    const webAPIs = {
      AbortController: global.AbortController, AbortSignal: global.AbortSignal,
      FormData: global.FormData, Blob: global.Blob, File: global.File,
    }
    const environment = await builtinEnvironments.jsdom.setup(global, options)
    Object.assign(global, webAPIs)
    return environment
  },
} satisfies Environment
