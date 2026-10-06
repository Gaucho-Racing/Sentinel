import { useQuery } from "@tanstack/react-query"
import { LayoutGrid, Boxes, Plus, Search, Share2 } from "lucide-react"
import { lazy, Suspense, useEffect, useMemo, useRef, useState } from "react"
import { Link, useSearchParams } from "react-router-dom"

import { GroupCard } from "@/components/GroupCard"
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
import { api } from "@/lib/api"
import { useAllApplicationGroupLinks, useApplications } from "@/lib/applications"
import { fuzzyFilter } from "@/lib/fuzzy"
import { ALL_GROUP_SOURCES, SOURCE_LABEL, type Group, type GroupSource } from "@/lib/groups"
import { cn } from "@/lib/utils"

import { GroupsByApplication } from "./views/GroupsByApplication"
import type { GraphHandle } from "./graph/GroupGraphView"

// Split out: React Flow and dagre add ~240 kB to the bundle and only this view
// needs them, so they stay out of the entry chunk until someone opens the tab.
const GroupGraphView = lazy(() => import("./graph/GroupGraphView"))

const VIEWS = [
  { id: "grid", label: "All", icon: LayoutGrid },
  { id: "apps", label: "By app", icon: Boxes },
  { id: "graph", label: "Graph", icon: Share2 },
] as const

type ViewID = (typeof VIEWS)[number]["id"]

const SORTS = [
  { id: "name", label: "Name" },
  { id: "members", label: "Most members" },
  { id: "pending", label: "Most pending" },
  { id: "newest", label: "Newest" },
] as const

type SortID = (typeof SORTS)[number]["id"]

function isViewID(value: string | null): value is ViewID {
  return VIEWS.some((v) => v.id === value)
}

function sortGroups(groups: Group[], sort: SortID): Group[] {
  const byName = (a: Group, b: Group) => a.name.localeCompare(b.name)
  const copy = [...groups]
  switch (sort) {
    case "members":
      return copy.sort((a, b) => b.member_count - a.member_count || byName(a, b))
    case "pending":
      return copy.sort((a, b) => b.pending_count - a.pending_count || byName(a, b))
    case "newest":
      return copy.sort((a, b) => b.created_at.localeCompare(a.created_at) || byName(a, b))
    default:
      return copy.sort(byName)
  }
}

