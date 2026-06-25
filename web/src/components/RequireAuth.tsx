import { Navigate, Outlet, useLocation } from "react-router-dom"

import { Skeleton } from "@/components/ui/skeleton"
import { useAuth } from "@/lib/auth"

export function RequireAuth() {
  const location = useLocation()
  const { isAuthenticated, isLoading } = useAuth()

  if (isLoading) {
    return (
      <main className="flex min-h-svh items-center justify-center bg-background px-4">
        <Skeleton className="h-10 w-64 rounded-lg" />
      </main>
    )
  }

  if (!isAuthenticated) {
    return <Navigate to="/auth/login" state={{ from: location }} replace />
  }
  return <Outlet />
}
