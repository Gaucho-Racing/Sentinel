// Presentation + input helpers for time-boxed memberships and join requests.

export type DurationPreset = "1w" | "1mo" | "6mo" | "1y"
export type DurationUnit = "hours" | "days" | "weeks" | "months"

export const DURATION_PRESETS: { value: DurationPreset; label: string }[] = [
  { value: "1w", label: "1 week" },
  { value: "1mo", label: "1 month" },
  { value: "6mo", label: "6 months" },
  { value: "1y", label: "1 year" },
]

// Per-unit caps that keep total duration <= 1 year, matching the backend
// validation in core/api/group.go::validateMembershipExpiration.
export const MAX_BY_UNIT: Record<DurationUnit, number> = {
  hours: 8760,
  days: 365,
  weeks: 52,
  months: 12,
}

export function addPreset(now: Date, preset: DurationPreset): Date {
  const d = new Date(now)
  switch (preset) {
    case "1w":
      return new Date(d.getTime() + 7 * 86_400_000)
    case "1mo":
      d.setMonth(d.getMonth() + 1)
      return d
    case "6mo":
      d.setMonth(d.getMonth() + 6)
      return d
    case "1y":
      d.setFullYear(d.getFullYear() + 1)
      return d
  }
}

export function addCustom(now: Date, amount: number, unit: DurationUnit): Date {
  const d = new Date(now)
  switch (unit) {
    case "hours":
      return new Date(d.getTime() + amount * 3_600_000)
    case "days":
      return new Date(d.getTime() + amount * 86_400_000)
    case "weeks":
      return new Date(d.getTime() + amount * 7 * 86_400_000)
    case "months":
      d.setMonth(d.getMonth() + amount)
      return d
  }
}


export function formatAbsoluteDate(iso: string): string {
  if (!iso) return ""
  return new Date(iso).toLocaleDateString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
  })
}

// "in 30 days", "in 6 months", "expired" — relative to right now.
export function formatExpiresIn(iso: string): string {
  if (!iso) return ""
  const ms = new Date(iso).getTime() - Date.now()
  if (ms <= 0) return "expired"
  return "in " + formatDurationMS(ms)
}

// "30 days", "6 months", "1 year" — span between two ISO dates. Used for
// "requested X access" displays where we want to show the originally chosen
// duration (created_at → expires_at) rather than the time remaining.
export function formatDurationBetween(startIso: string, endIso: string): string {
  if (!startIso || !endIso) return ""
  return formatDurationMS(new Date(endIso).getTime() - new Date(startIso).getTime())
}

function formatDurationMS(ms: number): string {
  if (ms <= 0) return ""
  const days = Math.round(ms / 86_400_000)
  if (days < 1) return "<1 day"
  if (days === 1) return "1 day"
  if (days < 30) return `${days} days`
  const months = Math.round(days / 30)
  if (months < 12) return `${months} ${months === 1 ? "month" : "months"}`
  const years = Math.round(months / 12)
  return `${years} ${years === 1 ? "year" : "years"}`
}
