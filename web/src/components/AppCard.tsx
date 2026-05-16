import { ChevronRight } from "lucide-react"
import { Link } from "react-router-dom"

import type { Application } from "@/lib/applications"

function initial(name: string) {
  return name.slice(0, 1).toUpperCase()
}

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

export function AppCard({
  app,
  lastAccessedAt,
}: {
  app: Application
  lastAccessedAt?: string
}) {
  const accessed = relativeTime(lastAccessedAt)
  return (
    <Link
      to={`/applications/${app.id}`}
      className="group flex flex-col gap-3 rounded-lg border border-border/60 bg-card p-4 transition-colors hover:bg-muted/40"
    >
      <div className="flex items-start justify-between">
        <div className="flex size-10 items-center justify-center overflow-hidden rounded-md bg-gradient-to-br from-gr-pink to-gr-purple text-base font-semibold text-white">
          {app.icon_url ? (
            <img src={app.icon_url} alt={app.name} className="size-full object-cover" />
          ) : (
            initial(app.name)
          )}
        </div>
        <ChevronRight className="size-3.5 text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100" />
      </div>
      <div>
        <p className="text-sm font-medium leading-none">{app.name}</p>
        <p className="mt-1 line-clamp-2 text-xs text-muted-foreground">{app.description}</p>
      </div>
      {accessed && (
        <p className="mt-auto text-[11px] text-muted-foreground">Last accessed {accessed}</p>
      )}
    </Link>
  )
}
