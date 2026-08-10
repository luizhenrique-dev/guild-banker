'use client'
import { Component, type ReactNode } from 'react'
import { AlertTriangle } from 'lucide-react'

interface Props {
  children: ReactNode
  fallback?: ReactNode
}

interface State {
  hasError: boolean
  error: Error | null
}

// 📌 PADRÃO: ErrorBoundary clássico (class component é obrigatório para getDerivedStateFromError)
export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.state = { hasError: false, error: null }
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, info: React.ErrorInfo) {
    console.error('[ErrorBoundary]', error, info)
  }

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) return this.props.fallback
      return (
        <div className="flex flex-col items-center justify-center gap-4 py-16 text-center">
          <AlertTriangle size={40} className="text-[#cf202f]" />
          <h2 className="text-lg font-semibold text-[#0a0b0d]">Algo deu errado</h2>
          <p className="max-w-md text-sm text-[#5b616e]">
            {this.state.error?.message ?? 'Erro inesperado'}
          </p>
          <button
            type="button"
            onClick={() => this.setState({ hasError: false, error: null })}
            className="rounded-full bg-[#0052ff] px-5 py-2 text-sm font-semibold text-white hover:bg-[#003ecc]"
          >
            Tentar novamente
          </button>
        </div>
      )
    }

    return this.props.children
  }
}
