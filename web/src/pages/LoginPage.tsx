import { Loader2 } from "lucide-react"
import { useState } from "react"
import { useLocation, useNavigate } from "react-router-dom"
import { toast } from "sonner"

import { Button } from "@/components/ui/button"
import { Textarea } from "@/components/ui/textarea"
import { api } from "@/lib/api"
import { clearSession, saveSession } from "@/lib/auth"

export default function LoginPage() {
  const navigate = useNavigate()
  const location = useLocation()
  const fromLocation = (
    location.state as {
      from?: { pathname: string; search?: string; hash?: string }
    } | null
  )?.from
  const from = fromLocation
    ? `${fromLocation.pathname}${fromLocation.search ?? ""}${fromLocation.hash ?? ""}`
    : "/accounts"
  const [token, setToken] = useState("")
  const [loading, setLoading] = useState(false)

  async function handleSubmit(event: React.FormEvent) {
    event.preventDefault()
    if (!token.trim()) return
    setLoading(true)
    saveSession({ accessToken: token.trim() })
    try {
      await api.get("/accounts")
      navigate(from, { replace: true })
    } catch (error) {
      clearSession()
      const message =
        (error as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        "Sentinel rejected that token."
      toast.error(message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <main className="flex min-h-svh items-center justify-center px-4 py-12">
      <form onSubmit={handleSubmit} className="w-full max-w-sm space-y-7">
        <div className="flex flex-col items-center gap-3 text-center">
          <img src="/favicon.svg" alt="Vault" className="size-12 rounded-xl" />
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">Sign in to Vault</h1>
            <p className="mt-1 text-sm text-muted-foreground">Use a Sentinel bearer token.</p>
          </div>
        </div>

        <Textarea
          value={token}
          onChange={(event) => setToken(event.target.value)}
          placeholder="eyJhbGciOi..."
          rows={7}
          className="font-mono text-xs"
          required
        />

        <Button type="submit" className="h-10 w-full" disabled={loading}>
          {loading && <Loader2 className="size-4 animate-spin" />}
          Sign in
        </Button>
      </form>
    </main>
  )
}
