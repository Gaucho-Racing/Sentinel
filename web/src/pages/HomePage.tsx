import { ArrowRight, ExternalLink } from "lucide-react"
import { Link } from "react-router-dom"

import { PageContainer } from "@/components/PageContainer"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { type Application, mockApplications, mockRecentLogins, mockUser } from "@/lib/mock"

const RECENT_APPS_LIMIT = 6
const RECENT_ACTIVITY_LIMIT = 5

function formatTime(iso: string) {
  return new Date(iso).toLocaleString(undefined, {
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  })
}

function relativeTime(iso?: string) {
  if (!iso) return "—"
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

function ViewAllLink({ to, children }: { to: string; children: string }) {
  return (
    <Button asChild variant="ghost" size="sm" className="text-muted-foreground">
      <Link to={to}>
        {children}
        <ArrowRight className="ml-1 size-3.5" />
      </Link>
    </Button>
  )
}

function AppCard({ app }: { app: Application }) {
  const initial = app.name.slice(0, 1).toUpperCase()
  return (
    <a
      href={app.url ?? "#"}
      target="_blank"
      rel="noreferrer"
      className="group flex flex-col gap-3 rounded-lg border border-border/60 bg-card p-4 transition-colors hover:bg-muted/40"
    >
      <div className="flex items-start justify-between">
        <div className="flex size-10 items-center justify-center rounded-md bg-gradient-to-br from-gr-pink to-gr-purple text-base font-semibold text-white">
          {initial}
        </div>
        <ExternalLink className="size-3.5 text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100" />
      </div>
      <div>
        <p className="text-sm font-medium leading-none">{app.name}</p>
        <p className="mt-1 line-clamp-2 text-xs text-muted-foreground">{app.description}</p>
      </div>
      <p className="mt-auto text-[11px] text-muted-foreground">
        Last accessed {relativeTime(app.lastAccessedAt)}
      </p>
    </a>
  )
}

export default function HomePage() {
  const firstName = mockUser.name.split(" ")[0]

  const recentApps = [...mockApplications]
    .sort((a, b) => new Date(b.lastAccessedAt ?? 0).getTime() - new Date(a.lastAccessedAt ?? 0).getTime())
    .slice(0, RECENT_APPS_LIMIT)

  const recentActivity = mockRecentLogins.slice(0, RECENT_ACTIVITY_LIMIT)

  return (
    <PageContainer>
      <section className="mb-10">
        <p className="text-sm text-muted-foreground">Welcome back</p>
        <h1 className="mt-1 text-3xl font-semibold tracking-tight">
          Hello, <span className="text-gr-pink">{firstName}</span>
        </h1>
      </section>

      <section className="mb-10">
        <div className="mb-4 flex items-end justify-between">
          <div>
            <h2 className="text-lg font-semibold tracking-tight">Recently accessed</h2>
            <p className="text-sm text-muted-foreground">Apps you've signed into through Sentinel.</p>
          </div>
          <ViewAllLink to="/applications">View all</ViewAllLink>
        </div>
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {recentApps.map((app) => (
            <AppCard key={app.id} app={app} />
          ))}
        </div>
      </section>

      <section>
        <Card>
          <CardHeader className="flex flex-row items-end justify-between space-y-0">
            <div className="space-y-1">
              <CardTitle>Recent activity</CardTitle>
              <CardDescription>Tokens issued for your account, newest first.</CardDescription>
            </div>
            <ViewAllLink to="/settings">View all</ViewAllLink>
          </CardHeader>
          <CardContent className="p-0">
            <ul className="divide-y divide-border">
              {recentActivity.map((login) => (
                <li
                  key={login.id}
                  className="grid grid-cols-1 gap-1 px-6 py-4 sm:grid-cols-[160px_1fr_auto] sm:items-center sm:gap-4"
                >
                  <span className="text-xs tabular-nums text-muted-foreground">{formatTime(login.at)}</span>
                  <div>
                    <p className="text-sm font-medium leading-none">{login.applicationName}</p>
                    <p className="mt-1 font-mono text-xs text-muted-foreground">{login.scope}</p>
                  </div>
                  <span className="font-mono text-xs text-muted-foreground">{login.ipAddress}</span>
                </li>
              ))}
            </ul>
          </CardContent>
        </Card>
      </section>
    </PageContainer>
  )
}
