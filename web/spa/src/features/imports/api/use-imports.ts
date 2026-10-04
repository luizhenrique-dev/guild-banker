import { useEffect, useMemo } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useSession } from '@/src/app/session-provider'
import { importsApi } from './imports.api'
import { transactionKeys } from '../../transactions/api/use-transactions'
import type { UploadImportResponse } from '../types'
import { toast } from 'sonner'
import i18n from '@/src/shared/lib/i18n'

export function useImports(guildID: number) {
  const session = useSession()
  const qc = useQueryClient()
  const scope = useMemo(() => ({ guildID, controller: new AbortController() }), [guildID])
  const { controller } = scope
  const key = ['imports', session.user?.id, guildID] as const
  const revision = useQuery<UploadImportResponse | null>({ queryKey: key, queryFn: () => null, enabled: false, initialData: null })
  useEffect(() => () => {
    controller.abort()
    qc.removeQueries({ queryKey: ['imports', session.user?.id, guildID], exact: true })
  }, [controller, qc, session.user?.id, guildID])
  const active = () => session.active() && !controller.signal.aborted
  const capture = () => ({ active, sessionActive: session.active, key, userID: session.user?.id, guildID })
  const upload = useMutation({
    onMutate: capture,
    mutationFn: (file: File) => session.request((options) => importsApi.upload(guildID, file, options), controller.signal),
    onSuccess: (result, _file, scope) => {
      if (!scope?.active()) return
      qc.setQueryData(scope.key, result)
      toast.success(i18n.t('import.toast.uploaded'))
    },
    onError: (_error, _input, scope) => { if (scope?.active()) toast.error(i18n.t('fx.toast.error')) },
  })
  const confirm = useMutation({
    onMutate: capture,
    mutationFn: (importID: number) => session.request((options) => importsApi.confirm(guildID, importID, options), controller.signal),
    onSuccess: (result, _id, scope) => {
      if (!scope?.active()) return
      qc.setQueryData(scope.key, null)
      toast.success(i18n.t('import.toast.confirmed', { created: result.created }))
    },
    onError: (_error, _input, scope) => { if (scope?.active()) toast.error(i18n.t('fx.toast.error')) },
    onSettled: (_result, _error, _id, scope) => {
      // Cancellation cannot undo a confirmation already committed by the server.
      if (scope?.sessionActive()) void qc.invalidateQueries({ queryKey: transactionKeys.lists(scope.userID, scope.guildID) })
    },
  })
  return { revision: revision.data, upload, confirm }
}
