import { useQuery, useQueryClient } from "@tanstack/react-query"

import { api } from "@/lib/api"

const SESSION_KEY = "vault_session"

export type Session = {
  accessToken: string
  refreshToken: string
  expiresIn: number
}

export type CurrentSession = {
  entity_id: string
  user_id: string
  scope: string
  groups: string[]
}

export function saveSession(session: Session) {
  localStorage.setItem(SESSION_KEY, JSON.stringify(session))
}

export function loadSession(): Session | null {
  const raw = localStorage.getItem(SESSION_KEY)
  if (!raw) return null
  try {
    const session = JSON.parse(raw) as Partial<Session>
    if (!session.accessToken || !session.refreshToken) {
      localStorage.removeItem(SESSION_KEY)
      return null
    }
    return {
      accessToken: session.accessToken,
      refreshToken: session.refreshToken,
      expiresIn: Number(session.expiresIn) || 0,
    }
  } catch {
    localStorage.removeItem(SESSION_KEY)
    return null
  }
}

export function clearSession() {
  localStorage.removeItem(SESSION_KEY)
}

export function useAuth() {
  const queryClient = useQueryClient()
  const tokenSession = loadSession()
  const sessionQuery = useQuery({
    queryKey: ["session", tokenSession?.accessToken],
    queryFn: async () => {
      const response = await api.get<CurrentSession>("/auth/session")
      return response.data
    },
    enabled: !!tokenSession?.accessToken,
    retry: false,
    staleTime: 5 * 60 * 1000,
  })

  function logout() {
    clearSession()
    queryClient.clear()
    window.location.href = "/auth/login"
  }

  return {
    session: sessionQuery.data,
    tokenSession,
    isLoading: sessionQuery.isLoading,
    isAuthenticated: !!tokenSession,
    logout,
  }
}
