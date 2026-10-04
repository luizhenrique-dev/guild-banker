'use client'

export function LoadingSkeleton({ rows = 3 }: { rows?: number }) {
  return (
    <div className="space-y-3">
      {Array.from({ length: rows }).map((_, i) => (
        <div
          key={i}
          className="h-16 animate-pulse rounded-xl bg-secondary"
        />
      ))}
    </div>
  )
}

export function TableSkeleton({ rows = 3 }: { rows?: number }) {
  return (
    <div className="overflow-hidden rounded-[24px] border border-border bg-card">
      <div className="border-b border-border bg-muted px-6 py-3">
        <div className="h-4 w-48 animate-pulse rounded bg-border" />
      </div>
      {Array.from({ length: rows }).map((_, i) => (
        <div key={i} className="flex items-center gap-6 border-b border-border/50 px-6 py-4">
          <div className="h-9 w-9 animate-pulse rounded-full bg-secondary" />
          <div className="h-4 w-32 animate-pulse rounded bg-secondary" />
          <div className="h-4 w-20 animate-pulse rounded bg-secondary" />
          <div className="ml-auto h-4 w-24 animate-pulse rounded bg-secondary" />
        </div>
      ))}
    </div>
  )
}
