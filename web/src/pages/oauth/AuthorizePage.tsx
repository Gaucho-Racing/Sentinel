import { useQuery } from "@tanstack/react-query"
import { Loader2 } from "lucide-react"
import { useEffect, useMemo, useRef, useState } from "react"
import { Navigate, useLocation, useSearchParams } from "react-router-dom"

import { OutlineButton } from "@/components/OutlineButton"
import { SuccessCheck } from "@/components/SuccessCheck"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Button } from "@/components/ui/button"
import { api } from "@/lib/api"
import { loadSession, useAuth } from "@/lib/auth"
import { resolveScopes } from "@/lib/scopes"
import { cn } from "@/lib/utils"

const CONVERGE_MS = 250
const CHECKMARK_DRAW_MS = 650
const HOLD_MS = 250

type Action = "approve" | "deny"

type ValidateResponse = {
  client_id: string
  redirect_uri: string
  scope: string
  prompt: string
  app_name: string
  app_icon_url: string
}

function initials(name: string) {
  return name
    .split(" ")
    .map((part) => part[0])
    .filter(Boolean)
    .slice(0, 2)
    .join("")
    .toUpperCase()
}

function buildRedirect(redirectUri: string, params: Record<string, string>) {
  try {
    const url = new URL(redirectUri)
    for (const [key, value] of Object.entries(params)) {
      url.searchParams.set(key, value)
    }
    return url.toString()
  } catch {
    return redirectUri
  }
}

function errorMessage(err: unknown): string | undefined {
  return (err as { response?: { data?: { error?: string } } })?.response?.data?.error
}

