import { Outlet } from "react-router-dom"

import { AppHeader } from "@/components/AppHeader"
import { AppSidebar } from "@/components/AppSidebar"
import { TooltipProvider } from "@/components/ui/tooltip"

export function AppShell() {
  return (
    <TooltipProvider>
      <div className="grid min-h-svh bg-background text-foreground lg:grid-cols-[264px_minmax(0,1fr)]">
        <AppSidebar />
        <div className="min-w-0 bg-background">
          <AppHeader />
          <Outlet />
        </div>
      </div>
    </TooltipProvider>
  )
}
