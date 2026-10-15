import { LockKeyhole } from "lucide-react"
import { useState, type FormEvent } from "react"

import { SuccessCheck } from "@/components/SuccessCheck"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { api } from "@/lib/api"
import { cn } from "@/lib/utils"

const UNLOCK_MS = 900

type Step = "password" | "unlocking" | "email"
type ErrorBody = { error?: string; code?: string }

function errorBody(err: unknown): ErrorBody | undefined {
  return (err as { response?: { data?: ErrorBody } })?.response?.data
}

export function ChangeEmailDialog({
  open,
  currentEmail,
  onOpenChange,
  onChanged,
}: {
  open: boolean
  currentEmail: string
  onOpenChange: (open: boolean) => void
  onChanged: (email: string) => void | Promise<void>
}) {
  const [step, setStep] = useState<Step>("password")
  const [password, setPassword] = useState("")
  const [reauthToken, setReauthToken] = useState<string | null>(null)
  const [email, setEmail] = useState("")
  const [confirmEmail, setConfirmEmail] = useState("")
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)

  function close() {
    setStep("password")
    setPassword("")
    setReauthToken(null)
    setEmail("")
    setConfirmEmail("")
    setError(null)
    onOpenChange(false)
  }

  async function handlePassword(event: FormEvent) {
    event.preventDefault()
    if (busy) return
    setBusy(true)
    setError(null)
    try {
      const response = await api.post<{ reauth_token: string }>("/users/@me/reauthenticate", { password })
      setReauthToken(response.data.reauth_token)
      setPassword("")
      setStep("unlocking")
      setTimeout(() => setStep("email"), UNLOCK_MS)
    } catch (err) {
      setPassword("")
      setError(errorBody(err)?.error ?? "Couldn't confirm your password. Try again.")
    } finally {
      setBusy(false)
    }
  }

  async function handleEmail(event: FormEvent) {
    event.preventDefault()
    if (busy) return
    const newEmail = email.trim()
    if (newEmail.toLowerCase() !== confirmEmail.trim().toLowerCase()) {
      setError("The email addresses don't match.")
      return
    }
    if (newEmail.toLowerCase() === currentEmail.toLowerCase()) {
      setError("That's already your email address.")
      return
    }
    setBusy(true)
    setError(null)
    try {
      const response = await api.put<{ email: string }>(
        "/users/@me/email",
        { email: newEmail },
        { headers: { "X-Reauth-Token": reauthToken ?? "" } },
      )
      await onChanged(response.data.email)
      close()
    } catch (err) {
      const body = errorBody(err)
      if (body?.code === "reauth_required") {
        setReauthToken(null)
        setStep("password")
        setError("Your confirmation expired. Enter your password again.")
      } else {
        setError(body?.error ?? "Couldn't change your email. Try again.")
      }
    } finally {
      setBusy(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={(next) => (next ? onOpenChange(true) : close())}>
      <DialogContent className="overflow-hidden">
        {step === "email" ? (
          <form
            key="email"
            onSubmit={handleEmail}
            className="space-y-4 animate-in fade-in-0 slide-in-from-right-8 duration-300"
          >
            <DialogHeader>
              <DialogTitle>Change email</DialogTitle>
              <DialogDescription>
                Your email is currently <span className="font-medium text-foreground">{currentEmail}</span>. We'll send a
                code to the new address to verify it.
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-2">
              <Label htmlFor="change-email-new">New email</Label>
              <Input
                id="change-email-new"
                type="email"
                autoComplete="email"
                value={email}
                onChange={(event) => setEmail(event.target.value)}
                required
                autoFocus
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="change-email-confirm">Confirm new email</Label>
              <Input
                id="change-email-confirm"
                type="email"
                autoComplete="off"
                value={confirmEmail}
                onChange={(event) => setConfirmEmail(event.target.value)}
                aria-invalid={!!error}
                required
              />
              {error && <p className="text-sm text-destructive">{error}</p>}
            </div>
            <div className="flex justify-end gap-2">
              <Button type="button" variant="outline" disabled={busy} onClick={close}>
                Cancel
              </Button>
              <Button type="submit" disabled={busy || !email.trim() || !confirmEmail.trim()}>
                {busy ? "Changing…" : "Change email"}
              </Button>
            </div>
          </form>
        ) : (
          <form key="password" onSubmit={handlePassword} className="relative">
            <div
              className={cn(
                "space-y-4 transition-all duration-300",
                step === "unlocking" && "pointer-events-none scale-95 opacity-0",
              )}
            >
              <div className="flex justify-center pt-2">
                <div className="flex size-14 items-center justify-center rounded-full border border-border bg-muted/40">
                  <LockKeyhole className="size-6 text-muted-foreground" />
                </div>
              </div>
              <DialogHeader className="items-center text-center">
                <DialogTitle>Confirm it's you</DialogTitle>
                <DialogDescription>Changing your email is a sensitive action. Enter your password to continue.</DialogDescription>
              </DialogHeader>
              <div className="space-y-2">
                <Label htmlFor="change-email-password" className="sr-only">
                  Password
                </Label>
                <Input
                  id="change-email-password"
                  type="password"
                  autoComplete="current-password"
                  placeholder="Password"
                  value={password}
                  onChange={(event) => setPassword(event.target.value)}
                  aria-invalid={!!error}
                  required
                  autoFocus
                />
                {error && <p className="text-sm text-destructive">{error}</p>}
              </div>
              <div className="flex justify-end gap-2">
                <Button type="button" variant="outline" disabled={busy} onClick={close}>
                  Cancel
                </Button>
                <Button type="submit" disabled={busy || !password}>
                  {busy ? "Checking…" : "Continue"}
                </Button>
              </div>
            </div>
            {step === "unlocking" && (
              <div aria-hidden className="absolute inset-0 flex items-center justify-center">
                <SuccessCheck className="size-20" />
              </div>
            )}
          </form>
        )}
      </DialogContent>
    </Dialog>
  )
}
