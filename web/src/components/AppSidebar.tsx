import { KeyRound, LockKeyhole, ShieldCheck } from "lucide-react"
import { Link, useLocation } from "react-router-dom"

import { cn } from "@/lib/utils"

const navItems = [{ to: "/accounts", label: "Accounts", icon: LockKeyhole }]

function isActive(currentPath: string, target: string) {
  return currentPath === target || currentPath.startsWith(`${target}/`)
}

export function AppSidebar() {
  const { pathname } = useLocation()

  return (
    <aside className="hidden border-r border-sidebar-border/70 bg-sidebar text-sidebar-foreground lg:flex lg:min-h-svh lg:flex-col">
      <div className="flex h-16 items-center gap-3 border-b border-sidebar-border/70 px-4">
        <div className="flex size-9 items-center justify-center rounded-lg bg-primary text-primary-foreground shadow-sm shadow-primary/25">
          <KeyRound className="size-5" />
        </div>
        <div className="min-w-0">
          <div className="text-base font-semibold leading-none">Vault</div>
          <div className="mt-1 text-xs text-muted-foreground">Gaucho Racing</div>
        </div>
      </div>
      <nav className="flex-1 space-y-1 px-3 py-4">
        {navItems.map((item) => {
          const active = isActive(pathname, item.to)
          return (
            <Link
              key={item.to}
              to={item.to}
              className={cn(
                "flex h-10 items-center gap-2 rounded-lg px-3 text-sm font-medium transition-colors",
                active
                  ? "bg-primary text-primary-foreground shadow-sm shadow-primary/15"
                  : "text-muted-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground",
              )}
            >
              <item.icon className="size-4" />
              <span>{item.label}</span>
            </Link>
          )
        })}
      </nav>
      <div className="border-t border-sidebar-border/70 p-4">
        <div className="rounded-lg bg-background/55 p-3 shadow-sm shadow-black/[0.03] dark:bg-background/35 dark:shadow-black/20">
          <div className="flex items-center gap-2 text-xs font-medium text-sidebar-foreground">
            <ShieldCheck className="size-3.5 text-primary" />
            Sentinel protected
          </div>
          <div className="mt-1 font-mono text-xs text-muted-foreground">v0.1.0</div>
        </div>
      </div>
    </aside>
  )
}
