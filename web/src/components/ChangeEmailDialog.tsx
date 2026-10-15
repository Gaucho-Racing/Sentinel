import { useState, type FormEvent } from "react"

import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { api } from "@/lib/api"

type Step = "email" | "password"

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
  const [step, setStep] = useState<Step>("email")
  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const newEmail = email.trim()

  function close() {
    setStep("email")
    setEmail("")
    setPassword("")
    setError(null)
    onOpenChange(false)
  }

  function handleContinue(event: FormEvent) {
    event.preventDefault()
    if (newEmail.toLowerCase() === currentEmail.toLowerCase()) {
      setError("That's already your email address.")
      return
    }
    setError(null)
    setStep("password")
  }

  async function handleConfirm(event: FormEvent) {
    event.preventDefault()
    if (saving) return
    setSaving(true)
    setError(null)
    try {
      const response = await api.put<{ email: string }>("/users/@me/email", { email: newEmail, password })
      await onChanged(response.data.email)
      close()
    } catch (err) {
      const data = (err as { response?: { data?: { error?: string; code?: string } } })?.response?.data
      setError(data?.error ?? "Couldn't change your email. Try again.")
      if (data?.code === "incorrect_password") {
        setPassword("")
      } else {
        setStep("email")
      }
    } finally {
      setSaving(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={(next) => (next ? onOpenChange(true) : close())}>
      <DialogContent>
        {step === "email" ? (
          <form onSubmit={handleContinue} className="space-y-4">
            <DialogHeader>
              <DialogTitle>Change email</DialogTitle>
              <DialogDescription>
                Your email is currently <span className="font-medium text-foreground">{currentEmail}</span>.
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
                aria-invalid={!!error}
                required
                autoFocus
              />
              {error && <p className="text-sm text-destructive">{error}</p>}
            </div>
            <div className="flex justify-end gap-2">
              <Button type="button" variant="outline" onClick={close}>
                Cancel
              </Button>
              <Button type="submit" disabled={!newEmail}>
                Continue
              </Button>
            </div>
          </form>
        ) : (
          <form onSubmit={handleConfirm} className="space-y-4">
            <DialogHeader>
              <DialogTitle>Confirm it's you</DialogTitle>
              <DialogDescription>
                Enter your password to change your email to{" "}
                <span className="font-medium text-foreground">{newEmail}</span>. You'll verify it with a code next.
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-2">
              <Label htmlFor="change-email-password">Password</Label>
              <Input
                id="change-email-password"
                type="password"
                autoComplete="current-password"
                value={password}
                onChange={(event) => setPassword(event.target.value)}
                aria-invalid={!!error}
                required
                autoFocus
              />
              {error && <p className="text-sm text-destructive">{error}</p>}
            </div>
            <div className="flex justify-end gap-2">
              <Button
                type="button"
                variant="outline"
                disabled={saving}
                onClick={() => {
                  setError(null)
                  setStep("email")
                }}
              >
                Back
              </Button>
              <Button type="submit" disabled={saving || !password}>
                {saving ? "Changing…" : "Change email"}
              </Button>
            </div>
          </form>
        )}
      </DialogContent>
    </Dialog>
  )
}
