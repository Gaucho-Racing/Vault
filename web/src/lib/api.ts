import axios, { AxiosError, type InternalAxiosRequestConfig } from "axios"

export const apiBaseURL = `${import.meta.env.VITE_API_URL ?? ""}/api`

export const api = axios.create({
  baseURL: apiBaseURL,
  withCredentials: true,
})

type RetriedConfig = InternalAxiosRequestConfig & { _retried?: boolean }

api.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const original = error.config as RetriedConfig | undefined
    const status = error.response?.status
    const url = original?.url ?? ""
    const isAuthEndpoint = url.includes("/auth/login") || url.includes("/auth/callback") || url.includes("/auth/refresh")

    if (status !== 401 || !original || original._retried || isAuthEndpoint) {
      return Promise.reject(error)
    }
    original._retried = true

    try {
      await api.post("/auth/refresh")
      return api(original)
    } catch {
      window.location.href = "/auth/login"
      return Promise.reject(error)
    }
  },
)
