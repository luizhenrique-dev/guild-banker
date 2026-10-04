'use client'
// ============================================================
// 📌 PADRÃO: Modal genérico com Radix Dialog
// Usa @radix-ui/react-dialog já instalado no projeto.
// ============================================================
import * as Dialog from '@radix-ui/react-dialog'
import { X } from 'lucide-react'
import { useRef, type ReactNode } from 'react'
import { useTranslation } from 'react-i18next'

interface ModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  title: string
  children: ReactNode
  maxWidth?: string
}

export function Modal({ open, onOpenChange, title, children, maxWidth = 'max-w-lg' }: ModalProps) {
  const { t } = useTranslation()
  const content = useRef<HTMLDivElement | null>(null)
  const returnFocus = useRef<HTMLElement | null>(null)
  return (
    <Dialog.Root open={open} onOpenChange={onOpenChange}>
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-40 bg-black/40 data-[state=open]:animate-fade-in" />
        <Dialog.Content
          ref={content}
          aria-describedby={undefined}
          onOpenAutoFocus={(event) => {
            returnFocus.current = document.activeElement as HTMLElement
            const input = content.current?.querySelector<HTMLInputElement>('input:not([readonly])')
            if (input) { event.preventDefault(); input.focus() }
          }}
          onCloseAutoFocus={(event) => { event.preventDefault(); returnFocus.current?.focus() }}
          className={`fixed left-1/2 top-1/2 z-50 w-[calc(100%-2rem)] max-h-[calc(100dvh-2rem)] overflow-y-auto ${maxWidth} -translate-x-1/2 -translate-y-1/2 rounded-2xl bg-card p-6 shadow-xl data-[state=open]:animate-fade-in`}
        >
          <div className="mb-4 flex items-center justify-between">
            <Dialog.Title className="text-lg font-semibold text-card-foreground">
              {title}
            </Dialog.Title>
            <Dialog.Close asChild>
              <button
                type="button"
                aria-label={t('common.close')}
                className="inline-flex h-8 w-8 items-center justify-center rounded-full text-muted-foreground transition-colors hover:bg-muted hover:text-card-foreground"
              >
                <X size={18} />
              </button>
            </Dialog.Close>
          </div>
          {children}
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  )
}
