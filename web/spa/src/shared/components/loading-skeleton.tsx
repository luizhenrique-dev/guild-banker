'use client'

export function LoadingSkeleton({ rows = 3 }: { rows?: number }) {
  return (
    <div className="space-y-3">
      {Array.from({ length: rows }).map((_, i) => (
        <div
          key={i}
          className="h-16 animate-pulse rounded-xl bg-[#eef0f3]"
        />
      ))}
    </div>
  )
}

export function TableSkeleton({ rows = 3 }: { rows?: number }) {
  return (
    <div className="overflow-hidden rounded-[24px] border border-[#dee1e6] bg-white">
      <div className="border-b border-[#dee1e6] bg-[#f7f7f7] px-6 py-3">
        <div className="h-4 w-48 animate-pulse rounded bg-[#dee1e6]" />
      </div>
      {Array.from({ length: rows }).map((_, i) => (
        <div key={i} className="flex items-center gap-6 border-b border-[#eef0f3] px-6 py-4">
          <div className="h-9 w-9 animate-pulse rounded-full bg-[#eef0f3]" />
          <div className="h-4 w-32 animate-pulse rounded bg-[#eef0f3]" />
          <div className="h-4 w-20 animate-pulse rounded bg-[#eef0f3]" />
          <div className="ml-auto h-4 w-24 animate-pulse rounded bg-[#eef0f3]" />
        </div>
      ))}
    </div>
  )
}
