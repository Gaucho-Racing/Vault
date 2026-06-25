import { useQueryClient } from "@tanstack/react-query"

const SESSION_KEY = "vault_session"

export type Session = {
  accessToken: string
}

export function saveSession(session: Session) {
  localStorage.setItem(SESSION_KEY, JSON.stringify(session))
}

export function loadSession(): Session | null {
  const raw = localStorage.getItem(SESSION_KEY)
  if (!raw) return null
  try {
    return JSON.parse(raw) as Session
  } catch {
    return null
  }
}

export function clearSession() {
  localStorage.removeItem(SESSION_KEY)
}

export function useAuth() {
  const queryClient = useQueryClient()
  const session = loadSession()

  function logout() {
    clearSession()
    queryClient.clear()
    window.location.href = "/auth/login"
  }

  return {
    session,
    isAuthenticated: !!session?.accessToken,
    logout,
  }
}
