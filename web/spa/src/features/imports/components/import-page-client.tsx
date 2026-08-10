'use client'
import { useState, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { FileUp, Upload, CheckCircle, Trash2, AlertTriangle } from 'lucide-react'
import { apiClient, getActiveGuildId } from '@/src/shared/lib/api-client'
import { toast } from 'sonner'
import type { UploadImportResponse, ImportItemResponse } from '../types'

export function ImportPageClient() {
  const { t } = useTranslation()
  const [file, setFile] = useState<File | null>(null)
  const [dragOver, setDragOver] = useState(false)
  const [uploading, setUploading] = useState(false)
  const [importResult, setImportResult] = useState<UploadImportResponse | null>(null)
  const [confirming, setConfirming] = useState(false)

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    setDragOver(false)
    const droppedFile = e.dataTransfer.files?.[0]
    if (droppedFile?.name?.endsWith('.csv')) {
      setFile(droppedFile)
    }
  }, [])

  const handleUpload = async () => {
    if (!file) return
    setUploading(true)
    try {
      const formData = new FormData()
      formData.append('file', file)
      const guildId = getActiveGuildId()
      const { data } = await apiClient.post<UploadImportResponse>(
        `/api/v1/guilds/${guildId}/imports`,
        formData,
        { headers: { 'Content-Type': 'multipart/form-data' } }
      )
      setImportResult(data)
      toast.success(t('import.toast.uploaded'))
    } catch {
      toast.error(t('fx.toast.error'))
    } finally {
      setUploading(false)
    }
  }

  const handleConfirm = async () => {
    if (!importResult) return
    setConfirming(true)
    try {
      const guildId = getActiveGuildId()
      const { data } = await apiClient.post(
        `/api/v1/guilds/${guildId}/imports/${importResult.importId}:confirm`
      )
      const created = (data as { created?: number })?.created ?? 0
      toast.success(t('import.toast.confirmed', { created }))
      setImportResult(null)
      setFile(null)
    } catch {
      toast.error(t('fx.toast.error'))
    } finally {
      setConfirming(false)
    }
  }

  const statusBadge = (status: string) => {
    switch (status) {
      case 'READY':
        return <span className="inline-flex items-center gap-1 text-[12px] font-semibold text-[#05b169]"><CheckCircle size={12} /> Ready</span>
      case 'DUPLICATE':
        return <span className="inline-flex items-center gap-1 text-[12px] font-semibold text-amber-600"><AlertTriangle size={12} /> Duplicata</span>
      case 'DISCARDED':
        return <span className="inline-flex items-center gap-1 text-[12px] font-semibold text-[#7c828a]"><Trash2 size={12} /> Descartado</span>
      default:
        return null
    }
  }

  return (
    <main className="flex-1 px-8 py-8">
      <div className="mb-6">
        <h1 className="text-[32px] font-normal tracking-[-0.8px]">{t('import.title')}</h1>
        <p className="text-[13px] text-[#5b616e]">{t('import.subtitle')}</p>
      </div>

      {/* Dropzone */}
      {!importResult && (
        <div
          onDragOver={(e) => { e.preventDefault(); setDragOver(true) }}
          onDragLeave={() => setDragOver(false)}
          onDrop={handleDrop}
          className={`mb-6 flex flex-col items-center justify-center gap-4 rounded-[24px] border-2 border-dashed bg-white py-16 transition-colors ${
            dragOver ? 'border-[#0052ff] bg-blue-50/30' : 'border-[#dee1e6]'
          }`}
        >
          <FileUp size={40} className="text-[#7c828a]" />
          <p className="text-[15px] text-[#5b616e]">{t('import.dropzone')}</p>
          <p className="text-[12px] text-[#7c828a]">{t('import.maxSize')}</p>
          <label className="cursor-pointer rounded-full bg-[#eef0f3] px-5 py-2 text-[14px] font-semibold text-[#5b616e] transition-colors hover:bg-[#dee1e6]">
            <input
              type="file"
              accept=".csv"
              className="hidden"
              onChange={(e) => {
                const f = e.target.files?.[0]
                if (f) setFile(f)
              }}
            />
            Selecionar arquivo
          </label>
          {file && (
            <div className="flex items-center gap-3">
              <span className="text-[14px] font-medium">{file.name}</span>
              <button
                type="button"
                onClick={handleUpload}
                disabled={uploading}
                className="flex items-center gap-2 rounded-full bg-[#0052ff] px-5 py-2 text-[14px] font-semibold text-white transition-colors hover:bg-[#003ecc] disabled:opacity-50"
              >
                <Upload size={16} />
                {uploading ? t('common.loading') : t('import.upload')}
              </button>
            </div>
          )}
        </div>
      )}

      {/* Tabela de revisão */}
      {importResult && (
        <>
          <div className="mb-4 flex items-center justify-between">
            <h2 className="text-lg font-semibold">{t('import.review')}</h2>
            <div className="flex items-center gap-2 text-[13px] text-[#5b616e]">
              <span>{importResult.summary?.candidates ?? 0} candidatos</span>
              <span>·</span>
              <span>{importResult.summary?.duplicates ?? 0} duplicatas</span>
            </div>
          </div>
          <section className="mb-6 overflow-hidden rounded-[24px] border border-[#dee1e6] bg-white">
            <table className="w-full text-left text-[14px]">
              <thead className="border-b border-[#dee1e6] bg-[#f7f7f7] text-[12px] uppercase tracking-wide text-[#7c828a]">
                <tr>
                  <th className="py-3 pl-6 font-semibold">Descrição</th>
                  <th className="py-3 font-semibold">Categoria</th>
                  <th className="py-3 text-right font-semibold">{t('fx.col.amount')}</th>
                  <th className="py-3 pl-6 font-semibold">Data</th>
                  <th className="py-3 font-semibold">{t('fx.col.status')}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-[#eef0f3]">
                {(importResult.items ?? []).map((item: ImportItemResponse) => (
                  <tr key={item.itemId} className={item.status === 'DISCARDED' ? 'opacity-40' : ''}>
                    <td className="py-4 pl-6 font-medium">{item.description}</td>
                    <td className="py-4">
                      <span className="rounded-full bg-[#eef0f3] px-2.5 py-1 text-[12px] font-semibold text-[#5b616e]">
                        {t(`txCategory.${item.category}`)}
                      </span>
                    </td>
                    <td className="py-4 text-right font-mono font-medium">
                      {new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(parseFloat(item.amount) || 0)}
                    </td>
                    <td className="py-4 pl-6 font-mono text-[13px] text-[#5b616e]">
                      {new Date(item.occurredAt).toLocaleDateString('pt-BR', { timeZone: 'UTC' })}
                    </td>
                    <td className="py-4">{statusBadge(item.status)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </section>
          <div className="flex justify-end">
            <button
              type="button"
              onClick={handleConfirm}
              disabled={confirming}
              className="flex items-center gap-2 rounded-full bg-[#0052ff] px-6 py-3 text-[14px] font-semibold text-white transition-colors hover:bg-[#003ecc] disabled:opacity-50"
            >
              <CheckCircle size={16} />
              {confirming ? t('common.loading') : t('import.confirm')}
            </button>
          </div>
        </>
      )}
    </main>
  )
}
