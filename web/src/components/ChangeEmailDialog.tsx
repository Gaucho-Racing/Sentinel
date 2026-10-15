import { useState, type FormEvent } from "react"

import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
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
  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  function reset() {
    setEmail("")
    setPassword("")
    setError(null)
  }

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    if (saving) return
    setSaving(true)
    setError(null)
    try {
      const response = await api.put<{ email: string }>("/users/@me/email", { email: email.trim(), password })
      await onChanged(response.data.email)
      reset()
      onOpenChange(false)
    } catch (err) {
      setError(
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ??
          "Couldn't change your email. Try again.",
      )
    } finally {
      setSaving(false)
    }
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(next) => {
        if (!next) reset()
        onOpenChange(next)
      }}
    >
      <DialogContent>
        <form onSubmit={handleSubmit} className="space-y-4">
          <DialogHeader>
            <DialogTitle>Change email</DialogTitle>
            <DialogDescription>
              We'll send a code to the new address. Until you verify it, you won't be able to sign in to other team apps,
              and you'll sign in to Sentinel with the new address.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-2">
            <Label htmlFor="change-email-current">Current email</Label>
            <Input id="change-email-current" value={currentEmail} readOnly />
          </div>
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
            <Label htmlFor="change-email-password">Password</Label>
            <Input
              id="change-email-password"
              type="password"
              autoComplete="current-password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              required
            />
          </div>
          {error && <p className="text-sm text-destructive">{error}</p>}
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={saving || !email.trim() || !password}>
              {saving ? "Changing…" : "Change email"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
