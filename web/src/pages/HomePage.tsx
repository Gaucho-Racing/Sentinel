import { useQueries, useQuery } from "@tanstack/react-query"
import { ArrowRight } from "lucide-react"
import { Link } from "react-router-dom"

import { AppCard } from "@/components/AppCard"
import { PageContainer } from "@/components/PageContainer"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { api } from "@/lib/api"
import type { Application } from "@/lib/applications"
import { useAuth } from "@/lib/auth"

const RECENT_APPS_LIMIT = 6
const RECENT_ACTIVITY_LIMIT = 5

type EntityLogin = {
  id: string
  entity_id: string
  client_id: string
  scope: string
  access_token_id: string
  refresh_token_id: string
  ip_address: string
  created_at: string
}

type AccessedApplication = Application & { last_accessed_at: string }

function formatTime(iso: string) {
  return new Date(iso).toLocaleString(undefined, {
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  })
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

function ActivityRowSkeleton() {
  return (
    <li className="grid grid-cols-1 gap-1 px-6 py-4 sm:grid-cols-[160px_1fr_auto] sm:items-center sm:gap-4">
      <Skeleton className="h-3 w-32" />
      <div className="space-y-2">
        <Skeleton className="h-3 w-24" />
        <Skeleton className="h-3 w-48" />
      </div>
      <Skeleton className="h-3 w-24" />
    </li>
  )
}

export default function HomePage() {
  const { user, isLoading: userLoading } = useAuth()
  const firstName = user?.user?.first_name
  const userId = user?.user?.id

  const loginsQuery = useQuery({
    queryKey: ["logins", userId],
    queryFn: async () => {
      const res = await api.get<EntityLogin[]>(`/users/${userId}/logins`, {
        params: { limit: RECENT_ACTIVITY_LIMIT },
      })
      return res.data
    },
    enabled: !!userId,
  })

  const recentAppsQuery = useQuery({
    queryKey: ["recentApplications", userId],
    queryFn: async () => {
      const res = await api.get<AccessedApplication[]>(`/users/${userId}/recent-applications`, {
        params: { limit: RECENT_APPS_LIMIT },
      })
      return res.data
    },
    enabled: !!userId,
  })

  const recentApps = recentAppsQuery.data ?? []

  const recentActivity = loginsQuery.data ?? []

  const uniqueClientIds = [...new Set(recentActivity.map((l) => l.client_id))]
  const appQueries = useQueries({
    queries: uniqueClientIds.map((cid) => ({
      queryKey: ["application", cid],
      queryFn: async () => {
        const res = await api.get<Application>(`/applications/client/${cid}`)
        return res.data
      },
      staleTime: 5 * 60 * 1000,
    })),
  })
  const appNameByClientId = new Map<string, string>()
  appQueries.forEach((q, i) => {
    if (q.data?.name) appNameByClientId.set(uniqueClientIds[i], q.data.name)
  })

  return (
    <PageContainer>
      <section className="mb-10">
        <p className="text-sm text-muted-foreground">Welcome back</p>
        <h1 className="mt-1 flex items-center gap-2 text-3xl font-semibold tracking-tight">
          Hello,{" "}
          {userLoading || !firstName ? (
            <Skeleton className="h-8 w-32" />
          ) : (
            <span className="text-gr-pink">{firstName}</span>
          )}
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
          {recentAppsQuery.isLoading
            ? Array.from({ length: 3 }).map((_, i) => (
                <Skeleton key={i} className="h-32 rounded-lg" />
              ))
            : recentApps.length === 0
              ? (
                <p className="col-span-full py-6 text-center text-sm text-muted-foreground">
                  Sign into a team app to see it here.
                </p>
              )
              : recentApps.map((app) => (
                  <AppCard key={app.id} app={app} lastAccessedAt={app.last_accessed_at} />
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
              {loginsQuery.isLoading
                ? Array.from({ length: 3 }).map((_, i) => <ActivityRowSkeleton key={i} />)
                : recentActivity.length === 0
                  ? (
                    <li className="px-6 py-6 text-center text-sm text-muted-foreground">
                      No sign-in activity yet.
                    </li>
                  )
                  : recentActivity.map((login) => (
                    <li
                      key={login.id}
                      className="grid grid-cols-1 gap-1 px-6 py-4 sm:grid-cols-[160px_1fr_auto] sm:items-center sm:gap-4"
                    >
                      <span className="text-xs tabular-nums text-muted-foreground">
                        {formatTime(login.created_at)}
                      </span>
                      <div>
                        <p className="text-sm font-medium leading-none">
                          {login.client_id}
                          {appNameByClientId.get(login.client_id) && (
                            <span className="ml-2 font-normal text-muted-foreground">
                              · {appNameByClientId.get(login.client_id)}
                            </span>
                          )}
                        </p>
                        <p className="mt-1 font-mono text-xs text-muted-foreground">{login.scope}</p>
                      </div>
                      <span className="font-mono text-xs text-muted-foreground">{login.ip_address}</span>
                    </li>
                  ))}
            </ul>
          </CardContent>
        </Card>
      </section>
    </PageContainer>
  )
}
