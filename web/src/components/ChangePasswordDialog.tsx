import { useEffect, useRef, useState, type FormEvent } from "react"

import { CODE_LENGTH, CodeInput } from "@/components/CodeInput"
import { SuccessCheck } from "@/components/SuccessCheck"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { api } from "@/lib/api"
import { validatePassword } from "@/lib/password"

const DONE_MS = 1600

type Step = "password" | "code" | "done"
type ErrorBody = { error?: string; code?: string; retry_after_seconds?: number }
type SendResponse = { email: string; resend_after: string }

function errorBody(err: unknown): ErrorBody | undefined {
  return (err as { response?: { data?: ErrorBody } })?.response?.data
}

export function ChangePasswordDialog({
  open,
  onOpenChange,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
}) {
  const [step, setStep] = useState<Step>("password")
  const [password, setPassword] = useState("")
  const [confirmPassword, setConfirmPassword] = useState("")
  const [code, setCode] = useState("")
  const [sentTo, setSentTo] = useState("")
  const [resendAt, setResendAt] = useState<number | null>(null)
  const [now, setNow] = useState(() => Date.now())
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const codeRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    if (resendAt === null) return
    const timer = setInterval(() => setNow(Date.now()), 1000)
    return () => clearInterval(timer)
  }, [resendAt])

  function close() {
    setStep("password")
    setPassword("")
    setConfirmPassword("")
    setCode("")
    setResendAt(null)
    setError(null)
    onOpenChange(false)
  }

  async function sendCode() {
    setBusy(true)
    setError(null)
    try {
      const response = await api.post<SendResponse>("/users/@me/password/code", { new_password: password })
      setSentTo(response.data.email)
      setResendAt(new Date(response.data.resend_after).getTime())
      setNow(Date.now())
      setCode("")
      setStep("code")
    } catch (err) {
      const body = errorBody(err)
      if (body?.retry_after_seconds) {
        setResendAt(Date.now() + body.retry_after_seconds * 1000)
        setNow(Date.now())
        setError("Please wait before requesting another code.")
      } else {
        setError(body?.error ?? "Couldn't send the code. Try again.")
        if (body?.code === "invalid_password") setStep("password")
      }
    } finally {
      setBusy(false)
    }
  }

  function handlePassword(event: FormEvent) {
    event.preventDefault()
    if (busy) return
    const problem = validatePassword(password)
    if (problem) {
      setError(problem)
      return
    }
    if (password !== confirmPassword) {
      setError("The passwords don't match.")
      return
    }
    void sendCode()
  }

  async function submitCode(value: string) {
    if (busy || value.length !== CODE_LENGTH) return
    setBusy(true)
    setError(null)
    try {
      await api.put("/users/@me/password", { new_password: password, code: value })
      setStep("done")
      setTimeout(close, DONE_MS)
    } catch (err) {
      const body = errorBody(err)
      setCode("")
      if (body?.code === "wrong_code") {
        setError("That code isn't right. Check the email and try again.")
      } else if (body?.code === "no_active_code") {
        setError("That code has expired. Send a new one.")
      } else if (body?.code === "too_many_attempts") {
        setError("Too many incorrect attempts. Send a new code.")
      } else {
        setError(body?.error ?? "Couldn't change your password. Try again.")
      }
      codeRef.current?.focus()
    } finally {
      setBusy(false)
    }
  }

  function handleCodeChange(value: string) {
    const digits = value.replace(/\D/g, "").slice(0, CODE_LENGTH)
    setCode(digits)
    if (digits) setError(null)
    if (digits.length === CODE_LENGTH) void submitCode(digits)
  }

  const resendIn = resendAt === null ? 0 : Math.max(0, Math.ceil((resendAt - now) / 1000))

  return (
    <Dialog open={open} onOpenChange={(next) => (next ? onOpenChange(true) : close())}>
      <DialogContent className="overflow-hidden">
        {step === "password" && (
          <form onSubmit={handlePassword} className="space-y-4">
            <DialogHeader>
              <DialogTitle>Change password</DialogTitle>
              <DialogDescription>
                At least 8 characters with a number and a capital letter. We'll email you a code to confirm the change.
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-2">
              <Label htmlFor="new-password">New password</Label>
              <Input
                id="new-password"
                type="password"
                autoComplete="new-password"
                value={password}
                onChange={(event) => setPassword(event.target.value)}
                required
                autoFocus
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="confirm-password">Confirm new password</Label>
              <Input
                id="confirm-password"
                type="password"
                autoComplete="new-password"
                value={confirmPassword}
                onChange={(event) => setConfirmPassword(event.target.value)}
                aria-invalid={!!error}
                required
              />
              {error && <p className="text-sm text-destructive">{error}</p>}
            </div>
            <div className="flex justify-end gap-2">
              <Button type="button" variant="outline" disabled={busy} onClick={close}>
                Cancel
              </Button>
              <Button type="submit" disabled={busy || !password || !confirmPassword}>
                {busy ? "Sending code…" : "Continue"}
              </Button>
            </div>
          </form>
        )}

        {step === "code" && (
          <div className="space-y-5 animate-in fade-in-0 slide-in-from-right-8 duration-300">
            <DialogHeader className="items-center text-center">
              <DialogTitle>Check your email</DialogTitle>
              <DialogDescription>
                We sent a 6-digit code to <span className="font-medium text-foreground">{sentTo}</span>. Enter it to
                change your password.
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-2 px-2">
              <CodeInput inputRef={codeRef} value={code} onChange={handleCodeChange} disabled={busy} invalid={!!error} />
              {error && <p className="text-center text-sm text-destructive">{error}</p>}
            </div>
            <div className="flex items-center justify-between gap-2">
              <button
                type="button"
                onClick={() => void sendCode()}
                disabled={busy || resendIn > 0}
                className="text-sm text-muted-foreground transition-colors hover:text-gr-pink disabled:pointer-events-none"
              >
                {resendIn > 0 ? `Resend in ${resendIn}s` : "Resend code"}
              </button>
              <Button
                type="button"
                variant="outline"
                disabled={busy}
                onClick={() => {
                  setError(null)
                  setStep("password")
                }}
              >
                Back
              </Button>
            </div>
          </div>
        )}

        {step === "done" && (
          <div className="flex flex-col items-center gap-4 py-6 text-center">
            <SuccessCheck className="size-20" />
            <div className="space-y-1 animate-in fade-in-0 slide-in-from-bottom-2 delay-500 duration-300 fill-mode-both">
              <DialogTitle>Password updated</DialogTitle>
              <DialogDescription>Use your new password the next time you sign in.</DialogDescription>
            </div>
          </div>
        )}
      </DialogContent>
    </Dialog>
  )
}
