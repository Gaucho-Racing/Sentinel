// Shared, non-component helpers for the analytics views. Kept out of
// primitives.tsx so that file can export only components (react-refresh).

// Brand-forward categorical palette. Ordered so the first two are the Gaucho
// Racing brand colors; the rest are chosen to stay legible on both the light
// and dark card backgrounds.
export const PALETTE = [
  "#e105a3", // gr-pink
  "#8412fc", // gr-purple
  "#0ea5e9", // sky
  "#22c55e", // green
  "#f59e0b", // amber
  "#14b8a6", // teal
  "#6366f1", // indigo
  "#ef4444", // red
]

// Theme tokens usable directly as SVG stroke/fill values inside recharts.
export const AXIS_COLOR = "var(--color-muted-foreground)"
export const GRID_COLOR = "var(--color-border)"

export function formatNumber(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`
  if (n >= 1_000) return `${(n / 1_000).toFixed(n >= 10_000 ? 0 : 1)}k`
  return `${n}`
}

// Humanizes an audit action enum (APPLICATION_SECRET_REVEALED) into a title
// ("Application Secret Revealed"). Accepts number too so it can be handed
// straight to recharts tick/label formatters.
export function humanizeAction(action: string | number): string {
  return String(action)
    .toLowerCase()
    .split("_")
    .map((w) => w.charAt(0).toUpperCase() + w.slice(1))
    .join(" ")
}
