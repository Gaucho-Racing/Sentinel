import type { ComponentType, ReactNode } from "react"

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { cn } from "@/lib/utils"

export function StatCard({
  label,
  value,
  sub,
  icon: Icon,
}: {
  label: string
  value: ReactNode
  sub?: string
  icon?: ComponentType<{ className?: string }>
}) {
  return (
    <Card>
      <CardContent className="p-5">
        <div className="flex items-center justify-between">
          <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
            {label}
          </p>
          {Icon && <Icon className="size-4 text-muted-foreground" />}
        </div>
        <p className="mt-2 text-2xl font-semibold tracking-tight">{value}</p>
        {sub && <p className="mt-1 text-xs text-muted-foreground">{sub}</p>}
      </CardContent>
    </Card>
  )
}

export function StatCardSkeleton() {
  return (
    <Card>
      <CardContent className="p-5">
        <Skeleton className="h-3 w-24" />
        <Skeleton className="mt-3 h-7 w-16" />
        <Skeleton className="mt-2 h-3 w-20" />
      </CardContent>
    </Card>
  )
}

// ChartCard is the boxed container every chart lives in. It owns the header,
// the fixed-height plot area, and the loading/empty states — children are
// expected to render their own <ResponsiveContainer>.
export function ChartCard({
  title,
  description,
  action,
  isLoading,
  isEmpty,
  emptyText = "No data for this range yet.",
  height = 300,
  className,
  children,
}: {
  title: string
  description?: string
  action?: ReactNode
  isLoading?: boolean
  isEmpty?: boolean
  emptyText?: string
  height?: number
  className?: string
  children: ReactNode
}) {
  return (
    <Card className={className}>
      <CardHeader className="flex flex-row items-start justify-between space-y-0">
        <div className="space-y-1">
          <CardTitle className="text-base">{title}</CardTitle>
          {description && <CardDescription>{description}</CardDescription>}
        </div>
        {action}
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <Skeleton className="w-full rounded-md" style={{ height }} />
        ) : isEmpty ? (
          <div
            className="flex items-center justify-center text-sm text-muted-foreground"
            style={{ height }}
          >
            {emptyText}
          </div>
        ) : (
          <div style={{ height }}>{children}</div>
        )}
      </CardContent>
    </Card>
  )
}

// ChartTooltip is a themed replacement for recharts' default tooltip. Pass it
// to a <Tooltip content={<ChartTooltip />} /> — recharts injects active/payload.
export function ChartTooltip({
  active,
  payload,
  label,
  labelFormatter,
}: {
  active?: boolean
  payload?: Array<{ name?: string; value?: number | string; color?: string; dataKey?: string }>
  label?: string | number
  labelFormatter?: (label: string | number) => string
}) {
  if (!active || !payload || payload.length === 0) return null
  return (
    <div className="rounded-lg border border-border bg-card px-3 py-2 text-xs shadow-md">
      {label !== undefined && (
        <p className="mb-1 font-medium text-foreground">
          {labelFormatter ? labelFormatter(label) : label}
        </p>
      )}
      <div className="space-y-0.5">
        {payload.map((entry, i) => (
          <div key={i} className="flex items-center gap-2">
            <span
              className="size-2 shrink-0 rounded-[2px]"
              style={{ backgroundColor: entry.color }}
            />
            <span className="text-muted-foreground">{entry.name ?? entry.dataKey}</span>
            <span className="ml-auto font-medium text-foreground">{entry.value}</span>
          </div>
        ))}
      </div>
    </div>
  )
}

// RangeToggle is a compact segmented control for picking a time window. Values
// are opaque to the control — the parent decides what the numbers mean (days,
// months) and how to query.
export function RangeToggle<T extends number>({
  value,
  onChange,
  options,
}: {
  value: T
  onChange: (value: T) => void
  options: Array<{ value: T; label: string }>
}) {
  return (
    <div className="inline-flex rounded-md border border-border p-0.5">
      {options.map((opt) => (
        <button
          key={opt.value}
          type="button"
          onClick={() => onChange(opt.value)}
          className={cn(
            "rounded-[5px] px-2.5 py-1 text-xs font-medium transition-colors",
            value === opt.value
              ? "bg-muted text-foreground"
              : "text-muted-foreground hover:text-foreground",
          )}
        >
          {opt.label}
        </button>
      ))}
    </div>
  )
}

const WEEKDAYS = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"]

// Heatmap renders a weekday × hour grid of sign-in volume. Intensity is the
// cell count scaled against the busiest cell, tinted with the brand pink.
export function Heatmap({ cells }: { cells: Array<{ weekday: number; hour: number; count: number }> }) {
  const grid = new Map<string, number>()
  let max = 0
  for (const c of cells) {
    grid.set(`${c.weekday}-${c.hour}`, c.count)
    if (c.count > max) max = c.count
  }

  const intensity = (count: number) => {
    if (count <= 0) return "var(--color-muted)"
    const alpha = 0.15 + 0.85 * (count / max)
    return `rgba(225, 5, 163, ${alpha.toFixed(3)})`
  }

  return (
    <div className="overflow-x-auto">
      <div className="min-w-[560px]">
        <div className="flex flex-col gap-1">
          {WEEKDAYS.map((day, wd) => (
            <div key={day} className="flex items-center gap-1">
              <span className="w-8 shrink-0 text-[10px] text-muted-foreground">{day}</span>
              <div className="flex flex-1 gap-1">
                {Array.from({ length: 24 }).map((_, hour) => {
                  const count = grid.get(`${wd}-${hour}`) ?? 0
                  return (
                    <div
                      key={hour}
                      title={`${day} ${hour}:00 — ${count} sign-in${count === 1 ? "" : "s"}`}
                      className="aspect-square flex-1 rounded-[3px] border border-border/40"
                      style={{ backgroundColor: intensity(count) }}
                    />
                  )
                })}
              </div>
            </div>
          ))}
          <div className="flex items-center gap-1">
            <span className="w-8 shrink-0" />
            <div className="flex flex-1 justify-between px-0.5 text-[10px] text-muted-foreground">
              <span>12a</span>
              <span>6a</span>
              <span>12p</span>
              <span>6p</span>
              <span>11p</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