export default function AuthorizePage() {
  const [params] = useSearchParams()
  const location = useLocation()
  const session = loadSession()
  const { user } = useAuth()

  const clientId = params.get("client_id")
  const redirectUri = params.get("redirect_uri") ?? ""
  const scope = params.get("scope") ?? ""
  const state = params.get("state")
  const nonce = params.get("nonce")
  const prompt = params.get("prompt") ?? ""

  const [busy, setBusy] = useState<Action | null>(null)
  const [success, setSuccess] = useState(false)
  const autoApproved = useRef(false)

  // Echo `state` back to the client on every redirect — it's the client's
  // CSRF token and the spec requires it round-trips untouched.
  const withState = (extra: Record<string, string>) =>
    state ? { ...extra, state } : extra

  const validate = useQuery({
    queryKey: ["oauth-authorize", clientId, redirectUri, scope, prompt, session?.entityId],
    queryFn: async () => {
      const search = new URLSearchParams({
        client_id: clientId ?? "",
        redirect_uri: redirectUri,
        scope,
        entity_id: session?.entityId ?? "",
      })
      if (prompt) search.set("prompt", prompt)
      const res = await api.get<ValidateResponse>(`/oauth/authorize?${search.toString()}`)
      return res.data
    },
    enabled: !!session && !!clientId,
    retry: false,
  })

  const resolvedScope = validate.data?.scope ?? scope
  const scopes = useMemo(() => resolveScopes(resolvedScope), [resolvedScope])

  async function complete(action: Action) {
    if (busy) return
    setBusy(action)

    if (action === "deny") {
      window.location.href = buildRedirect(redirectUri, withState({ error: "access_denied" }))
      return
    }

    try {
      const search = new URLSearchParams({
        client_id: clientId ?? "",
        redirect_uri: redirectUri,
        scope: resolvedScope,
      })
      // Bind the OIDC nonce to the authorization code so the backend can echo
      // it into the issued ID token.
      if (nonce) search.set("nonce", nonce)
      const res = await api.post<{ code: string; redirect_uri: string }>(
        `/oauth/authorize?${search.toString()}`,
        { entity_id: session?.entityId },
      )
      setSuccess(true)
      await new Promise((resolve) =>
        setTimeout(resolve, CONVERGE_MS + CHECKMARK_DRAW_MS + HOLD_MS),
      )
      window.location.href = buildRedirect(res.data.redirect_uri, withState({ code: res.data.code }))
    } catch (err) {
      // Surface the OAuth error to the client app per the spec rather than
      // stranding the user on a dead consent screen.
      window.location.href = buildRedirect(
        redirectUri,
        withState({ error: errorMessage(err) ?? "server_error" }),
      )
    }
  }

  // prompt=none means the user recently consented — approve silently without
  // flashing the consent screen. The backend resolves the effective prompt.
  useEffect(() => {
    if (validate.data?.prompt === "none" && !autoApproved.current) {
      autoApproved.current = true
      void complete("approve")
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [validate.data?.prompt])

  if (!session) {
    return <Navigate to="/auth/login" state={{ from: location }} replace />
  }

  if (!clientId || !redirectUri || !scope) {
    return (
      <main className="flex min-h-svh items-center justify-center px-4 py-12">
        <div className="w-full max-w-sm space-y-2 text-center">
          <h1 className="text-xl font-semibold tracking-tight">Invalid request</h1>
          <p className="text-sm text-muted-foreground">
            This authorization link is missing required parameters.
          </p>
        </div>
      </main>
    )
  }

  if (validate.isLoading || validate.data?.prompt === "none") {
    return (
      <main className="flex min-h-svh items-center justify-center px-4 py-12">
        <Loader2 className="size-6 animate-spin text-muted-foreground" />
      </main>
    )
  }

  if (validate.isError || !validate.data) {
    return (
      <main className="flex min-h-svh items-center justify-center px-4 py-12">
        <div className="w-full max-w-sm space-y-2 text-center">
          <h1 className="text-xl font-semibold tracking-tight">Can't authorize</h1>
          <p className="text-sm text-muted-foreground">
            {errorMessage(validate.error) ?? "This authorization request couldn't be validated."}
          </p>
        </div>
      </main>
    )
  }

  const app = validate.data
  const userName = user?.user
    ? `${user.user.first_name} ${user.user.last_name}`.trim()
    : ""
  const userEmail = user?.user?.email ?? ""
  const userAvatar = user?.user?.avatar_url

  return (
    <main className="relative flex min-h-svh items-center justify-center px-4 py-12">
      <div
        className={cn(
          "w-full max-w-md space-y-8 transition-all ease-in",
          success
            ? "scale-0 opacity-0 duration-[250ms]"
            : "scale-100 opacity-100 duration-200",
        )}
      >
        <div className="flex flex-col items-center gap-3 text-center">
          {app.app_icon_url ? (
            <Avatar className="size-14 rounded-xl">
              <AvatarImage src={app.app_icon_url} alt={app.app_name} />
              <AvatarFallback className="rounded-xl bg-gradient-to-br from-gr-pink to-gr-purple text-xl font-semibold text-white">
                {app.app_name.slice(0, 1).toUpperCase()}
              </AvatarFallback>
            </Avatar>
          ) : (
            <div className="flex size-14 items-center justify-center rounded-xl bg-gradient-to-br from-gr-pink to-gr-purple text-xl font-semibold text-white">
              {app.app_name.slice(0, 1).toUpperCase()}
            </div>
          )}
          <div>
            <h1 className="text-xl font-semibold tracking-tight">{app.app_name}</h1>
            <p className="mt-1 text-sm text-muted-foreground">
              wants to access your Sentinel account
            </p>
          </div>
        </div>

        <div className="space-y-2">
          <p className="text-xs uppercase tracking-wider text-muted-foreground">Signed in as</p>
          <div className="flex items-center gap-3">
            <Avatar className="size-9">
              <AvatarImage src={userAvatar} alt={userName} />
              <AvatarFallback>{userName ? initials(userName) : "?"}</AvatarFallback>
            </Avatar>
            <div>
              <p className="text-sm font-medium leading-none">{userName || "Your account"}</p>
              <p className="mt-1 text-xs text-muted-foreground">{userEmail}</p>
            </div>
          </div>
        </div>

        <div className="space-y-3">
          <p className="text-xs uppercase tracking-wider text-muted-foreground">
            This will let {app.app_name}:
          </p>
          <ul className="space-y-3">
            {scopes.map((scope) => (
              <li key={scope.key} className="flex items-start gap-3">
                <div className="mt-0.5 flex size-7 shrink-0 items-center justify-center rounded-md bg-muted/60 text-muted-foreground">
                  <scope.icon className="size-3.5" />
                </div>
                <div className="min-w-0 flex-1">
                  <p className="text-sm leading-tight">{scope.label}</p>
                  <p className="mt-1 text-xs text-muted-foreground">{scope.description}</p>
                  {!scope.known && (
                    <p className="mt-1 font-mono text-[11px] text-muted-foreground">{scope.key}</p>
                  )}
                </div>
              </li>
            ))}
          </ul>
        </div>

        <div className="space-y-3">
          <p className="text-xs text-muted-foreground">
            You'll be redirected to{" "}
            <span className="break-all font-mono text-foreground">{app.redirect_uri || "—"}</span>
          </p>

          <div className="flex gap-2">
            <Button
              type="button"
              variant="outline"
              className="h-10 flex-1 rounded-xl"
              disabled={busy !== null}
              onClick={() => complete("deny")}
            >
              {busy === "deny" ? <Loader2 className="size-4 animate-spin" /> : "Cancel"}
            </Button>
            <OutlineButton
              type="button"
              className="flex-1"
              loading={busy === "approve"}
              disabled={busy !== null}
              onClick={() => complete("approve")}
            >
              Authorize
            </OutlineButton>
          </div>
        </div>
      </div>

      {success && (
        <div
          aria-hidden
          className="pointer-events-none absolute inset-0 flex items-center justify-center"
        >
          <SuccessCheck className="size-20" />
        </div>
      )}
    </main>
  )
}
