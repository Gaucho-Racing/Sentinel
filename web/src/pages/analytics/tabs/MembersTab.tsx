import { useState } from "react"
import {
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  ComposedChart,
  Legend,
  Line,
  Pie,
  PieChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts"

import { ChartCard, ChartTooltip, RangeToggle } from "@/components/analytics/primitives"
import { AXIS_COLOR, GRID_COLOR, PALETTE } from "@/components/analytics/utils"
import {
  useAuthMethods,
  useMemberDemographics,
  useUserGrowth,
  type CategoryCount,
} from "@/lib/analytics"

const MONTH_RANGES: Array<{ value: number; label: string }> = [
  { value: 6, label: "6m" },
  { value: 12, label: "12m" },
  { value: 24, label: "24m" },
]

function monthLabel(value: string | number) {
  const [y, m] = String(value).split("-")
  const d = new Date(Date.UTC(Number(y), Number(m) - 1, 1))
  return `${d.toLocaleString("en-US", { month: "short", timeZone: "UTC" })} '${String(y).slice(2)}`
}

function CategoryBar({
  data,
  color,
}: {
  data: CategoryCount[]
  color: string
}) {
  return (
    <ResponsiveContainer width="100%" height="100%">
      <BarChart data={data} margin={{ top: 8, right: 8, bottom: 0, left: -12 }}>
        <CartesianGrid strokeDasharray="3 3" stroke={GRID_COLOR} vertical={false} />
        <XAxis
          dataKey="label"
          tick={{ fontSize: 11, fill: AXIS_COLOR }}
          tickLine={false}
          axisLine={{ stroke: GRID_COLOR }}
          interval={0}
          angle={data.length > 6 ? -30 : 0}
          textAnchor={data.length > 6 ? "end" : "middle"}
          height={data.length > 6 ? 60 : 30}
        />
        <YAxis
          tick={{ fontSize: 11, fill: AXIS_COLOR }}
          tickLine={false}
          axisLine={false}
          allowDecimals={false}
          width={36}
        />
        <Tooltip cursor={{ fill: "var(--color-muted)", opacity: 0.4 }} content={<ChartTooltip />} />
        <Bar dataKey="count" name="Members" fill={color} radius={[4, 4, 0, 0]} />
      </BarChart>
    </ResponsiveContainer>
  )
}

function Donut({ data }: { data: CategoryCount[] }) {
  return (
    <ResponsiveContainer width="100%" height="100%">
      <PieChart>
        <Pie
          data={data}
          dataKey="count"
          nameKey="label"
          cx="50%"
          cy="50%"
          innerRadius={55}
          outerRadius={85}
          paddingAngle={2}
        >
          {data.map((_, i) => (
            <Cell key={i} stroke="var(--color-card)" strokeWidth={2} fill={PALETTE[i % PALETTE.length]} />
          ))}
        </Pie>
        <Tooltip content={<ChartTooltip />} />
        <Legend
          verticalAlign="middle"
          align="right"
          layout="vertical"
          iconType="circle"
          formatter={(value) => <span className="text-xs text-muted-foreground">{value}</span>}
        />
      </PieChart>
    </ResponsiveContainer>
  )
}

export function MembersTab() {
  const [months, setMonths] = useState<number>(12)
  const growth = useUserGrowth(months)
  const demographics = useMemberDemographics()
  const authMethods = useAuthMethods()

  const authData: CategoryCount[] = authMethods.data
    ? [
        { label: "Email", count: authMethods.data.email },
        { label: "Phone", count: authMethods.data.phone },
        { label: "Discord", count: authMethods.data.discord },
        { label: "Google", count: authMethods.data.google },
        { label: "GitHub", count: authMethods.data.github },
      ].filter((m) => m.count > 0)
    : []

  const gradYear = demographics.data?.by_grad_year ?? []
  const major = demographics.data?.by_major ?? []
  const gradLevel = demographics.data?.by_graduate_level ?? []

  return (
    <div className="space-y-6">
      <ChartCard
        title="Member growth"
        description="New members per month and cumulative roster size."
        action={<RangeToggle value={months} onChange={setMonths} options={MONTH_RANGES} />}
        isLoading={growth.isLoading}
        isEmpty={(growth.data ?? []).every((p) => p.new_users === 0 && p.cumulative === 0)}
        height={320}
      >
        <ResponsiveContainer width="100%" height="100%">
          <ComposedChart data={growth.data ?? []} margin={{ top: 8, right: 8, bottom: 0, left: -12 }}>
            <CartesianGrid strokeDasharray="3 3" stroke={GRID_COLOR} vertical={false} />
            <XAxis
              dataKey="date"
              tickFormatter={monthLabel}
              tick={{ fontSize: 11, fill: AXIS_COLOR }}
              tickLine={false}
              axisLine={{ stroke: GRID_COLOR }}
              minTickGap={16}
            />
            <YAxis
              tick={{ fontSize: 11, fill: AXIS_COLOR }}
              tickLine={false}
              axisLine={false}
              allowDecimals={false}
              width={40}
            />
            <Tooltip content={<ChartTooltip labelFormatter={monthLabel} />} />
            <Legend
              iconType="circle"
              formatter={(value) => <span className="text-xs text-muted-foreground">{value}</span>}
            />
            <Bar dataKey="new_users" name="New members" fill={PALETTE[0]} radius={[4, 4, 0, 0]} />
            <Line
              type="monotone"
              dataKey="cumulative"
              name="Total members"
              stroke={PALETTE[1]}
              strokeWidth={2}
              dot={false}
            />
          </ComposedChart>
        </ResponsiveContainer>
      </ChartCard>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <ChartCard
          title="By graduation year"
          isLoading={demographics.isLoading}
          isEmpty={gradYear.length === 0}
          height={280}
        >
          <CategoryBar data={gradYear} color={PALETTE[1]} />
        </ChartCard>
        <ChartCard
          title="Top majors"
          isLoading={demographics.isLoading}
          isEmpty={major.length === 0}
          height={280}
        >
          <CategoryBar data={major} color={PALETTE[2]} />
        </ChartCard>
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <ChartCard
          title="Graduate level"
          isLoading={demographics.isLoading}
          isEmpty={gradLevel.length === 0}
          height={260}
        >
          <Donut data={gradLevel} />
        </ChartCard>
        <ChartCard
          title="Auth methods in use"
          description="Distinct members per method (multi-auth counts in each)."
          isLoading={authMethods.isLoading}
          isEmpty={authData.length === 0}
          height={260}
        >
          <Donut data={authData} />
        </ChartCard>
      </div>
    </div>
  )
}
