import { useQuery, useQueryClient } from "@tanstack/react-query"

import { api } from "@/lib/api"

export type Session = {
  entity_id: string
  user_id: string
  scope: string
  groups: string[]
}

export function useAuth() {
  const queryClient = useQueryClient()
  const sessionQuery = useQuery({
    queryKey: ["session"],
    queryFn: async () => {
      const response = await api.get<Session>("/auth/session")
      return response.data
    },
    retry: false,
  })

  async function logout() {
    await api.post("/auth/logout")
    queryClient.clear()
    window.location.href = "/auth/login"
  }

  return {
    session: sessionQuery.data,
    isLoading: sessionQuery.isLoading,
    isAuthenticated: !!sessionQuery.data,
    logout,
  }
}
