import { Loader2 } from "lucide-react"
import { useMemo, useState } from "react"
import { useSearchParams } from "react-router-dom"

import { OutlineButton } from "@/components/OutlineButton"
import { SuccessCheck } from "@/components/SuccessCheck"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Button } from "@/components/ui/button"
import { type Application, mockApplications, mockUser } from "@/lib/mock"
import { resolveScopes } from "@/lib/scopes"
import { cn } from "@/lib/utils"

const MOCK_LATENCY_MS = 1000
const CONVERGE_MS = 250
const CHECKMARK_DRAW_MS = 650
const HOLD_MS = 250

type Action = "approve" | "deny"

function initials(name: string) {
  return name
    .split(" ")
    .map((part) => part[0])
    .filter(Boolean)
    .slice(0, 2)
    .join("")
    .toUpperCase()
}

function findApp(clientId: string | null): Application | undefined {
  if (!clientId) return undefined
  return mockApplications.find((app) => app.clientId === clientId)
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

export default function AuthorizePage() {
  const [params] = useSearchParams()
  const clientId = params.get("client_id")
  const redirectUri = params.get("redirect_uri") ?? ""
  const scope = params.get("scope") ?? ""

  const app = useMemo(() => findApp(clientId), [clientId])
  const scopes = useMemo(() => resolveScopes(scope), [scope])

  const [busy, setBusy] = useState<Action | null>(null)
  const [success, setSuccess] = useState(false)

  async function complete(action: Action) {
    if (busy) return
    setBusy(action)
    await new Promise((resolve) => setTimeout(resolve, MOCK_LATENCY_MS))

    if (action === "approve") {
      setSuccess(true)
      await new Promise((resolve) =>
        setTimeout(resolve, CONVERGE_MS + CHECKMARK_DRAW_MS + HOLD_MS),
      )
      window.location.href = buildRedirect(redirectUri, { code: "mock_authorization_code" })
    } else {
      window.location.href = buildRedirect(redirectUri, { error: "access_denied" })
    }
  }

  if (!app) {
    return (
      <main className="flex min-h-svh items-center justify-center px-4 py-12">
        <div className="w-full max-w-sm space-y-2 text-center">
          <h1 className="text-xl font-semibold tracking-tight">Unknown application</h1>
          <p className="text-sm text-muted-foreground">
            No registered application matches{" "}
            <code className="font-mono text-xs">{clientId ?? "(no client_id)"}</code>.
          </p>
        </div>
      </main>
    )
  }

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
          <div className="flex size-14 items-center justify-center rounded-xl bg-gradient-to-br from-gr-pink to-gr-purple text-xl font-semibold text-white">
            {app.name.slice(0, 1).toUpperCase()}
          </div>
          <div>
            <h1 className="text-xl font-semibold tracking-tight">{app.name}</h1>
            <p className="mt-1 text-sm text-muted-foreground">
              wants to access your Sentinel account
            </p>
          </div>
        </div>

        <div className="space-y-2">
          <p className="text-xs uppercase tracking-wider text-muted-foreground">Signed in as</p>
          <div className="flex items-center gap-3">
            <Avatar className="size-9">
              <AvatarImage src={mockUser.avatarUrl} alt={mockUser.name} />
              <AvatarFallback>{initials(mockUser.name)}</AvatarFallback>
            </Avatar>
            <div>
              <p className="text-sm font-medium leading-none">{mockUser.name}</p>
              <p className="mt-1 text-xs text-muted-foreground">{mockUser.email}</p>
            </div>
          </div>
        </div>

        <div className="space-y-3">
          <p className="text-xs uppercase tracking-wider text-muted-foreground">
            This will let {app.name}:
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
            <span className="break-all font-mono text-foreground">{redirectUri || "—"}</span>
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
