import { useState } from "react"
import { Area, AreaChart, CartesianGrid, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts"

import { ChartCard, ChartTooltip, Heatmap, RangeToggle } from "@/components/analytics/primitives"
import { AXIS_COLOR, GRID_COLOR, PALETTE } from "@/components/analytics/utils"
import { useLoginHeatmap, useLoginTimeSeries } from "@/lib/analytics"

function shortDate(value: string | number) {
  const d = new Date(`${value}T00:00:00Z`)
  return `${d.getUTCMonth() + 1}/${d.getUTCDate()}`
}

const RANGES: Array<{ value: number; label: string }> = [
  { value: 7, label: "7d" },
  { value: 30, label: "30d" },
  { value: 90, label: "90d" },
]

export function SigninsTab() {
  const [days, setDays] = useState<number>(30)
  const series = useLoginTimeSeries(days)
  const heatmap = useLoginHeatmap(90)

  return (
    <div className="space-y-6">
      <ChartCard
        title="Sign-in trend"
        description="Total sign-ins vs. distinct members signing in."
        action={<RangeToggle value={days} onChange={setDays} options={RANGES} />}
        isLoading={series.isLoading}
        isEmpty={(series.data ?? []).every((p) => p.logins === 0)}
        height={340}
      >
        <ResponsiveContainer width="100%" height="100%">
          <AreaChart data={series.data ?? []} margin={{ top: 8, right: 8, bottom: 0, left: -12 }}>
            <defs>
              <linearGradient id="siLogins" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor={PALETTE[0]} stopOpacity={0.35} />
                <stop offset="100%" stopColor={PALETTE[0]} stopOpacity={0} />
              </linearGradient>
              <linearGradient id="siUsers" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor={PALETTE[1]} stopOpacity={0.25} />
                <stop offset="100%" stopColor={PALETTE[1]} stopOpacity={0} />
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
              fill="url(#siLogins)"
            />
            <Area
              type="monotone"
              dataKey="unique_users"
              name="Distinct members"
              stroke={PALETTE[1]}
              strokeWidth={2}
              fill="url(#siUsers)"
            />
          </AreaChart>
        </ResponsiveContainer>
      </ChartCard>

      <ChartCard
        title="When members sign in"
        description="Sign-ins by day of week and hour (UTC), last 90 days."
        isLoading={heatmap.isLoading}
        isEmpty={(heatmap.data ?? []).length === 0}
        height={260}
      >
        <Heatmap cells={heatmap.data ?? []} />
      </ChartCard>
    </div>
  )
}
