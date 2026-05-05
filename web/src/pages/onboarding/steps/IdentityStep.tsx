import { Check, Loader2, X } from "lucide-react"
import { useEffect, useState } from "react"

import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { api } from "@/lib/api"
import { cn } from "@/lib/utils"
import type { StepProps } from "@/pages/onboarding/types"

const DEBOUNCE_MS = 400
const MIN_LENGTH = 3

type Availability = "idle" | "checking" | "available" | "taken" | "error"

export function IdentityStep({ data, update }: StepProps) {
  const [availability, setAvailability] = useState<Availability>("idle")

  useEffect(() => {
    const username = data.username.trim()
    if (username.length < MIN_LENGTH) {
      setAvailability("idle")
      return
    }
    setAvailability("checking")
    let cancelled = false
    const timer = setTimeout(() => {
      api
        .get<{ available: boolean }>("/users/check-username", { params: { username } })
        .then((res) => {
          if (cancelled) return
          setAvailability(res.data.available ? "available" : "taken")
        })
        .catch(() => {
          if (cancelled) return
          setAvailability("error")
        })
    }, DEBOUNCE_MS)
    return () => {
      cancelled = true
      clearTimeout(timer)
    }
  }, [data.username])

  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <h2 className="text-lg font-semibold tracking-tight">Who are you?</h2>
        <p className="text-sm text-muted-foreground">
          Your name and a username for the team to know you by.
        </p>
      </div>

      <div className="space-y-4">
        <div className="grid grid-cols-2 gap-3">
          <div className="space-y-2">
            <Label htmlFor="firstName">First name</Label>
            <Input
              id="firstName"
              autoComplete="given-name"
              value={data.firstName}
              onChange={(e) => update({ firstName: e.target.value })}
              required
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="lastName">Last name</Label>
            <Input
              id="lastName"
              autoComplete="family-name"
              value={data.lastName}
              onChange={(e) => update({ lastName: e.target.value })}
              required
            />
          </div>
        </div>

        <div className="space-y-2">
          <Label htmlFor="username">Username</Label>
          <Input
            id="username"
            autoComplete="username"
            placeholder="bharat"
            value={data.username}
            onChange={(e) => update({ username: e.target.value })}
            required
          />
          <UsernameStatus availability={availability} />
        </div>
      </div>
    </div>
  )
}

function UsernameStatus({ availability }: { availability: Availability }) {
  if (availability === "idle") {
    return (
      <p className="text-xs text-muted-foreground">
        Lowercase letters, numbers, dots, and underscores. Used in URLs and mentions.
      </p>
    )
  }

  const config = {
    checking: { Icon: Loader2, text: "Checking…", className: "text-muted-foreground", spin: true },
    available: { Icon: Check, text: "Username is available", className: "text-emerald-500", spin: false },
    taken: { Icon: X, text: "Username is already taken", className: "text-destructive", spin: false },
    error: { Icon: X, text: "Couldn't check availability", className: "text-muted-foreground", spin: false },
  }[availability]

  return (
    <p className={cn("flex items-center gap-1.5 text-xs", config.className)}>
      <config.Icon className={cn("size-3.5", config.spin && "animate-spin")} />
      {config.text}
    </p>
  )
}
