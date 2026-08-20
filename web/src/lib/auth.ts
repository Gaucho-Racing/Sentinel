import { useQuery, useQueryClient } from "@tanstack/react-query"

import { api } from "@/lib/api"

const SESSION_KEY = "sentinel_session"
const LOGIN_RETURN_TO_KEY = "sentinel_login_return_to"

function isSafeReturnTo(path: string): boolean {
  return path.startsWith("/") && !path.startsWith("//")
}

// The pre-login path (e.g. /oauth/authorize?…&redirect_uri=…) has to outlive
// the bounce to /auth/login. React Router location.state is dropped on the
// initial history.replace, a refresh, and the Discord round-trip — so both
// email and Discord login were landing back on a stripped authorize URL.
export function saveLoginReturnTo(path: string) {
  if (!isSafeReturnTo(path) || path === "/") return
  const existing = sessionStorage.getItem(LOGIN_RETURN_TO_KEY)
  // Don't clobber a full authorize URL with a pathname-only bounce.
  if (existing && existing.includes("?") && !path.includes("?")) return
  sessionStorage.setItem(LOGIN_RETURN_TO_KEY, path)
}

export function saveLoginReturnFrom(location: {
  pathname: string
  search?: string
  hash?: string
}) {
  let path = `${location.pathname}${location.search ?? ""}${location.hash ?? ""}`
  // On first load RR can report an empty search while the address bar still
  // has ?client_id=…&redirect_uri=…. Prefer the bar when pathnames match.
  if (
    !path.includes("?") &&
    window.location.search &&
    window.location.pathname === location.pathname
  ) {
    path = `${window.location.pathname}${window.location.search}${window.location.hash}`
  }
  saveLoginReturnTo(path)
}

export function peekLoginReturnTo(): string | null {
  const raw = sessionStorage.getItem(LOGIN_RETURN_TO_KEY)
  if (!raw || !isSafeReturnTo(raw)) return null
  return raw
}

export function consumeLoginReturnTo(): string {
  const value = peekLoginReturnTo() ?? "/"
  sessionStorage.removeItem(LOGIN_RETURN_TO_KEY)
  return value
}

export type ReturnLocation = { pathname: string; search: string; hash: string }

export function locationFromReturnPath(path: string): ReturnLocation {
  const url = new URL(isSafeReturnTo(path) ? path : "/", window.location.origin)
  return { pathname: url.pathname, search: url.search, hash: url.hash }
}

export function consumeLoginReturnLocation(): ReturnLocation {
  return locationFromReturnPath(consumeLoginReturnTo())
}

export type Session = {
  accessToken: string
  refreshToken: string
  expiresIn: number
  entityId: string
}

export function saveSession(s: Session) {
  localStorage.setItem(SESSION_KEY, JSON.stringify(s))
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

// API entity shape — mirror of core/model/entity.go. `service_account` is
// omitempty server-side; `user` only set when type === "USER".
export type Entity = {
  id: string
  type: "USER" | "SERVICE_ACCOUNT"
  created_at: string
  email_auth?: { entity_id: string; email: string; created_at: string }
  phone_auth?: { entity_id: string; phone_number: string; created_at: string }
  external_auths: Array<{
    entity_id: string
    provider: "DISCORD" | "GOOGLE" | "GITHUB"
    external_id: string
    created_at: string
  }>
  user?: {
    id: string
    entity_id: string
    username: string
    first_name: string
    last_name: string
    email: string
    phone_number: string
    gender: string
    birthday: string
    graduate_level: string
    graduation_year: number
    major: string
    shirt_size: string
    jacket_size: string
    sae_registration_number: string
    avatar_url: string
    initial_role: string
    groups: string[]
    created_at: string
    updated_at: string
  }
  service_account?: {
    id: string
    entity_id: string
    application_id: string
    name: string
    created_by: string
    created_at: string
  }
}

export function useAuth() {
  const session = loadSession()
  const qc = useQueryClient()
  const entityId = session?.entityId
  const query = useQuery({
    queryKey: ["currentEntity", entityId],
    queryFn: async () => {
      const res = await api.get<Entity>("/entities/@me")
      return res.data
    },
    enabled: !!entityId,
    staleTime: 5 * 60 * 1000,
  })

  function logout() {
    clearSession()
    qc.clear()
    window.location.href = "/auth/login"
  }

  return {
    session,
    user: query.data,
    isLoading: query.isLoading,
    isAuthenticated: !!session,
    refresh: () => qc.invalidateQueries({ queryKey: ["currentEntity"] }),
    logout,
  }
}
