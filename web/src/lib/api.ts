import axios, { AxiosError, type InternalAxiosRequestConfig } from "axios"

import { clearSession, loadSession, saveSession } from "@/lib/auth"

export const api = axios.create({
  baseURL: `${import.meta.env.VITE_API_URL}/api`,
  withCredentials: false,
})

// Attach the access token on every outgoing request when a session exists.
api.interceptors.request.use((config) => {
  const session = loadSession()
  if (session) {
    config.headers.Authorization = `Bearer ${session.accessToken}`
  }
  return config
})

// On 401, try refreshing once. If that also fails, clear the session and
// bounce to login. The refresh request itself is exempt so we don't loop.
type RetriedConfig = InternalAxiosRequestConfig & { _retried?: boolean }

let refreshing: Promise<string | null> | null = null

api.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const original = error.config as RetriedConfig | undefined
    const status = error.response?.status
    const url = original?.url ?? ""

    if (status !== 401 || !original || original._retried || url.includes("/auth/refresh")) {
      return Promise.reject(error)
    }
    original._retried = true

    const session = loadSession()
    if (!session?.refreshToken) {
      clearSession()
      window.location.href = "/auth/login"
      return Promise.reject(error)
    }

    // Coalesce concurrent refreshes into one network call.
    refreshing ??= (async () => {
      try {
        const res = await api.post<{
          access_token: string
          refresh_token: string
          expires_in: number
          entity_id: string
        }>("/auth/refresh", { refresh_token: session.refreshToken })
        saveSession({
          accessToken: res.data.access_token,
          refreshToken: res.data.refresh_token,
          expiresIn: res.data.expires_in,
          entityId: res.data.entity_id,
        })
        return res.data.access_token
      } catch {
        return null
      } finally {
        refreshing = null
      }
    })()

    const newAccessToken = await refreshing
    if (!newAccessToken) {
      clearSession()
      window.location.href = "/auth/login"
      return Promise.reject(error)
    }
    original.headers.Authorization = `Bearer ${newAccessToken}`
    return api(original)
  },
)