export default function GroupsPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const viewParam = searchParams.get("view")
  const view: ViewID = isViewID(viewParam) ? viewParam : "grid"

  const [query, setQuery] = useState("")
  const [sort, setSort] = useState<SortID>("name")
  const [sources, setSources] = useState<GroupSource[]>([])
  const graphRef = useRef<GraphHandle>(null)
  const [suggestionsOpen, setSuggestionsOpen] = useState(false)
  const searchRef = useRef<HTMLDivElement>(null)

  const groupsQuery = useQuery({
    queryKey: ["groups"],
    queryFn: async () => {
      const res = await api.get<Group[]>("/groups")
      return res.data
    },
  })
  // Only the by-app view needs the link table; the graph view fetches its own
  // copy through the same cache key.
  const applicationsQuery = useApplications(view === "apps")
  const linksQuery = useAllApplicationGroupLinks(view === "apps")

  const groups = useMemo(() => groupsQuery.data ?? [], [groupsQuery.data])
  const needle = query.trim()

  const visible = useMemo(() => {
    const filtered =
      sources.length === 0
        ? groups
        : groups.filter((g) => g.allowed_sources?.some((s) => sources.includes(s)))
    // A query ranks by relevance; the sort control only applies without one.
    return needle
      ? fuzzyFilter(filtered, needle, (g) => [g.name, g.description])
      : sortGroups(filtered, sort)
  }, [groups, sources, needle, sort])

  // In the graph view the search box jumps the viewport instead of filtering,
  // so it offers suggestions rather than narrowing the canvas.
  const suggestions = view === "graph" && needle ? visible.slice(0, 8) : []

  useEffect(() => {
    const onPointerDown = (event: PointerEvent) => {
      if (!searchRef.current?.contains(event.target as globalThis.Node)) {
        setSuggestionsOpen(false)
      }
    }
    document.addEventListener("pointerdown", onPointerDown)
    return () => document.removeEventListener("pointerdown", onPointerDown)
  }, [])

  const setView = (next: ViewID) => {
    setSearchParams(
      (params) => {
        const updated = new URLSearchParams(params)
        if (next === "grid") updated.delete("view")
        else updated.set("view", next)
        if (next !== "graph") updated.delete("focus")
        return updated
      },
      { replace: true },
    )
  }

  const jumpTo = (groupID: string) => {
    graphRef.current?.centerOnGroup(groupID)
    setQuery("")
    setSuggestionsOpen(false)
  }

  const toggleSource = (source: GroupSource) =>
    setSources((current) =>
      current.includes(source)
        ? current.filter((s) => s !== source)
        : [...current, source],
    )

  const isGraph = view === "graph"
  const emptyLabel = needle
    ? `No groups match "${query}".`
    : sources.length > 0
      ? "No groups use the selected sources."
      : "No groups yet."

  return (
    <PageContainer
      className={cn(isGraph && "flex h-[calc(100svh-3.5rem)] max-w-none flex-col py-6")}
    >
      <div className="mb-6 flex items-start justify-between gap-4">
        <PageHeader
          title="Groups"
          description="Group memberships allow gating additional access across Gaucho Racing."
        />
        <Button asChild>
          <Link to="/groups/new">
            <Plus className="mr-1 size-3.5" />
            New group
          </Link>
        </Button>
      </div>

      <div className="mb-6 flex flex-wrap items-center gap-2">
        <div ref={searchRef} className="relative min-w-56 flex-1 sm:max-w-xs">
          <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            type="search"
            placeholder={isGraph ? "Jump to a group…" : "Search groups…"}
            value={query}
            onChange={(e) => {
              setQuery(e.target.value)
              setSuggestionsOpen(true)
            }}
            onFocus={() => setSuggestionsOpen(true)}
            onKeyDown={(e) => {
              if (e.key === "Escape") setSuggestionsOpen(false)
              if (e.key === "Enter" && isGraph && suggestions[0]) jumpTo(suggestions[0].id)
            }}
            className="h-8 pl-9"
          />
          {suggestionsOpen && suggestions.length > 0 && (
            <div className="absolute z-20 mt-1 w-full overflow-hidden rounded-lg border border-border/60 bg-popover shadow-lg">
              {suggestions.map((group) => (
                <button
                  key={group.id}
                  type="button"
                  onClick={() => jumpTo(group.id)}
                  className="flex w-full items-center justify-between gap-2 px-3 py-2 text-left text-sm transition-colors hover:bg-muted/60"
                >
                  <span className="truncate">{group.name}</span>
                  <span className="shrink-0 text-xs text-muted-foreground">
                    {group.member_count}
                  </span>
                </button>
              ))}
            </div>
          )}
        </div>

        {!isGraph && (
          <>
            <div className="flex items-center gap-1">
              {ALL_GROUP_SOURCES.map((source) => {
                const active = sources.includes(source)
                return (
                  <Button
                    key={source}
                    variant={active ? "secondary" : "outline"}
                    size="sm"
                    aria-pressed={active}
                    onClick={() => toggleSource(source)}
                    className={cn("font-mono text-xs", !active && "text-muted-foreground")}
                  >
                    {SOURCE_LABEL[source]}
                  </Button>
                )
              })}
            </div>

            <Select value={sort} onValueChange={(v) => setSort(v as SortID)}>
              <SelectTrigger size="sm" className="w-36" aria-label="Sort groups">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {SORTS.map((option) => (
                  <SelectItem key={option.id} value={option.id}>
                    {option.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
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

      {view === "grid" && (
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {groupsQuery.isLoading ? (
            Array.from({ length: 6 }).map((_, i) => (
              <Skeleton key={i} className="h-32 rounded-lg" />
            ))
          ) : visible.length === 0 ? (
            <p className="col-span-full py-12 text-center text-sm text-muted-foreground">
              {emptyLabel}
            </p>
          ) : (
            visible.map((group) => <GroupCard key={group.id} group={group} />)
          )}
        </div>
      )}

      {view === "apps" && (
        <GroupsByApplication
          groups={visible}
          applications={applicationsQuery.data ?? []}
          links={linksQuery.data ?? []}
          isLoading={groupsQuery.isLoading || applicationsQuery.isLoading || linksQuery.isLoading}
          emptyLabel={emptyLabel}
        />
      )}

      {isGraph && (
        <Suspense
          fallback={
            <div className="flex min-h-0 flex-1 overflow-hidden rounded-xl border border-border/60">
              <Skeleton className="size-full rounded-none" />
            </div>
          }
        >
          <GroupGraphView ref={graphRef} />
        </Suspense>
      )}
    </PageContainer>
  )
}
