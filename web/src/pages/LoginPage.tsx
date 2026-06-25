import { ArrowRight, KeyRound } from "lucide-react"
import { useLocation, useSearchParams } from "react-router-dom"

import { Button } from "@/components/ui/button"
import { apiBaseURL } from "@/lib/api"

export default function LoginPage() {
  const location = useLocation()
  const [searchParams] = useSearchParams()
  const fromLocation = (
    location.state as {
      from?: { pathname: string; search?: string; hash?: string }
    } | null
  )?.from
  const from = fromLocation
    ? `${fromLocation.pathname}${fromLocation.search ?? ""}${fromLocation.hash ?? ""}`
    : "/accounts"

  function login() {
    window.location.href = `${apiBaseURL}/auth/login?return_to=${encodeURIComponent(from)}`
  }

  return (
    <main className="flex min-h-svh items-center justify-center px-4 py-12">
      <div className="w-full max-w-sm space-y-7">
        <div className="flex flex-col items-center gap-3 text-center">
          <div className="flex size-12 items-center justify-center rounded-xl bg-gradient-to-br from-gr-pink to-gr-purple text-white">
            <KeyRound className="size-6" />
          </div>
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">Sign in to Vault</h1>
            <p className="mt-1 text-sm text-muted-foreground">Continue with your Sentinel account.</p>
          </div>
        </div>

        {searchParams.get("error") && (
          <div className="rounded-lg border border-destructive/40 bg-destructive/10 px-3 py-2 text-sm text-destructive">
            Sentinel sign-on failed. Please try again.
          </div>
        )}

        <Button type="button" className="h-10 w-full" onClick={login}>
          Sentinel Sign On
          <ArrowRight className="size-4" />
        </Button>
      </div>
    </main>
  )
}
