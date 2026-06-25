import { KeyRound } from "lucide-react"
import { Link, useLocation } from "react-router-dom"

import { cn } from "@/lib/utils"

const navItems = [{ to: "/accounts", label: "Accounts", icon: KeyRound }]

function isActive(currentPath: string, target: string) {
  return currentPath === target || currentPath.startsWith(`${target}/`)
}

export function AppSidebar() {
  const { pathname } = useLocation()

  return (
    <aside className="hidden border-r bg-sidebar text-sidebar-foreground lg:flex lg:min-h-svh lg:flex-col">
      <div className="flex h-14 items-center gap-2 px-4">
        <div className="flex size-8 items-center justify-center rounded-lg bg-gradient-to-br from-gr-pink to-gr-purple text-white">
          <KeyRound className="size-4" />
        </div>
        <div className="min-w-0">
          <div className="text-sm font-semibold leading-none">Vault</div>
          <div className="mt-1 text-xs text-muted-foreground">Gaucho Racing</div>
        </div>
      </div>
      <nav className="flex-1 px-2 py-3">
        {navItems.map((item) => {
          const active = isActive(pathname, item.to)
          return (
            <Link
              key={item.to}
              to={item.to}
              className={cn(
                "flex h-9 items-center gap-2 rounded-lg px-2.5 text-sm transition-colors",
                active
                  ? "bg-sidebar-accent text-sidebar-accent-foreground"
                  : "text-muted-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground",
              )}
            >
              <item.icon className="size-4" />
              <span>{item.label}</span>
            </Link>
          )
        })}
      </nav>
      <div className="border-t px-4 py-3 text-xs text-muted-foreground">
        <span>Gaucho Racing · </span>
        <span className="font-mono">v0.1.0</span>
      </div>
    </aside>
  )
}
