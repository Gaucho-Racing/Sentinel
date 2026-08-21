import { useState } from "react"
import { Bar, BarChart, CartesianGrid, Cell, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts"

import { ChartCard, ChartTooltip, RangeToggle } from "@/components/analytics/primitives"
import { AXIS_COLOR, GRID_COLOR, humanizeAction, PALETTE } from "@/components/analytics/utils"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Skeleton } from "@/components/ui/skeleton"
import { useAuditEvents, useAuditSummary, type AuditEvent } from "@/lib/analytics"
import { useUsers, userName } from "@/lib/users"

const RANGES: Array<{ value: number; label: string }> = [
  { value: 7, label: "7d" },
  { value: 30, label: "30d" },
  { value: 90, label: "90d" },
]

const ALL = "ALL"

// Actions that read/reveal rather than mutate are worth flagging in red.
const SENSITIVE = new Set(["APPLICATION_SECRET_REVEALED"])

function timestamp(iso: string) {
  const d = new Date(iso)
  return d.toLocaleString("en-US", {
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  })
}

function metaSummary(event: AuditEvent): string {
  if (!event.metadata) return ""
  return Object.entries(event.metadata)
    .filter(([, v]) => v !== "" && v != null)
    .map(([k, v]) => `${k}: ${String(v)}`)
    .join(" · ")
}

export function AuditTab() {
  const [days, setDays] = useState<number>(30)
  const [action, setAction] = useState<string>(ALL)

  const summary = useAuditSummary(days)
  const events = useAuditEvents({ action: action === ALL ? undefined : action, limit: 100 })
  const users = useUsers()

  const actorName = (id: string) => {
    const u = (users.data ?? []).find((x) => x.entity_id === id)
    return u ? userName(u) : id
  }

  const summaryData = summary.data ?? []

  return (
    <div className="space-y-6">
      <ChartCard
        title="Administrative activity"
        description="Recorded admin actions by type. Covers app changes, membership edits, approvals, and secret reveals."
        action={<RangeToggle value={days} onChange={setDays} options={RANGES} />}
        isLoading={summary.isLoading}
        isEmpty={summaryData.length === 0}
        emptyText="No audited actions in this range yet."
        height={Math.max(220, summaryData.length * 40)}
      >
        <ResponsiveContainer width="100%" height="100%">
          <BarChart
            data={summaryData}
            layout="vertical"
            margin={{ top: 4, right: 16, bottom: 4, left: 8 }}
          >
            <CartesianGrid strokeDasharray="3 3" stroke={GRID_COLOR} horizontal={false} />
            <XAxis
              type="number"
              tick={{ fontSize: 11, fill: AXIS_COLOR }}
              tickLine={false}
              axisLine={{ stroke: GRID_COLOR }}
              allowDecimals={false}
            />
            <YAxis
              type="category"
              dataKey="label"
              tickFormatter={humanizeAction}
              tick={{ fontSize: 11, fill: AXIS_COLOR }}
              tickLine={false}
              axisLine={false}
              width={180}
            />
            <Tooltip
              cursor={{ fill: "var(--color-muted)", opacity: 0.4 }}
              content={<ChartTooltip labelFormatter={humanizeAction} />}
            />
            <Bar dataKey="count" name="Events" radius={[0, 4, 4, 0]}>
              {summaryData.map((row, i) => (
                <Cell
                  key={i}
                  fill={SENSITIVE.has(row.label) ? PALETTE[7] : PALETTE[i % PALETTE.length]}
                />
              ))}
            </Bar>
          </BarChart>
        </ResponsiveContainer>
      </ChartCard>

      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0">
          <CardTitle className="text-base">Recent activity</CardTitle>
          <Select value={action} onValueChange={setAction}>
            <SelectTrigger className="w-[200px]">
              <SelectValue placeholder="All actions" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value={ALL}>All actions</SelectItem>
              {summaryData.map((row) => (
                <SelectItem key={row.label} value={row.label}>
                  {humanizeAction(row.label)}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </CardHeader>
        <CardContent className="p-0">
          {events.isLoading ? (
            <div className="space-y-3 p-6">
              {Array.from({ length: 8 }).map((_, i) => (
                <Skeleton key={i} className="h-8 w-full" />
              ))}
            </div>
          ) : (events.data ?? []).length === 0 ? (
            <p className="py-10 text-center text-sm text-muted-foreground">
              No matching audit events.
            </p>
          ) : (
            <ul className="divide-y divide-border">
              {(events.data ?? []).map((event) => (
                <li key={event.id} className="flex flex-wrap items-center gap-x-3 gap-y-1 px-6 py-3">
                  <Badge
                    variant={SENSITIVE.has(event.action) ? "destructive" : "secondary"}
                    className="font-normal"
                  >
                    {humanizeAction(event.action)}
                  </Badge>
                  <span className="text-sm font-medium">{actorName(event.actor_id)}</span>
                  {metaSummary(event) && (
                    <span className="truncate text-xs text-muted-foreground">
                      {metaSummary(event)}
                    </span>
                  )}
                  <span className="ml-auto whitespace-nowrap text-xs text-muted-foreground">
                    {timestamp(event.created_at)}
                  </span>
                </li>
              ))}
            </ul>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
