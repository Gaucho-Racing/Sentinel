import { LayoutGrid, Rows3, Search, X } from "lucide-react"
import { useMemo, useState } from "react"
import { Link, useSearchParams } from "react-router-dom"

import { MemberAvatar, MemberCard } from "@/components/MemberCard"
import { PageContainer, PageHeader } from "@/components/PageContainer"
import { Badge } from "@/components/ui/badge"
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
import { academicSummary, userName, useUsers, type Member } from "@/lib/users"
import { cn } from "@/lib/utils"

const VIEWS = [
  { id: "table", label: "Table", icon: Rows3 },
  { id: "grid", label: "Cards", icon: LayoutGrid },
] as const

type ViewID = (typeof VIEWS)[number]["id"]

const ANY = "__any__"

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

  const [query, setQuery] = useState("")
  const [groupID, setGroupID] = useState(ANY)
  const [gradYear, setGradYear] = useState(ANY)
  const [level, setLevel] = useState(ANY)
  const [major, setMajor] = useState(ANY)
  const [role, setRole] = useState(ANY)

  const usersQuery = useUsers()
  const groupsQuery = useGroups()
  const membershipsQuery = useAllGroupMembers()

  const members = useMemo(() => usersQuery.data ?? [], [usersQuery.data])
  const groups = useMemo(() => groupsQuery.data ?? [], [groupsQuery.data])

  // entity_id -> group names, built from the membership table rather than the
  // `groups` name array on the user so the group filter keys on a stable ID.
  const groupsByEntity = useMemo(() => {
    const nameByID = new Map(groups.map((g) => [g.id, g.name]))
    const map = new Map<string, { id: string; name: string }[]>()
    for (const membership of membershipsQuery.data ?? []) {
      const name = nameByID.get(membership.group_id)
      if (!name) continue
      const list = map.get(membership.entity_id)
      const entry = { id: membership.group_id, name }
      if (list) list.push(entry)
      else map.set(membership.entity_id, [entry])
    }
    for (const list of map.values()) list.sort((a, b) => a.name.localeCompare(b.name))
    return map
  }, [groups, membershipsQuery.data])

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
      if (groupID !== ANY) {
        const held = groupsByEntity.get(m.entity_id) ?? []
        if (!held.some((g) => g.id === groupID)) return false
      }
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
  }, [members, groupsByEntity, needle, groupID, gradYear, level, major, role])

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
    setGroupID(ANY)
    setGradYear(ANY)
    setLevel(ANY)
    setMajor(ANY)
    setRole(ANY)
  }

  const isLoading = usersQuery.isLoading
  const groupNames = (member: Member) =>
    (groupsByEntity.get(member.entity_id) ?? []).map((g) => g.name)

  return (
    <PageContainer>
      <div className="mb-6">
        <PageHeader
          title="Members"
          description="Everyone with a Sentinel account, and the groups they belong to."
        />
      </div>

      <div className="mb-4 flex flex-wrap items-center gap-2">
        <div className="relative min-w-56 flex-1 sm:max-w-xs">
          <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            type="search"
            placeholder="Search members…"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            className="h-8 pl-9"
          />
        </div>

        <Select value={groupID} onValueChange={setGroupID}>
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

        <FilterSelect label="Year" value={gradYear} options={options.gradYears} onChange={setGradYear} />
        <FilterSelect label="Level" value={level} options={options.levels} onChange={setLevel} />
        <FilterSelect label="Major" value={major} options={options.majors} onChange={setMajor} />
        <FilterSelect label="Role" value={role} options={options.roles} onChange={setRole} />

        {filtersActive && (
          <Button variant="ghost" size="sm" onClick={clearFilters}>
            <X className="size-3.5" />
            Clear
          </Button>
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

      <p className="mb-4 text-xs text-muted-foreground">
        {isLoading
          ? "Loading members…"
          : `${visible.length} of ${members.length} member${members.length === 1 ? "" : "s"}`}
      </p>

      {isLoading ? (
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
          {visible.map((member) => (
            <MemberCard key={member.entity_id} member={member} groups={groupNames(member)} />
          ))}
        </div>
      ) : (
        <div className="overflow-hidden rounded-lg border border-border/60">
          <ul className="divide-y divide-border">
            {visible.map((member) => {
              const names = groupNames(member)
              const academic = academicSummary(member)
              return (
                <li key={member.entity_id}>
                  <Link
                    to={`/members/${member.entity_id}`}
                    className="flex items-center gap-3 px-4 py-2.5 transition-colors hover:bg-muted/40"
                  >
                    <MemberAvatar member={member} className="size-8" fallbackClassName="text-xs" />
                    <div className="min-w-0 flex-1">
                      <p className="truncate text-sm font-medium leading-none">
                        {userName(member)}
                      </p>
                      <p className="mt-1 truncate text-xs text-muted-foreground">
                        {member.username ? `@${member.username}` : member.entity_id}
                        {member.email ? ` · ${member.email}` : ""}
                      </p>
                    </div>
                    <p className="hidden min-w-0 flex-1 truncate text-xs text-muted-foreground md:block">
                      {academic}
                    </p>
                    <div className="hidden shrink-0 items-center gap-1 lg:flex">
                      {names.slice(0, 2).map((name) => (
                        <Badge key={name} variant="outline" className="h-5 text-[10px]">
                          {name}
                        </Badge>
                      ))}
                      {names.length > 2 && (
                        <span className="text-[10px] text-muted-foreground">
                          +{names.length - 2}
                        </span>
                      )}
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
