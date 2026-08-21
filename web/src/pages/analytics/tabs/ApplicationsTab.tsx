import { useState } from "react"
import { Bar, BarChart, CartesianGrid, Cell, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts"

import { ChartCard, ChartTooltip, RangeToggle } from "@/components/analytics/primitives"
import { AXIS_COLOR, formatNumber, GRID_COLOR, PALETTE } from "@/components/analytics/utils"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { useTopApplications } from "@/lib/analytics"

const RANGES: Array<{ value: number; label: string }> = [
  { value: 7, label: "7d" },
  { value: 30, label: "30d" },
  { value: 90, label: "90d" },
]

export function ApplicationsTab() {
  const [days, setDays] = useState<number>(30)
  const top = useTopApplications(days, 12)
  const data = top.data ?? []

  return (
    <div className="space-y-6">
      <ChartCard
        title="Most-used applications"
        description="Ranked by sign-ins through Sentinel."
        action={<RangeToggle value={days} onChange={setDays} options={RANGES} />}
        isLoading={top.isLoading}
        isEmpty={data.length === 0}
        emptyText="No application sign-ins in this range."
        height={Math.max(240, data.length * 34)}
      >
        <ResponsiveContainer width="100%" height="100%">
          <BarChart data={data} layout="vertical" margin={{ top: 4, right: 16, bottom: 4, left: 8 }}>
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
            <Bar dataKey="logins" name="Sign-ins" radius={[0, 4, 4, 0]}>
              {data.map((_, i) => (
                <Cell key={i} fill={PALETTE[i % PALETTE.length]} />
              ))}
            </Bar>
          </BarChart>
        </ResponsiveContainer>
      </ChartCard>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Application breakdown</CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          {top.isLoading ? (
            <div className="space-y-3 p-6">
              {Array.from({ length: 6 }).map((_, i) => (
                <Skeleton key={i} className="h-6 w-full" />
              ))}
            </div>
          ) : data.length === 0 ? (
            <p className="py-10 text-center text-sm text-muted-foreground">Nothing to show.</p>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-border text-left text-xs uppercase tracking-wide text-muted-foreground">
                    <th className="px-6 py-3 font-medium">Application</th>
                    <th className="px-6 py-3 text-right font-medium">Sign-ins</th>
                    <th className="px-6 py-3 text-right font-medium">Members</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border">
                  {data.map((app) => (
                    <tr key={app.client_id}>
                      <td className="px-6 py-3">
                        <div className="font-medium">{app.name}</div>
                        <div className="font-mono text-xs text-muted-foreground">{app.client_id}</div>
                      </td>
                      <td className="px-6 py-3 text-right">{formatNumber(app.logins)}</td>
                      <td className="px-6 py-3 text-right">{formatNumber(app.unique_users)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
