import type { ReactNode } from "react"

import { cn } from "@/lib/utils"

export function PageContainer({
  children,
  className,
}: {
  children: ReactNode
  className?: string
}) {
  return (
    <main className={cn("mx-auto w-full max-w-6xl px-4 py-8 lg:px-8", className)}>
      {children}
    </main>
  )
}

export function PageHeader({
  title,
  description,
  action,
}: {
  title: string
  description?: string
  action?: ReactNode
}) {
  return (
    <div className="mb-8 flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">{title}</h1>
        {description && <p className="mt-1 text-sm text-muted-foreground">{description}</p>}
      </div>
      {action}
    </div>
  )
}
