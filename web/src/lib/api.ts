import axios, { AxiosError } from "axios"

import { clearSession, loadSession } from "@/lib/auth"

export const api = axios.create({
  baseURL: `${import.meta.env.VITE_API_URL ?? ""}/api`,
  withCredentials: false,
})

api.interceptors.request.use((config) => {
  const session = loadSession()
  if (session?.accessToken) {
    config.headers.Authorization = `Bearer ${session.accessToken}`
  }
  return config
})

api.interceptors.response.use(
  (response) => response,
  (error: AxiosError) => {
    if (error.response?.status === 401 && !window.location.pathname.startsWith("/auth/login")) {
      clearSession()
      window.location.href = "/auth/login"
    }
    return Promise.reject(error)
  },
)
