import { useEffect, useRef, useState, type FormEvent } from "react"

import { CODE_LENGTH, CodeInput } from "@/components/CodeInput"
import { PhoneNumberInput } from "@/components/PhoneNumberInput"
import { ReauthStep } from "@/components/ReauthStep"
import { SuccessCheck } from "@/components/SuccessCheck"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Label } from "@/components/ui/label"
import { api } from "@/lib/api"
import { formatPhoneNumber } from "@/lib/phone"

const DONE_MS = 1600

export type PhoneDialogMode = "verify" | "change"
type Step = "reauth" | "number" | "code" | "done"
type ErrorBody = { error?: string; code?: string; retry_after_seconds?: number }
type SendResponse = { phone_number: string; resend_after: string }

function errorBody(err: unknown): ErrorBody | undefined {
  return (err as { response?: { data?: ErrorBody } })?.response?.data
}

export function PhoneDialog({
  mode,
  currentPhone,
  onOpenChange,
  onUpdated,
}: {
  mode: PhoneDialogMode | null
  currentPhone: string
  onOpenChange: (open: boolean) => void
  onUpdated: () => void | Promise<void>
}) {
  const [step, setStep] = useState<Step>(mode === "verify" ? "code" : "reauth")
  const [reauthToken, setReauthToken] = useState("")
  const [reauthError, setReauthError] = useState<string | null>(null)
  const [phone, setPhone] = useState("")
  const [code, setCode] = useState("")
  const [sentTo, setSentTo] = useState("")
  const [resendAt, setResendAt] = useState<number | null>(null)
  const [now, setNow] = useState(() => Date.now())
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const codeRef = useRef<HTMLInputElement>(null)
  const autoSent = useRef(false)

  useEffect(() => {
    if (resendAt === null) return
    const timer = setInterval(() => setNow(Date.now()), 1000)
    return () => clearInterval(timer)
  }, [resendAt])

  async function sendCode() {
    setBusy(true)
    setError(null)
    try {
      const response = await api.post<SendResponse>("/users/@me/phone/verification")
      setSentTo(response.data.phone_number)
      setResendAt(new Date(response.data.resend_after).getTime())
      setNow(Date.now())
      setCode("")
      setStep("code")
    } catch (err) {
      const body = errorBody(err)
      if (body?.retry_after_seconds) {
        setSentTo((current) => current || currentPhone)
        setResendAt(Date.now() + body.retry_after_seconds * 1000)
        setNow(Date.now())
      } else {
        setError(body?.error ?? "Couldn't send the code. Try again.")
      }
    } finally {
      setBusy(false)
    }
  }

  useEffect(() => {
    if (mode === "verify" && !autoSent.current) {
      autoSent.current = true
      void sendCode()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [mode])

  function close() {
    onOpenChange(false)
  }

  async function handleNumber(event: FormEvent) {
    event.preventDefault()
    if (busy) return
    setBusy(true)
    setError(null)
    try {
      await api.put("/users/@me/phone", { phone_number: phone }, { headers: { "X-Reauth-Token": reauthToken } })
      await onUpdated()
      setBusy(false)
      await sendCode()
    } catch (err) {
      const body = errorBody(err)
      if (body?.code === "reauth_required") {
        setReauthError("Your confirmation expired. Enter your password again.")
        setReauthToken("")
        setStep("reauth")
      } else {
        setError(body?.error ?? "Couldn't update your phone number. Try again.")
      }
      setBusy(false)
    }
  }

  async function submitCode(value: string) {
    if (busy || value.length !== CODE_LENGTH) return
    setBusy(true)
    setError(null)
    try {
      await api.post("/users/@me/phone/verification/confirm", { code: value })
      setStep("done")
      await onUpdated()
      setTimeout(close, DONE_MS)
    } catch (err) {
      const body = errorBody(err)
      setCode("")
      if (body?.code === "wrong_code") {
        setError("That code isn't right. Check your texts and try again.")
      } else if (body?.code === "no_active_code") {
        setError("That code has expired. Send a new one.")
      } else if (body?.code === "too_many_attempts") {
        setError("Too many incorrect attempts. Send a new code.")
      } else {
        setError(body?.error ?? "Couldn't verify the code. Try again.")
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
    <Dialog open={mode !== null} onOpenChange={(next) => (next ? onOpenChange(true) : close())}>
      <DialogContent className="overflow-hidden">
        {step === "reauth" && (
          <ReauthStep
            key={reauthError ?? "reauth"}
            description="Changing your phone number is a sensitive action. Enter your password to continue."
            initialError={reauthError}
            onVerified={(token) => {
              setReauthToken(token)
              setReauthError(null)
              setStep("number")
            }}
            onCancel={close}
          />
        )}

        {step === "number" && (
          <form onSubmit={handleNumber} className="space-y-4 animate-in fade-in-0 slide-in-from-right-8 duration-300">
            <DialogHeader>
              <DialogTitle>{currentPhone ? "Change phone number" : "Add phone number"}</DialogTitle>
              <DialogDescription>
                {currentPhone ? (
                  <>
                    Your number is currently{" "}
                    <span className="font-medium text-foreground">{formatPhoneNumber(currentPhone)}</span>.{" "}
                  </>
                ) : null}
                We'll text a code to the new number to verify it.
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-1.5">
              <Label htmlFor="new-phone">New phone number</Label>
              <PhoneNumberInput id="new-phone" value={phone} onChange={setPhone} required autoFocus />
              {error && <p className="text-sm text-destructive">{error}</p>}
            </div>
            <div className="flex justify-end gap-2">
              <Button type="button" variant="outline" disabled={busy} onClick={close}>
                Cancel
              </Button>
              <Button type="submit" disabled={busy || phone.replace(/\D/g, "").length < 8}>
                {busy ? "Saving…" : "Continue"}
              </Button>
            </div>
          </form>
        )}

        {step === "code" && (
          <div className="space-y-5 animate-in fade-in-0 slide-in-from-right-8 duration-300">
            <DialogHeader className="items-center text-center">
              <DialogTitle>Check your texts</DialogTitle>
              <DialogDescription>
                {sentTo ? (
                  <>
                    We sent a 6-digit code to{" "}
                    <span className="font-medium text-foreground">{formatPhoneNumber(sentTo)}</span>.
                  </>
                ) : (
                  "Sending a code to your phone…"
                )}
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
              <Button type="button" variant="outline" disabled={busy} onClick={close}>
                Cancel
              </Button>
            </div>
          </div>
        )}

        {step === "done" && (
          <div className="flex flex-col items-center gap-4 py-6 text-center">
            <SuccessCheck className="size-20" />
            <div className="space-y-1 animate-in fade-in-0 slide-in-from-bottom-2 delay-500 duration-300 fill-mode-both">
              <DialogTitle>Phone number verified</DialogTitle>
              <DialogDescription>{formatPhoneNumber(sentTo)} is now verified on your account.</DialogDescription>
            </div>
          </div>
        )}
      </DialogContent>
    </Dialog>
  )
}
