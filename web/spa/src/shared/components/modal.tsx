'use client'
// ============================================================
// 📌 PADRÃO: Modal genérico com Radix Dialog
// Usa @radix-ui/react-dialog já instalado no projeto.
// ============================================================
import * as Dialog from '@radix-ui/react-dialog'
import { X } from 'lucide-react'
import type { ReactNode } from 'react'

interface ModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  title: string
  children: ReactNode
  maxWidth?: string
}

export function Modal({ open, onOpenChange, title, children, maxWidth = 'max-w-lg' }: ModalProps) {
  return (
    <Dialog.Root open={open} onOpenChange={onOpenChange}>
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-40 bg-black/40 data-[state=open]:animate-fade-in" />
        <Dialog.Content
          className={`fixed left-1/2 top-1/2 z-50 w-full ${maxWidth} -translate-x-1/2 -translate-y-1/2 rounded-2xl bg-white p-6 shadow-xl data-[state=open]:animate-fade-in`}
        >
          <div className="mb-4 flex items-center justify-between">
            <Dialog.Title className="text-lg font-semibold text-[#0a0b0d]">
              {title}
            </Dialog.Title>
            <Dialog.Close asChild>
              <button
                type="button"
                className="inline-flex h-8 w-8 items-center justify-center rounded-full text-[#7c828a] transition-colors hover:bg-[#f7f7f7] hover:text-[#0a0b0d]"
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
