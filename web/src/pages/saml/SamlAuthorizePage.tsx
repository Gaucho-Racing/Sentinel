import { useQuery } from "@tanstack/react-query"
import { Loader2 } from "lucide-react"
import { useState } from "react"
import { Navigate, useLocation, useSearchParams } from "react-router-dom"

import { OutlineButton } from "@/components/OutlineButton"
import { SuccessCheck } from "@/components/SuccessCheck"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Button } from "@/components/ui/button"
import { api } from "@/lib/api"
import { loadSession, useAuth } from "@/lib/auth"
import { cn } from "@/lib/utils"

const CONVERGE_MS = 250
const CHECKMARK_DRAW_MS = 650
const HOLD_MS = 250

type ValidateResponse = {
  sp_entity_id: string
  app_name: string
  app_icon_url: string
}

type AuthorizeResponse = {
  acs_url: string
  saml_response: string
  relay_state: string
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

function errorMessage(err: unknown): string | undefined {
  return (err as { response?: { data?: { error?: string } } })?.response?.data?.error
}

type DeniedApp = { name: string; iconUrl: string }

function accessDeniedApp(err: unknown): DeniedApp | null {
  const res = (err as {
    response?: { status?: number; data?: { error?: string; app_name?: string; app_icon_url?: string } }
  })?.response
  if (res?.status === 403 && res.data?.error === "access_denied") {
    return { name: res.data.app_name ?? "", iconUrl: res.data.app_icon_url ?? "" }
  }
  return null
}

function AppAvatar({ name, iconUrl }: { name: string; iconUrl?: string }) {
  const letter = (name.slice(0, 1) || "?").toUpperCase()
  if (iconUrl) {
    return (
      <Avatar className="size-14 rounded-xl">
        <AvatarImage src={iconUrl} alt={name} />
        <AvatarFallback className="rounded-xl bg-gradient-to-br from-gr-pink to-gr-purple text-xl font-semibold text-white">
          {letter}
        </AvatarFallback>
      </Avatar>
    )
  }
  return (
    <div className="flex size-14 items-center justify-center rounded-xl bg-gradient-to-br from-gr-pink to-gr-purple text-xl font-semibold text-white">
      {letter}
    </div>
  )
}

// postToACS builds and submits the HTTP-POST binding form to the SP's Assertion
// Consumer Service. A real cross-origin form POST is required — fetch can't
// deliver the assertion to the SP's session-setting endpoint.
function postToACS(acsUrl: string, samlResponse: string, relayState: string) {
  const form = document.createElement("form")
  form.method = "POST"
  form.action = acsUrl
  const add = (name: string, value: string) => {
    const input = document.createElement("input")
    input.type = "hidden"
    input.name = name
    input.value = value
    form.appendChild(input)
  }
  add("SAMLResponse", samlResponse)
  if (relayState) add("RelayState", relayState)
  document.body.appendChild(form)
  form.submit()
}

export default function SamlAuthorizePage() {
  const [params] = useSearchParams()
  const location = useLocation()
  const session = loadSession()
  const { user } = useAuth()

  const ssoRequest = params.get("sso_request")

  const [busy, setBusy] = useState<"approve" | "deny" | null>(null)
  const [success, setSuccess] = useState(false)
  const [deniedApp, setDeniedApp] = useState<DeniedApp | null>(null)

  const validate = useQuery({
    queryKey: ["saml-authorize", ssoRequest, session?.entityId],
    queryFn: async () => {
      const search = new URLSearchParams({
        sso_request: ssoRequest ?? "",
        entity_id: session?.entityId ?? "",
      })
      const res = await api.get<ValidateResponse>(`/saml/authorize?${search.toString()}`)
      return res.data
    },
    enabled: !!session && !!ssoRequest,
    retry: false,
  })

  async function complete(action: "approve" | "deny") {
    if (busy) return
    setBusy(action)

    if (action === "deny") {
      // SAML has no error redirect channel without a return URL, so a denial
      // simply lands the user back on the dashboard.
      window.location.href = "/"
      return
    }

    try {
      const res = await api.post<AuthorizeResponse>("/saml/authorize", {
        sso_request: ssoRequest,
        entity_id: session?.entityId,
      })
      setSuccess(true)
      await new Promise((resolve) =>
        setTimeout(resolve, CONVERGE_MS + CHECKMARK_DRAW_MS + HOLD_MS),
      )
      postToACS(res.data.acs_url, res.data.saml_response, res.data.relay_state)
    } catch (err) {
      const denied = accessDeniedApp(err)
      if (denied !== null) {
        setBusy(null)
        setDeniedApp({
          name: denied.name || validate.data?.app_name || "",
          iconUrl: denied.iconUrl || validate.data?.app_icon_url || "",
        })
        return
      }
      setBusy(null)
    }
  }

  if (!session) {
    return <Navigate to="/auth/login" state={{ from: location }} replace />
  }

  if (!ssoRequest) {
    return (
      <main className="flex min-h-svh items-center justify-center px-4 py-12">
        <div className="w-full max-w-sm space-y-2 text-center">
          <h1 className="text-xl font-semibold tracking-tight">Invalid request</h1>
          <p className="text-sm text-muted-foreground">
            This sign-in link is missing required parameters.
          </p>
        </div>
      </main>
    )
  }

  if (validate.isLoading) {
    return (
      <main className="flex min-h-svh items-center justify-center px-4 py-12">
        <Loader2 className="size-6 animate-spin text-muted-foreground" />
      </main>
    )
  }

  const denied = deniedApp ?? (validate.isError ? accessDeniedApp(validate.error) : null)
  if (denied) {
    const name = denied.name || "this application"
    return (
      <main className="flex min-h-svh items-center justify-center px-4 py-12">
        <div className="w-full max-w-md space-y-8">
          <div className="flex flex-col items-center gap-3 text-center">
            <AppAvatar name={name} iconUrl={denied.iconUrl} />
            <div>
              <h1 className="text-xl font-semibold tracking-tight">{name}</h1>
              <p className="mt-1 text-sm text-muted-foreground">
                You don't have access to this application
              </p>
            </div>
          </div>
          <p className="text-center text-sm text-muted-foreground">
            You're not in a group that's required to use{" "}
            <span className="font-medium text-foreground">{name}</span>. If you think this is a
            mistake, reach out to an administrator.
          </p>
        </div>
      </main>
    )
  }

  if (validate.isError || !validate.data) {
    return (
      <main className="flex min-h-svh items-center justify-center px-4 py-12">
        <div className="w-full max-w-sm space-y-2 text-center">
          <h1 className="text-xl font-semibold tracking-tight">Can't sign in</h1>
          <p className="text-sm text-muted-foreground">
            {errorMessage(validate.error) ?? "This sign-in request couldn't be validated."}
          </p>
        </div>
      </main>
    )
  }

  const app = validate.data
  const userName = user?.user ? `${user.user.first_name} ${user.user.last_name}`.trim() : ""
  const userEmail = user?.user?.email ?? ""
  const userAvatar = user?.user?.avatar_url

  return (
    <main className="relative flex min-h-svh items-center justify-center px-4 py-12">
      <div
        className={cn(
          "w-full max-w-md space-y-8 transition-all ease-in",
          success ? "scale-0 opacity-0 duration-[250ms]" : "scale-100 opacity-100 duration-200",
        )}
      >
        <div className="flex flex-col items-center gap-3 text-center">
          <AppAvatar name={app.app_name} iconUrl={app.app_icon_url} />
          <div>
            <h1 className="text-xl font-semibold tracking-tight">{app.app_name}</h1>
            <p className="mt-1 text-sm text-muted-foreground">wants to sign you in</p>
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
          <p className="text-xs text-muted-foreground">
            {app.app_name} will receive your name, email, and group memberships.
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
              Continue
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
