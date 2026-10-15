import { useState, type FormEvent } from "react"

import { ReauthStep } from "@/components/ReauthStep"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { api } from "@/lib/api"

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
  const [reauthToken, setReauthToken] = useState<string | null>(null)
  const [reauthError, setReauthError] = useState<string | null>(null)
  const [email, setEmail] = useState("")
  const [confirmEmail, setConfirmEmail] = useState("")
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)

  function close() {
    setReauthToken(null)
    setReauthError(null)
    setEmail("")
    setConfirmEmail("")
    setError(null)
    onOpenChange(false)
  }

  async function handleSubmit(event: FormEvent) {
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
      const body = (err as { response?: { data?: { error?: string; code?: string } } })?.response?.data
      if (body?.code === "reauth_required") {
        setReauthToken(null)
        setReauthError("Your confirmation expired. Enter your password again.")
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
        {!reauthToken ? (
          <ReauthStep
            key={reauthError ?? "reauth"}
            description="Changing your email is a sensitive action. Enter your password to continue."
            initialError={reauthError}
            onVerified={setReauthToken}
            onCancel={close}
          />
        ) : (
          <form onSubmit={handleSubmit} className="space-y-4 animate-in fade-in-0 slide-in-from-right-8 duration-300">
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
        )}
      </DialogContent>
    </Dialog>
  )
}
