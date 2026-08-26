import { Activity, Boxes, KeyRound, LogIn, UserPlus, Users, UsersRound } from "lucide-react"
import { Link } from "react-router-dom"
import { Area, AreaChart, CartesianGrid, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts"

import { ChartCard, ChartTooltip, StatCard, StatCardSkeleton } from "@/components/analytics/primitives"
import { AXIS_COLOR, formatNumber, GRID_COLOR, PALETTE } from "@/components/analytics/utils"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import {
  useAnalyticsOverview,
  useJoinRequestFunnel,
  useLoginTimeSeries,
  useTopApplications,
} from "@/lib/analytics"

function shortDate(value: string | number) {
  const d = new Date(`${value}T00:00:00Z`)
  return `${d.getUTCMonth() + 1}/${d.getUTCDate()}`
}

export function OverviewTab() {
  const overview = useAnalyticsOverview()
  const series = useLoginTimeSeries(30)
  const funnel = useJoinRequestFunnel(90)
  const topApps = useTopApplications(30, 5)

  const o = overview.data

  const stats = [
    { label: "Members", value: o?.total_users, sub: `+${o?.new_users_30d ?? 0} in 30d`, icon: Users },
    { label: "Active (30d)", value: o?.active_users_30d, sub: `${o?.active_users_7d ?? 0} in last 7d`, icon: Activity },
    { label: "Sign-ins (7d)", value: o?.logins_7d, sub: `${o?.logins_24h ?? 0} in last 24h`, icon: LogIn },
    { label: "New members (30d)", value: o?.new_users_30d, icon: UserPlus },
    { label: "Applications", value: o?.total_applications, icon: Boxes },
    { label: "Groups", value: o?.total_groups, icon: UsersRound },
    { label: "Service accounts", value: o?.total_service_accounts, icon: KeyRound },
    { label: "Pending requests", value: o?.pending_join_requests, sub: `${o?.audit_events_7d ?? 0} audit events 7d`, icon: Users },
  ]

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-2 gap-3 md:grid-cols-4">
        {overview.isLoading
          ? Array.from({ length: 8 }).map((_, i) => <StatCardSkeleton key={i} />)
          : stats.map((s) => (
              <StatCard
                key={s.label}
                label={s.label}
                value={formatNumber(s.value ?? 0)}
                sub={s.sub}
                icon={s.icon}
              />
            ))}
      </div>

      <ChartCard
        title="Sign-in activity"
        description="Total sign-ins and distinct members per day, last 30 days."
        isLoading={series.isLoading}
        isEmpty={(series.data ?? []).every((p) => p.logins === 0)}
      >
        <ResponsiveContainer width="100%" height="100%">
          <AreaChart data={series.data ?? []} margin={{ top: 8, right: 8, bottom: 0, left: -12 }}>
            <defs>
              <linearGradient id="ovLogins" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor={PALETTE[0]} stopOpacity={0.35} />
                <stop offset="100%" stopColor={PALETTE[0]} stopOpacity={0} />
              </linearGradient>
            </defs>
            <CartesianGrid strokeDasharray="3 3" stroke={GRID_COLOR} vertical={false} />
            <XAxis
              dataKey="date"
              tickFormatter={shortDate}
              tick={{ fontSize: 11, fill: AXIS_COLOR }}
              tickLine={false}
              axisLine={{ stroke: GRID_COLOR }}
              minTickGap={24}
            />
            <YAxis
              tick={{ fontSize: 11, fill: AXIS_COLOR }}
              tickLine={false}
              axisLine={false}
              allowDecimals={false}
              width={40}
            />
            <Tooltip content={<ChartTooltip labelFormatter={shortDate} />} />
            <Area
              type="monotone"
              dataKey="logins"
              name="Sign-ins"
              stroke={PALETTE[0]}
              strokeWidth={2}
              fill="url(#ovLogins)"
            />
            <Area
              type="monotone"
              dataKey="unique_users"
              name="Members"
              stroke={PALETTE[1]}
              strokeWidth={2}
              fill="none"
            />
          </AreaChart>
        </ResponsiveContainer>
      </ChartCard>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Top applications</CardTitle>
          </CardHeader>
          <CardContent className="p-0">
            {topApps.isLoading ? (
              <div className="space-y-3 p-6">
                {Array.from({ length: 5 }).map((_, i) => (
                  <Skeleton key={i} className="h-6 w-full" />
                ))}
              </div>
            ) : (topApps.data ?? []).length === 0 ? (
              <p className="py-10 text-center text-sm text-muted-foreground">
                No sign-ins in the last 30 days.
              </p>
            ) : (
              <ul className="divide-y divide-border">
                {(topApps.data ?? []).map((app, i) => (
                  <li key={app.client_id} className="flex items-center gap-3 px-6 py-3">
                    <span className="w-4 text-xs font-medium text-muted-foreground">{i + 1}</span>
                    <span className="truncate text-sm font-medium">{app.name}</span>
                    <span className="ml-auto text-sm text-muted-foreground">
                      {formatNumber(app.logins)}
                    </span>
                  </li>
                ))}
              </ul>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-base">Join requests (90d)</CardTitle>
          </CardHeader>
          <CardContent>
            {funnel.isLoading ? (
              <div className="space-y-3">
                {Array.from({ length: 4 }).map((_, i) => (
                  <Skeleton key={i} className="h-6 w-full" />
                ))}
              </div>
            ) : (
              <dl className="space-y-3 text-sm">
                <div className="flex items-center justify-between">
                  <dt className="text-muted-foreground">Pending (now)</dt>
                  <dd className="font-medium">{funnel.data?.pending ?? 0}</dd>
                </div>
                <div className="flex items-center justify-between">
                  <dt className="text-muted-foreground">Approved</dt>
                  <dd className="font-medium">{funnel.data?.approved ?? 0}</dd>
                </div>
                <div className="flex items-center justify-between">
                  <dt className="text-muted-foreground">Rejected</dt>
                  <dd className="font-medium">{funnel.data?.rejected ?? 0}</dd>
                </div>
                <div className="flex items-center justify-between">
                  <dt className="text-muted-foreground">Median time to decision</dt>
                  <dd className="font-medium">
                    {funnel.data ? `${funnel.data.median_decision_hours.toFixed(1)}h` : "—"}
                  </dd>
                </div>
              </dl>
            )}
            <div className="mt-4">
              <Link to="/groups" className="text-xs font-medium text-gr-pink hover:underline">
                Manage groups →
              </Link>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
