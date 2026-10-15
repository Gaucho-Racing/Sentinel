import { Loader2 } from "lucide-react"
import { useEffect, useRef, useState, type FormEvent } from "react"
import { Navigate, useNavigate } from "react-router-dom"
import { toast } from "sonner"

import { OutlineButton } from "@/components/OutlineButton"
import { SuccessCheck } from "@/components/SuccessCheck"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { api } from "@/lib/api"
import { consumeLoginReturnLocation, useAuth } from "@/lib/auth"
import { cn } from "@/lib/utils"

const CODE_LENGTH = 6
const CONVERGE_MS = 250
const CHECKMARK_DRAW_MS = 650
const HOLD_MS = 250

type SendResponse = { email: string; expires_at: string; resend_after: string }
type ErrorBody = { error?: string; code?: string; retry_after_seconds?: number }

function errorResponse(err: unknown) {
  return (err as { response?: { status?: number; data?: ErrorBody } })?.response
}

export default function VerifyEmailPage() {
  const navigate = useNavigate()
  const { user, isLoading, refresh, logout } = useAuth()
  const [code, setCode] = useState("")
  const [sending, setSending] = useState(false)
  const [verifying, setVerifying] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [resendAt, setResendAt] = useState<number | null>(null)
  const [now, setNow] = useState(() => Date.now())
  const [transitioning, setTransitioning] = useState(false)
  const autoSent = useRef(false)
  const inputRef = useRef<HTMLInputElement>(null)

  const email = user?.email_auth?.email
  const verified = !!user?.email_auth?.verified_at

  useEffect(() => {
    if (resendAt === null) return
    const timer = setInterval(() => setNow(Date.now()), 1000)
    return () => clearInterval(timer)
  }, [resendAt])

  async function sendCode(isAuto = false) {
    setSending(true)
    setError(null)
    try {
      const res = await api.post<SendResponse>("/users/@me/email/verification")
      setResendAt(new Date(res.data.resend_after).getTime())
      setNow(Date.now())
      if (!isAuto) toast.success(`Sent a new code to ${res.data.email}`)
      inputRef.current?.focus()
    } catch (err) {
      const res = errorResponse(err)
      if (res?.status === 429 && res.data?.retry_after_seconds) {
        setResendAt(Date.now() + res.data.retry_after_seconds * 1000)
        setNow(Date.now())
        if (!isAuto) setError(`Please wait before requesting another code.`)
      } else if (res?.status === 409) {
        await refresh()
      } else {
        setError(res?.data?.error ?? "Couldn't send the verification email. Try again.")
      }
    } finally {
      setSending(false)
    }
  }

  useEffect(() => {
    if (!user || verified || autoSent.current) return
    autoSent.current = true
    void sendCode(true)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [user, verified])

  async function verify(value: string) {
    if (verifying || value.length !== CODE_LENGTH) return
    setVerifying(true)
    setError(null)
    try {
      await api.post("/users/@me/email/verification/confirm", { code: value })
      setTransitioning(true)
      await new Promise((resolve) => setTimeout(resolve, CONVERGE_MS + CHECKMARK_DRAW_MS + HOLD_MS))
      await refresh()
      navigate(consumeLoginReturnLocation(), { replace: true })
    } catch (err) {
      const res = errorResponse(err)
      setCode("")
      switch (res?.data?.code) {
        case "wrong_code":
          setError("That code isn't right. Check the email and try again.")
          break
        case "no_active_code":
          setError("That code has expired. Send a new one.")
          break
        case "too_many_attempts":
          setError("Too many incorrect attempts. Send a new code.")
          break
        default:
          setError(res?.data?.error ?? "Couldn't verify the code. Try again.")
      }
      inputRef.current?.focus()
    } finally {
      setVerifying(false)
    }
  }

  function handleChange(value: string) {
    const digits = value.replace(/\D/g, "").slice(0, CODE_LENGTH)
    setCode(digits)
    if (digits.length === CODE_LENGTH) void verify(digits)
  }

  function handleSubmit(event: FormEvent) {
    event.preventDefault()
    void verify(code)
  }

  if (isLoading || !user) {
    return (
      <main className="flex min-h-svh items-center justify-center px-4 py-12">
        <Loader2 className="size-6 animate-spin text-muted-foreground" />
      </main>
    )
  }

  if (user.type !== "USER" || (verified && !transitioning)) {
    return <Navigate to={consumeLoginReturnLocation()} replace />
  }

  const resendIn = resendAt === null ? 0 : Math.max(0, Math.ceil((resendAt - now) / 1000))
  const busy = verifying || transitioning

  return (
    <main className="relative flex min-h-svh items-center justify-center px-4 py-12">
      <div
        className={cn(
          "w-full max-w-sm space-y-8 transition-all ease-in",
          transitioning ? "scale-0 opacity-0 duration-[250ms]" : "scale-100 opacity-100 duration-200",
        )}
      >
        <div className="flex flex-col items-center gap-3 text-center">
          <img src="/logo/gr-logo-blank.png" alt="Gaucho Racing" className="size-12" />
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">Verify your email</h1>
            <p className="mt-1 text-sm text-muted-foreground">
              {email ? (
                <>
                  We sent a 6-digit code to <span className="font-medium text-foreground">{email}</span>.
                </>
              ) : (
                "Your account has no email address. Ask a Sentinel admin for help."
              )}
            </p>
          </div>
        </div>

        {email && (
          <form onSubmit={handleSubmit} noValidate className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="code" className="sr-only">
                Verification code
              </Label>
              <Input
                ref={inputRef}
                id="code"
                value={code}
                onChange={(event) => handleChange(event.target.value)}
                inputMode="numeric"
                autoComplete="one-time-code"
                pattern="[0-9]*"
                maxLength={CODE_LENGTH}
                placeholder="000000"
                disabled={busy}
                autoFocus
                aria-invalid={!!error}
                className="h-14 pl-[calc(0.75rem+0.5em)] text-center font-mono text-2xl tracking-[0.5em] placeholder:text-muted-foreground/40"
              />
              {error && <p className="text-center text-sm text-destructive">{error}</p>}
            </div>

            <OutlineButton type="submit" loading={verifying} disabled={busy || code.length !== CODE_LENGTH}>
              Verify
            </OutlineButton>

            <p className="text-center text-sm text-muted-foreground">
              Didn't get it?{" "}
              <button
                type="button"
                onClick={() => void sendCode()}
                disabled={sending || resendIn > 0 || busy}
                className="text-foreground transition-colors hover:text-gr-pink disabled:pointer-events-none disabled:text-muted-foreground"
              >
                {sending ? "Sending…" : resendIn > 0 ? `Resend in ${resendIn}s` : "Resend code"}
              </button>
            </p>
          </form>
        )}

        <p className="text-center text-xs text-muted-foreground">
          Not you?{" "}
          <button type="button" onClick={logout} className="text-foreground transition-colors hover:text-gr-pink">
            Sign out
          </button>
        </p>
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
