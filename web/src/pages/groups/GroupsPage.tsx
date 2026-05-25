import { useQuery } from "@tanstack/react-query"
import { Plus, Search } from "lucide-react"
import { useState } from "react"
import { Link } from "react-router-dom"

import { GroupCard } from "@/components/GroupCard"
import { PageContainer, PageHeader } from "@/components/PageContainer"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Skeleton } from "@/components/ui/skeleton"
import { api } from "@/lib/api"
import type { Group } from "@/lib/groups"

export default function GroupsPage() {
  const [query, setQuery] = useState("")

  const groupsQuery = useQuery({
    queryKey: ["groups"],
    queryFn: async () => {
      const res = await api.get<Group[]>("/groups")
      return res.data
    },
  })

  const groups = groupsQuery.data ?? []
  const needle = query.trim().toLowerCase()
  const filtered = needle
    ? groups.filter(
        (g) =>
          g.name.toLowerCase().includes(needle) ||
          g.description.toLowerCase().includes(needle),
      )
    : groups
  const sorted = [...filtered].sort((a, b) => a.name.localeCompare(b.name))

  return (
    <PageContainer>
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

      <div className="relative mb-6">
        <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          type="search"
          placeholder="Search groups…"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          className="pl-9"
        />
      </div>

      <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
        {groupsQuery.isLoading ? (
          Array.from({ length: 6 }).map((_, i) => (
            <Skeleton key={i} className="h-32 rounded-lg" />
          ))
        ) : sorted.length === 0 ? (
          <p className="col-span-full py-12 text-center text-sm text-muted-foreground">
            {needle ? `No groups match "${query}".` : "No groups yet."}
          </p>
        ) : (
          sorted.map((group) => <GroupCard key={group.id} group={group} />)
        )}
      </div>
    </PageContainer>
  )
}
