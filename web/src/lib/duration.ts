// Presentation helpers for time-boxed memberships and join requests.

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
