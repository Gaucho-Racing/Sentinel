import { LockKeyhole } from "lucide-react"
import { useState, type FormEvent } from "react"

import { SuccessCheck } from "@/components/SuccessCheck"
import { Button } from "@/components/ui/button"
import { DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { api } from "@/lib/api"
import { cn } from "@/lib/utils"

const UNLOCK_MS = 900

export function ReauthStep({
  description,
  initialError,
  onVerified,
  onCancel,
}: {
  description: string
  initialError?: string | null
  onVerified: (reauthToken: string) => void
  onCancel: () => void
}) {
  const [password, setPassword] = useState("")
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(initialError ?? null)
  const [unlocking, setUnlocking] = useState(false)

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    if (busy || unlocking) return
    setBusy(true)
    setError(null)
    try {
      const response = await api.post<{ reauth_token: string }>("/users/@me/reauthenticate", { password })
      setUnlocking(true)
      setTimeout(() => onVerified(response.data.reauth_token), UNLOCK_MS)
    } catch (err) {
      setPassword("")
      setError(
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ??
          "Couldn't confirm your password. Try again.",
      )
    } finally {
      setBusy(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="relative">
      <div className={cn("space-y-4 transition-all duration-300", unlocking && "pointer-events-none scale-95 opacity-0")}>
        <div className="flex justify-center pt-2">
          <div className="flex size-14 items-center justify-center rounded-full border border-border bg-muted/40">
            <LockKeyhole className="size-6 text-muted-foreground" />
          </div>
        </div>
        <DialogHeader className="items-center text-center">
          <DialogTitle>Confirm it's you</DialogTitle>
          <DialogDescription>{description}</DialogDescription>
        </DialogHeader>
        <div className="space-y-2">
          <Label htmlFor="reauth-password" className="sr-only">
            Password
          </Label>
          <Input
            id="reauth-password"
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
          <Button type="button" variant="outline" disabled={busy} onClick={onCancel}>
            Cancel
          </Button>
          <Button type="submit" disabled={busy || !password}>
            {busy ? "Checking…" : "Continue"}
          </Button>
        </div>
      </div>
      {unlocking && (
        <div aria-hidden className="absolute inset-0 flex items-center justify-center">
          <SuccessCheck className="size-20" />
        </div>
      )}
    </form>
  )
}
