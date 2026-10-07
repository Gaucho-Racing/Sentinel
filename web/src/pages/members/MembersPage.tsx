import { LayoutGrid, Network, Rows3, Search, X } from "lucide-react"
import { lazy, Suspense, useMemo, useState } from "react"
import { Link, useSearchParams } from "react-router-dom"

import { MemberAvatar, MemberCard, MemberLinkedAccounts } from "@/components/MemberCard"
import { PageContainer, PageHeader } from "@/components/PageContainer"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Skeleton } from "@/components/ui/skeleton"
import { fuzzyFilter } from "@/lib/fuzzy"
import { useAllGroupMembers, useGroups } from "@/lib/groups"
import { academicSummary, userName, useUsers } from "@/lib/users"
import { cn } from "@/lib/utils"

// Split out: React Flow and dagre only matter for the org chart, and share the
// chunk with the group graph rather than landing in the entry bundle.
const OrgChartView = lazy(() => import("./org/OrgChartView"))

const VIEWS = [
  { id: "table", label: "Table", icon: Rows3 },
  { id: "grid", label: "Cards", icon: LayoutGrid },
  { id: "org", label: "Org chart", icon: Network },
] as const

type ViewID = (typeof VIEWS)[number]["id"]

const ANY = "__any__"
const PAGE_SIZE = 24

function isViewID(value: string | null): value is ViewID {
  return VIEWS.some((v) => v.id === value)
}

// Only offer a filter value that at least one member actually has, so the
// dropdowns stay short and never produce an empty result on their own.
function distinct(values: (string | number)[]): string[] {
  return [...new Set(values.map(String).map((v) => v.trim()).filter(Boolean))].sort()
}

