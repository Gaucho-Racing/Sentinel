import { Plus, Search } from "lucide-react"
import { useState } from "react"

import { GroupCard } from "@/components/GroupCard"
import { PageContainer, PageHeader } from "@/components/PageContainer"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { MOCK_GROUPS } from "@/lib/groups"

export default function GroupsPage() {
  const [query, setQuery] = useState("")

  const needle = query.trim().toLowerCase()
  const filtered = needle
    ? MOCK_GROUPS.filter(
        (g) =>
          g.name.toLowerCase().includes(needle) ||
          g.description.toLowerCase().includes(needle),
      )
    : MOCK_GROUPS
  const sorted = [...filtered].sort((a, b) => a.name.localeCompare(b.name))

  return (
    <PageContainer>
      <div className="mb-6 flex items-start justify-between gap-4">
        <PageHeader
          title="Groups"
          description="Group memberships allow gating additional access across Gaucho Racing."
        />
        <Button>
          <Plus className="mr-1 size-3.5" />
          New group
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
        {sorted.length === 0 ? (
          <p className="col-span-full py-12 text-center text-sm text-muted-foreground">
            No groups match "{query}".
          </p>
        ) : (
          sorted.map((group) => <GroupCard key={group.id} group={group} />)
        )}
      </div>
    </PageContainer>
  )
}
