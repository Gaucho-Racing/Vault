import { ArrowRight, KeyRound, Loader2 } from "lucide-react"
import { useEffect, useRef, useState } from "react"
import { useLocation, useNavigate, useSearchParams } from "react-router-dom"

import { Button } from "@/components/ui/button"
import { api } from "@/lib/api"
import { clearSession, saveSession } from "@/lib/auth"

const sentinelURL = import.meta.env.VITE_SENTINEL_URL ?? "https://sentinel-v5.gauchoracing.com"
const sentinelClientID = import.meta.env.VITE_SENTINEL_CLIENT_ID ?? "cfqmh6bfYFe8"
const oauthScope = "user:read groups:read"

type TokenResponse = {
  access_token: string
  refresh_token: string
  expires_in: number
}

export default function LoginPage() {
  const navigate = useNavigate()
  const location = useLocation()
  const [searchParams] = useSearchParams()
  const exchangedRef = useRef(false)
  const [loading, setLoading] = useState(searchParams.has("code"))
  const [errorMessage, setErrorMessage] = useState("")
  const fromLocation = (
    location.state as {
      from?: { pathname: string; search?: string; hash?: string }
    } | null
  )?.from
  const from = fromLocation
    ? `${fromLocation.pathname}${fromLocation.search ?? ""}${fromLocation.hash ?? ""}`
    : "/accounts"

  useEffect(() => {
    if (exchangedRef.current) return
    const code = searchParams.get("code")
    if (!code) return
    exchangedRef.current = true

    void (async () => {
      setLoading(true)
      try {
        const response = await api.post<TokenResponse>(
          `/auth/login?code=${encodeURIComponent(code)}`,
        )
        saveSession({
          accessToken: response.data.access_token,
          refreshToken: response.data.refresh_token,
          expiresIn: response.data.expires_in,
        })
        navigate(sanitizeReturnTo(searchParams.get("state") || from), { replace: true })
      } catch (error) {
        clearSession()
        const message =
          (error as { response?: { data?: { error?: string; message?: string } } })?.response
            ?.data?.error ??
          (error as { response?: { data?: { error?: string; message?: string } } })?.response
            ?.data?.message ??
          "Sentinel sign-on failed. Please try again."
        setErrorMessage(message)
        setLoading(false)
        navigate("/auth/login?error=oauth", { replace: true })
      }
    })()
  }, [from, navigate, searchParams])

  function login() {
    const params = new URLSearchParams({
      client_id: sentinelClientID,
      response_type: "code",
      redirect_uri: `${window.location.origin}/auth/login`,
      scope: oauthScope,
      prompt: "none",
      state: sanitizeReturnTo(from),
    })
    window.location.href = `${sentinelURL.replace(/\/+$/, "")}/oauth/authorize?${params.toString()}`
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

        {loading && (
          <div className="flex items-center justify-center py-6">
            <Loader2 className="size-8 animate-spin text-muted-foreground" />
          </div>
        )}

        {!loading && (searchParams.get("error") || errorMessage) && (
          <div className="rounded-lg border border-destructive/40 bg-destructive/10 px-3 py-2 text-sm text-destructive">
            {errorMessage || "Sentinel sign-on failed. Please try again."}
          </div>
        )}

        {!loading && (
          <Button type="button" className="h-10 w-full" onClick={login}>
            Sentinel Sign On
            <ArrowRight className="size-4" />
          </Button>
        )}
      </div>
    </main>
  )
}

function sanitizeReturnTo(value: string) {
  if (!value || !value.startsWith("/") || value.startsWith("//") || value.startsWith("/api/")) {
    return "/accounts"
  }
  return value
}
