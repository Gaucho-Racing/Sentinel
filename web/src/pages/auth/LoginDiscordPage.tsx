import { Loader2 } from "lucide-react"
import { useEffect, useRef, useState } from "react"
import { useNavigate, useSearchParams } from "react-router-dom"
import { toast } from "sonner"

import { OutlineButton } from "@/components/OutlineButton"
import { SuccessCheck } from "@/components/SuccessCheck"
import { DiscordIcon } from "@/components/icons/socials"
import { api } from "@/lib/api"
import { saveSession } from "@/lib/auth"
import { DISCORD_INVITE_URL } from "@/lib/links"
import { cn } from "@/lib/utils"

type LoginResponse = {
  access_token: string
  refresh_token: string
  expires_in: number
  entity_id: string
}

type ErrorBody = {
  error?: string
  message?: string
}

type Phase = "loading" | "no_account" | "error"

// Mirror the LoginPage transition so the user experiences the same
// converge → check → hold → navigate sequence on a successful Discord login.
const CONVERGE_MS = 250
const CHECKMARK_DRAW_MS = 650
const HOLD_MS = 250

export default function LoginDiscordPage() {
  const navigate = useNavigate()
  const [params] = useSearchParams()
  const [phase, setPhase] = useState<Phase>("loading")
  const [errorMessage, setErrorMessage] = useState<string>("")
  const [transitioning, setTransitioning] = useState(false)

  // The OAuth callback effect must run exactly once — React 18's
  // double-invoked effects in dev would otherwise burn the code (Discord
  // codes are single-use, the second exchange returns invalid_grant).
  const exchangedRef = useRef(false)

  useEffect(() => {
    if (exchangedRef.current) return
    exchangedRef.current = true

    const code = params.get("code")
    if (!code) {
      navigate("/auth/login", { replace: true })
      return
    }
    const returnTo = params.get("state") || "/"

    void (async () => {
      try {
        const res = await api.post<LoginResponse>(
          `/auth/login/discord?code=${encodeURIComponent(code)}`,
        )
        saveSession({
          accessToken: res.data.access_token,
          refreshToken: res.data.refresh_token,
          expiresIn: res.data.expires_in,
          entityId: res.data.entity_id,
        })
        setTransitioning(true)
        await new Promise((r) =>
          setTimeout(r, CONVERGE_MS + CHECKMARK_DRAW_MS + HOLD_MS),
        )
        if (document.startViewTransition) {
          document.startViewTransition(() => navigate(returnTo, { replace: true }))
        } else {
          navigate(returnTo, { replace: true })
        }
      } catch (err: unknown) {
        const body =
          (err as { response?: { data?: ErrorBody } })?.response?.data ?? {}
        if (body.error === "no_account") {
          setPhase("no_account")
          return
        }
        setErrorMessage(body.message || body.error || "Couldn't sign you in with Discord.")
        setPhase("error")
        toast.error(body.message || body.error || "Discord login failed.")
      }
    })()
  }, [navigate, params])

  return (
    <main className="relative flex min-h-svh items-center justify-center px-4 py-12">
      <div
        className={cn(
          "w-full max-w-sm space-y-8 transition-all ease-in",
          transitioning
            ? "scale-0 opacity-0 duration-[250ms]"
            : "scale-100 opacity-100 duration-200",
        )}
      >
        <div className="flex flex-col items-center gap-3 text-center">
          <img src="/logo/gr-logo-blank.png" alt="Gaucho Racing" className="size-12" />
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">
              {phase === "loading" && "Signing you in"}
              {phase === "no_account" && "No account found"}
              {phase === "error" && "Discord sign-in failed"}
            </h1>
            <p className="mt-1 text-sm text-muted-foreground">
              {phase === "loading" && "Finishing the Discord handshake…"}
              {phase === "no_account" && "We couldn't find a Sentinel account linked to this Discord user."}
              {phase === "error" && (errorMessage || "Try again from the login page.")}
            </p>
          </div>
        </div>

        {phase === "loading" && (
          <div className="flex items-center justify-center py-6">
            <Loader2 className="size-8 animate-spin text-muted-foreground" />
          </div>
        )}

        {phase === "no_account" && (
          <div className="space-y-4">
            <div className="rounded-md border border-border/60 bg-muted/30 p-4 text-sm">
              <p>
                Make sure you've joined the Gaucho Racing Discord and verified
                your account. In the <strong>#verification</strong> channel, run:
              </p>
              <pre className="mt-3 rounded bg-background px-3 py-2 font-mono text-xs">
                !verify &lt;first name&gt; &lt;last name&gt; &lt;email&gt;
              </pre>
            </div>
            <a
              href={DISCORD_INVITE_URL}
              target="_blank"
              rel="noreferrer"
              className="inline-flex w-full items-center justify-center gap-2 rounded-md bg-discord-blurple px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-discord-blurple/90"
            >
              <DiscordIcon className="size-4" />
              Join the Discord
            </a>
            <OutlineButton onClick={() => navigate("/auth/login", { replace: true })}>
              Back to sign in
            </OutlineButton>
          </div>
        )}

        {phase === "error" && (
          <OutlineButton onClick={() => navigate("/auth/login", { replace: true })}>
            Back to sign in
          </OutlineButton>
        )}
      </div>

      {transitioning && (
        <div
          aria-hidden
          className="pointer-events-none absolute inset-0 flex items-center justify-center"
          style={{ animationDelay: `${CONVERGE_MS}ms` }}
        >
          <SuccessCheck className="size-20" />
        </div>
      )}
    </main>
  )
}
