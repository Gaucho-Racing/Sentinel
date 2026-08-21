import { useState } from "react"
import { Bar, BarChart, CartesianGrid, Legend, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts"

import {
  ChartCard,
  ChartTooltip,
  RangeToggle,
  StatCard,
  StatCardSkeleton,
} from "@/components/analytics/primitives"
import { AXIS_COLOR, GRID_COLOR, PALETTE } from "@/components/analytics/utils"
import { useGroupMembership, useJoinRequestFunnel } from "@/lib/analytics"

const RANGES: Array<{ value: number; label: string }> = [
  { value: 30, label: "30d" },
  { value: 90, label: "90d" },
  { value: 365, label: "1y" },
]

export function GroupsTab() {
  const [days, setDays] = useState<number>(90)
  const membership = useGroupMembership()
  const funnel = useJoinRequestFunnel(days)

  const groups = (membership.data ?? []).filter((g) => g.member_count > 0)
  const f = funnel.data

  return (
    <div className="space-y-6">
      <ChartCard
        title="Group membership by source"
        description="How members land in each group — direct, conditional, or synced from Discord."
        isLoading={membership.isLoading}
        isEmpty={groups.length === 0}
        height={Math.max(260, groups.length * 38)}
      >
        <ResponsiveContainer width="100%" height="100%">
          <BarChart data={groups} layout="vertical" margin={{ top: 4, right: 16, bottom: 4, left: 8 }}>
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
              dataKey="name"
              tick={{ fontSize: 11, fill: AXIS_COLOR }}
              tickLine={false}
              axisLine={false}
              width={140}
            />
            <Tooltip cursor={{ fill: "var(--color-muted)", opacity: 0.4 }} content={<ChartTooltip />} />
            <Legend
              iconType="circle"
              formatter={(value) => <span className="text-xs text-muted-foreground">{value}</span>}
            />
            <Bar dataKey="direct" name="Direct" stackId="s" fill={PALETTE[0]} radius={[0, 0, 0, 0]} />
            <Bar dataKey="conditional" name="Conditional" stackId="s" fill={PALETTE[2]} />
            <Bar dataKey="discord" name="Discord" stackId="s" fill={PALETTE[5]} radius={[0, 4, 4, 0]} />
          </BarChart>
        </ResponsiveContainer>
      </ChartCard>

      <div className="grid grid-cols-2 gap-3 md:grid-cols-4">
        {funnel.isLoading ? (
          Array.from({ length: 4 }).map((_, i) => <StatCardSkeleton key={i} />)
        ) : (
          <>
            <StatCard label="Pending now" value={f?.pending ?? 0} />
            <StatCard label={`Approved (${days}d)`} value={f?.approved ?? 0} />
            <StatCard label={`Rejected (${days}d)`} value={f?.rejected ?? 0} />
            <StatCard
              label="Median decision"
              value={f ? `${f.median_decision_hours.toFixed(1)}h` : "—"}
            />
          </>
        )}
      </div>

      <ChartCard
        title="Join request outcomes"
        description="Requests reviewed within the selected window."
        action={<RangeToggle value={days} onChange={setDays} options={RANGES} />}
        isLoading={funnel.isLoading}
        isEmpty={!f || (f.pending === 0 && f.approved === 0 && f.rejected === 0)}
        height={260}
      >
        <ResponsiveContainer width="100%" height="100%">
          <BarChart
            data={[
              { label: "Pending", count: f?.pending ?? 0 },
              { label: "Approved", count: f?.approved ?? 0 },
              { label: "Rejected", count: f?.rejected ?? 0 },
            ]}
            margin={{ top: 8, right: 8, bottom: 0, left: -12 }}
          >
            <CartesianGrid strokeDasharray="3 3" stroke={GRID_COLOR} vertical={false} />
            <XAxis
              dataKey="label"
              tick={{ fontSize: 11, fill: AXIS_COLOR }}
              tickLine={false}
              axisLine={{ stroke: GRID_COLOR }}
            />
            <YAxis
              tick={{ fontSize: 11, fill: AXIS_COLOR }}
              tickLine={false}
              axisLine={false}
              allowDecimals={false}
              width={36}
            />
            <Tooltip cursor={{ fill: "var(--color-muted)", opacity: 0.4 }} content={<ChartTooltip />} />
            <Bar dataKey="count" name="Requests" radius={[4, 4, 0, 0]} fill={PALETTE[1]} />
          </BarChart>
        </ResponsiveContainer>
      </ChartCard>
    </div>
  )
}
