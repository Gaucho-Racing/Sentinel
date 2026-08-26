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

    // A 401 from these endpoints means "the credentials you just sent are
    // wrong," not "your bearer expired." Propagate so the caller can toast.
    const isAuthEndpoint = url.includes("/auth/login") || url.includes("/auth/refresh")

    if (status !== 401 || !original || original._retried || isAuthEndpoint) {
      return Promise.reject(error)
    }
    original._retried = true

    const session = loadSession()
    if (!session?.refreshToken) {
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

// Guard against 2xx responses that aren't JSON. Every endpoint this client
// talks to returns JSON, so an HTML body means the request never reached the
// API: the gateway has no route for that path, so it fell through to the SPA
// and nginx answered index.html with a 200. Without this, axios resolves those
// as success and callers read fields off an HTML string — a missing gateway
// route surfaces as a silently blank value instead of an error.
api.interceptors.response.use((response) => {
  const contentType = response.headers["content-type"]
  const hasBody =
    response.status !== 204 && response.data !== "" && response.data != null
  if (hasBody && typeof contentType === "string" && !contentType.includes("json")) {
    return Promise.reject(
      new AxiosError(
        `Expected JSON from ${response.config.url ?? "the API"} but got "${contentType}" ` +
          `(HTTP ${response.status}). The API gateway is probably missing a route for this path.`,
        AxiosError.ERR_BAD_RESPONSE,
        response.config,
        response.request,
        response,
      ),
    )
  }
  return response
})
