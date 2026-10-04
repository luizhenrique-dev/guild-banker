import { safeID } from '@/src/shared/api/id'
import { apiClient, type RequestOptions } from '@/src/shared/lib/api-client'
import { legacyRequestOptions } from '@/src/shared/lib/legacy-session'
import type { UploadImportResponse, ConfirmImportResponse } from '../types'

export const importsApi = {
  upload: (guildId: number, file: File, options: RequestOptions = legacyRequestOptions()) => {
    const body = new FormData(); body.append('file', file)
    return apiClient<UploadImportResponse>(`/api/v1/guilds/${safeID(guildId)}/imports`, { ...options, method: 'POST', body })
  },
  confirm: (guildId: number, importId: number, options: RequestOptions = legacyRequestOptions()) =>
    apiClient<ConfirmImportResponse>(`/api/v1/guilds/${safeID(guildId)}/imports/${safeID(importId)}/confirm`, { ...options, method: 'POST' }),
}
