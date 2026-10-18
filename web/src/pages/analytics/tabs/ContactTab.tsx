import { AtSign, Globe2, MapPin, Phone } from "lucide-react"
import { useMemo, useState } from "react"
import { Bar, BarChart, Cell, Pie, PieChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts"

import { ChoroplethMap, type MapScope } from "@/components/analytics/ChoroplethMap"
import { ChartCard, ChartTooltip, RangeToggle, StatCard, StatCardSkeleton } from "@/components/analytics/primitives"
import { AXIS_COLOR, PALETTE } from "@/components/analytics/utils"
import { useContactAnalytics, type DomainCount } from "@/lib/analytics"

const MAP_SCOPES: Array<{ value: MapScope; label: string }> = [
  { value: "us", label: "United States" },
  { value: "world", label: "World" },
]

const KIND_LABELS: Record<DomainCount["kind"], string> = {
  school: "School",
  team: "Gaucho Racing",
  personal: "Personal",
  other: "Other",
}

const KIND_COLORS: Record<DomainCount["kind"], string> = {
  school: PALETTE[0],
  team: PALETTE[1],
  personal: PALETTE[2],
  other: "var(--color-muted-foreground)",
}

const regionNames = new Intl.DisplayNames(["en"], { type: "region" })

function countryName(code: string) {
  try {
    return regionNames.of(code) ?? code
  } catch {
    return code
  }
}

function percent(part: number, total: number) {
  return total === 0 ? "—" : `${Math.round((part / total) * 100)}%`
}

function HorizontalBars({
  data,
  labelKey,
  colors,
}: {
  data: Array<Record<string, string | number>>
  labelKey: string
  colors: string[]
}) {
  return (
    <ResponsiveContainer width="100%" height="100%">
      <BarChart data={data} layout="vertical" margin={{ top: 0, right: 16, bottom: 0, left: 0 }}>
        <XAxis type="number" hide allowDecimals={false} />
        <YAxis
          type="category"
          dataKey={labelKey}
          width={170}
          tick={{ fontSize: 12, fill: AXIS_COLOR }}
          tickLine={false}
          axisLine={false}
        />
        <Tooltip cursor={{ fill: "var(--color-muted)", opacity: 0.4 }} content={<ChartTooltip />} />
        <Bar dataKey="count" name="Members" radius={[0, 4, 4, 0]} barSize={18}>
          {data.map((_, i) => (
            <Cell key={i} fill={colors[i % colors.length]} />
          ))}
        </Bar>
      </BarChart>
    </ResponsiveContainer>
  )
}

export function ContactTab() {
  const [scope, setScope] = useState<MapScope>("us")
  const contacts = useContactAnalytics()
  const phones = contacts.data?.phones
  const emails = contacts.data?.emails

  const mapValues = useMemo(() => {
    if (!phones) return []
    return scope === "us"
      ? phones.us_states.map((state) => ({ key: state.name ?? state.code, label: state.name ?? state.code, count: state.count, verified: state.verified }))
      : phones.countries.map((country) => ({ key: country.code, label: countryName(country.code), count: country.count, verified: country.verified }))
  }, [phones, scope])

  const areaCodes = (phones?.area_codes ?? []).slice(0, 12).map((entry) => ({
    label: entry.region ? `${entry.area_code} · ${entry.region}` : entry.area_code,
    count: entry.count,
  }))
  const countries = (phones?.countries ?? []).slice(0, 8).map((country) => ({ label: countryName(country.code), count: country.count }))
  const domains = (emails?.domains ?? []).slice(0, 10).map((domain) => ({ label: domain.domain, count: domain.count, kind: domain.kind }))
  const kinds = (emails?.kinds ?? []).map((kind) => ({
    label: KIND_LABELS[kind.label as DomainCount["kind"]] ?? kind.label,
    count: kind.count,
    color: KIND_COLORS[kind.label as DomainCount["kind"]] ?? PALETTE[3],
  }))

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-2 gap-3 md:grid-cols-4">
        {contacts.isLoading || !phones || !emails ? (
          Array.from({ length: 4 }, (_, i) => <StatCardSkeleton key={i} />)
        ) : (
          <>
            <StatCard label="Phone numbers" value={phones.total} sub={`${percent(phones.verified, phones.total)} verified`} icon={Phone} />
            <StatCard label="Countries" value={phones.countries.length} sub={`${phones.us_states.length} US states`} icon={Globe2} />
            <StatCard label="Area codes" value={phones.area_codes.length} sub="US, Canada, and NANP" icon={MapPin} />
            <StatCard label="Email addresses" value={emails.total} sub={`${percent(emails.verified, emails.total)} verified`} icon={AtSign} />
          </>
        )}
      </div>

      <ChartCard
        title="Where members' numbers are from"
        description={scope === "us" ? "Members by state, from their phone number." : "Members by country, from their phone number."}
        action={<RangeToggle value={scope} onChange={setScope} options={MAP_SCOPES} />}
        isLoading={contacts.isLoading}
        isEmpty={mapValues.length === 0}
        emptyText="No phone numbers yet."
        height={380}
      >
        <ChoroplethMap scope={scope} values={mapValues} height={380} />
      </ChartCard>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <ChartCard
          title="Top area codes"
          description="Most common area codes among member phone numbers."
          isLoading={contacts.isLoading}
          isEmpty={areaCodes.length === 0}
          height={Math.max(160, areaCodes.length * 28)}
        >
          <HorizontalBars data={areaCodes} labelKey="label" colors={[PALETTE[0]]} />
        </ChartCard>
        <ChartCard
          title="Countries"
          description="Where members' phone numbers are registered."
          isLoading={contacts.isLoading}
          isEmpty={countries.length === 0}
          height={Math.max(160, areaCodes.length * 28)}
        >
          <HorizontalBars data={countries} labelKey="label" colors={[PALETTE[1]]} />
        </ChartCard>
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        <ChartCard
          title="Email domains"
          description="Domains used by at least 3 members; the rest are grouped as Other."
          isLoading={contacts.isLoading}
          isEmpty={domains.length === 0}
          height={Math.max(160, domains.length * 28)}
          className="lg:col-span-2"
        >
          <HorizontalBars data={domains} labelKey="label" colors={domains.map((domain) => KIND_COLORS[domain.kind])} />
        </ChartCard>
        <ChartCard
          title="Email types"
          description="School, team, personal, and other addresses."
          isLoading={contacts.isLoading}
          isEmpty={kinds.length === 0}
          height={Math.max(160, domains.length * 28)}
        >
          <ResponsiveContainer width="100%" height="100%">
            <PieChart>
              <Pie data={kinds} dataKey="count" nameKey="label" cx="50%" cy="50%" innerRadius={55} outerRadius={85} paddingAngle={2}>
                {kinds.map((kind, i) => (
                  <Cell key={i} stroke="var(--color-card)" strokeWidth={2} fill={kind.color} />
                ))}
              </Pie>
              <Tooltip content={<ChartTooltip />} />
            </PieChart>
          </ResponsiveContainer>
        </ChartCard>
      </div>
    </div>
  )
}
