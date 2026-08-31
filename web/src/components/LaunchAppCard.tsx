import { ExternalLink } from "lucide-react"

import { ApplicationIcon } from "@/components/ApplicationIcon"
import type { Application } from "@/lib/applications"

function relativeTime(iso?: string) {
  if (!iso) return null
  const ms = Date.now() - new Date(iso).getTime()
  const minutes = Math.floor(ms / 60_000)
  if (minutes < 1) return "just now"
  if (minutes < 60) return `${minutes}m ago`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}h ago`
  const days = Math.floor(hours / 24)
  if (days < 30) return `${days}d ago`
  const months = Math.floor(days / 30)
  return `${months}mo ago`
}

// LaunchAppCard opens the app's launch_url in a new tab on click. Used on
// the dashboard's "Recently Accessed" section where the user wants to jump
// straight back into the app.
export function LaunchAppCard({
  app,
  lastAccessedAt,
}: {
  app: Application
  lastAccessedAt?: string
}) {
  const accessed = relativeTime(lastAccessedAt)
  return (
    <a
      href={app.launch_url || "#"}
      target="_blank"
      rel="noreferrer"
      className="group flex flex-col gap-3 rounded-lg border border-border/60 bg-card p-4 transition-colors hover:bg-muted/40"
    >
      <div className="flex items-start justify-between">
        <ApplicationIcon
          name={app.name}
          iconUrl={app.icon_url}
          className="size-10 rounded-md"
          fallbackClassName="text-base"
        />
        <ExternalLink className="size-3.5 text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100" />
      </div>
      <div>
        <p className="text-sm font-medium leading-none">{app.name}</p>
        <p className="mt-1 line-clamp-2 text-xs text-muted-foreground">{app.description}</p>
      </div>
      {accessed && (
        <p className="mt-auto text-[11px] text-muted-foreground">Last accessed {accessed}</p>
      )}
    </a>
  )
}