function FilterSelect({
  label,
  value,
  options,
  onChange,
  format,
}: {
  label: string
  value: string
  options: string[]
  onChange: (next: string) => void
  format?: (value: string) => string
}) {
  if (options.length === 0) return null
  return (
    <Select value={value} onValueChange={onChange}>
      <SelectTrigger size="sm" className="w-auto min-w-28" aria-label={label}>
        <SelectValue placeholder={label} />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value={ANY}>{label}: any</SelectItem>
        {options.map((option) => (
          <SelectItem key={option} value={option}>
            {format ? format(option) : option}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}

export default function MembersPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const viewParam = searchParams.get("view")
  const view: ViewID = isViewID(viewParam) ? viewParam : "table"
  const isOrg = view === "org"

  const [query, setQuery] = useState("")
  const [groupID, setGroupID] = useState(ANY)
  const [gradYear, setGradYear] = useState(ANY)
  const [level, setLevel] = useState(ANY)
  const [major, setMajor] = useState(ANY)
  const [role, setRole] = useState(ANY)
  const [page, setPage] = useState(1)

  const usersQuery = useUsers({ enabled: !isOrg })
  const groupsQuery = useGroups(!isOrg)
  const membershipsQuery = useAllGroupMembers(!isOrg)

  const members = useMemo(() => usersQuery.data ?? [], [usersQuery.data])
  const groups = useMemo(() => groupsQuery.data ?? [], [groupsQuery.data])

  const groupMembers = useMemo(
    () => new Set(
      (membershipsQuery.data ?? [])
        .filter((membership) => membership.group_id === groupID)
        .map((membership) => membership.entity_id),
    ),
    [membershipsQuery.data, groupID],
  )

  const options = useMemo(
    () => ({
      gradYears: distinct(members.map((m) => (m.graduation_year > 0 ? m.graduation_year : ""))),
      levels: distinct(members.map((m) => m.graduate_level)),
      majors: distinct(members.map((m) => m.major)),
      roles: distinct(members.map((m) => m.initial_role)),
    }),
    [members],
  )

  const needle = query.trim()
  const visible = useMemo(() => {
    const filtered = members.filter((m) => {
      if (gradYear !== ANY && String(m.graduation_year) !== gradYear) return false
      if (level !== ANY && m.graduate_level !== level) return false
      if (major !== ANY && m.major !== major) return false
      if (role !== ANY && m.initial_role !== role) return false
      if (groupID !== ANY && !groupMembers.has(m.entity_id)) return false
      return true
    })
    return needle
      ? fuzzyFilter(filtered, needle, (m) => [
          `${m.first_name} ${m.last_name}`,
          m.username,
          m.email,
          m.major,
        ])
      : [...filtered].sort((a, b) => userName(a).localeCompare(userName(b)))
  }, [members, groupMembers, needle, groupID, gradYear, level, major, role])

  const pageCount = Math.max(1, Math.ceil(visible.length / PAGE_SIZE))
  const currentPage = Math.min(page, pageCount)
  const pageStart = (currentPage - 1) * PAGE_SIZE
  const pageMembers = visible.slice(pageStart, pageStart + PAGE_SIZE)

  const setView = (next: ViewID) => {
    setSearchParams(
      (params) => {
        const updated = new URLSearchParams(params)
        if (next === "table") updated.delete("view")
        else updated.set("view", next)
        return updated
      },
      { replace: true },
    )
  }

  const filtersActive =
    groupID !== ANY || gradYear !== ANY || level !== ANY || major !== ANY || role !== ANY
  const clearFilters = () => {
    setPage(1)
    setGroupID(ANY)
    setGradYear(ANY)
    setLevel(ANY)
    setMajor(ANY)
    setRole(ANY)
  }

  const isLoading = usersQuery.isLoading

  return (
    <PageContainer
      className={cn(isOrg && "flex h-[calc(100svh-3.5rem)] max-w-none flex-col py-6")}
    >
      <div className="mb-6">
        <PageHeader
          title="Members"
          description="Everyone with a Sentinel account, their email, and linked accounts."
        />
      </div>

      <div className="mb-4 flex flex-wrap items-center gap-2">
        {!isOrg && (
          <>
            <div className="relative min-w-56 flex-1 sm:max-w-xs">
              <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                type="search"
                placeholder="Search members…"
                value={query}
                onChange={(e) => {
                  setQuery(e.target.value)
                  setPage(1)
                }}
                className="h-8 pl-9"
              />
            </div>

            <Select value={groupID} onValueChange={(value) => {
              setGroupID(value)
              setPage(1)
            }}>
              <SelectTrigger size="sm" className="w-auto min-w-28" aria-label="Group">
                <SelectValue placeholder="Group" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value={ANY}>Group: any</SelectItem>
                {[...groups]
                  .sort((a, b) => a.name.localeCompare(b.name))
                  .map((group) => (
                    <SelectItem key={group.id} value={group.id}>
                      {group.name}
                    </SelectItem>
                  ))}
              </SelectContent>
            </Select>

            <FilterSelect
              label="Year"
              value={gradYear}
              options={options.gradYears}
              onChange={(value) => {
                setGradYear(value)
                setPage(1)
              }}
            />
            <FilterSelect
              label="Level"
              value={level}
              options={options.levels}
              onChange={(value) => {
                setLevel(value)
                setPage(1)
              }}
            />
            <FilterSelect
              label="Major"
              value={major}
              options={options.majors}
              onChange={(value) => {
                setMajor(value)
                setPage(1)
              }}
            />
            <FilterSelect
              label="Role"
              value={role}
              options={options.roles}
              onChange={(value) => {
                setRole(value)
                setPage(1)
              }}
            />

            {filtersActive && (
              <Button variant="ghost" size="sm" onClick={clearFilters}>
                <X className="size-3.5" />
                Clear
              </Button>
            )}
          </>
        )}

        <div className="ml-auto flex items-center gap-0.5 rounded-lg border border-border/60 bg-card p-0.5">
          {VIEWS.map((option) => (
            <Button
              key={option.id}
              variant={view === option.id ? "secondary" : "ghost"}
              size="sm"
              aria-pressed={view === option.id}
              onClick={() => setView(option.id)}
              className={cn(view !== option.id && "text-muted-foreground")}
            >
              <option.icon className="size-3.5" />
              {option.label}
            </Button>
          ))}
        </div>
      </div>

      {!isOrg && (
        <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
          <p className="text-xs text-muted-foreground" aria-live="polite">
            {isLoading
              ? "Loading members…"
              : visible.length === 0
                ? `0 of ${members.length} members`
                : `${pageStart + 1}–${pageStart + pageMembers.length} of ${visible.length} members${filtersActive || needle ? ` (${members.length} total)` : ""}`}
          </p>
          {!isLoading && visible.length > 0 && (
            <nav aria-label="Members pagination" className="flex items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                disabled={currentPage === 1}
                onClick={() => setPage(currentPage - 1)}
              >
                Previous
              </Button>
              <span className="text-xs text-muted-foreground">
                Page {currentPage} of {pageCount}
              </span>
              <Button
                variant="outline"
                size="sm"
                disabled={currentPage === pageCount}
                onClick={() => setPage(currentPage + 1)}
              >
                Next
              </Button>
            </nav>
          )}
        </div>
      )}

      {isOrg ? (
        <Suspense
          fallback={
            <div className="flex min-h-0 flex-1 overflow-hidden rounded-xl border border-border/60">
              <Skeleton className="size-full rounded-none" />
            </div>
          }
        >
          <OrgChartView />
        </Suspense>
      ) : isLoading ? (
        <div className="space-y-2">
          {Array.from({ length: 8 }).map((_, i) => (
            <Skeleton key={i} className="h-14 rounded-lg" />
          ))}
        </div>
      ) : visible.length === 0 ? (
        <p className="py-12 text-center text-sm text-muted-foreground">
          {needle ? `No members match "${query}".` : "No members match these filters."}
        </p>
      ) : view === "grid" ? (
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {pageMembers.map((member) => (
            <MemberCard key={member.entity_id} member={member} />
          ))}
        </div>
      ) : (
        <div className="overflow-hidden rounded-lg border border-border/60">
          <div className="hidden grid-cols-3 gap-4 border-b border-border bg-muted/40 px-4 py-2 text-xs font-medium text-muted-foreground lg:grid" aria-hidden="true">
            <span>Member</span>
            <span>Email</span>
            <span>Linked accounts</span>
          </div>
          <ul className="divide-y divide-border">
            {pageMembers.map((member) => {
              const academic = academicSummary(member)
              return (
                <li key={member.entity_id}>
                  <Link
                    to={`/members/${member.entity_id}`}
                    className="grid gap-3 px-4 py-3 transition-colors hover:bg-muted/40 lg:grid-cols-3 lg:items-center lg:gap-4"
                  >
                    <div className="flex min-w-0 items-center gap-3">
                      <MemberAvatar member={member} className="size-8" fallbackClassName="text-xs" />
                      <div className="min-w-0 flex-1">
                        <p className="truncate text-sm font-medium leading-none">
                          {userName(member)}
                        </p>
                        <p className="mt-1 truncate text-xs text-muted-foreground">
                          {member.username ? `@${member.username}` : member.entity_id}
                        </p>
                        {academic && (
                          <p className="mt-1 truncate text-xs text-muted-foreground">
                            {academic}
                          </p>
                        )}
                      </div>
                    </div>
                    <div className="min-w-0">
                      <p className="mb-1 text-xs text-muted-foreground lg:sr-only">Email</p>
                      <p className="break-all text-sm">{member.email || "No email"}</p>
                    </div>
                    <div className="min-w-0">
                      <p className="mb-1 text-xs text-muted-foreground lg:sr-only">Linked accounts</p>
                      <MemberLinkedAccounts entityID={member.entity_id} />
                    </div>
                  </Link>
                </li>
              )
            })}
          </ul>
        </div>
      )}
    </PageContainer>
  )
}
