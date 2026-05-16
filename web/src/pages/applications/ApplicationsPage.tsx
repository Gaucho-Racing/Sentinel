import { useQuery } from "@tanstack/react-query"
import { Search } from "lucide-react"
import { useState } from "react"

import { AppCard } from "@/components/AppCard"
import { PageContainer, PageHeader } from "@/components/PageContainer"
import { Input } from "@/components/ui/input"
import { Skeleton } from "@/components/ui/skeleton"
import { api } from "@/lib/api"
import type { Application } from "@/lib/applications"

export default function ApplicationsPage() {
  const [query, setQuery] = useState("")

  const appsQuery = useQuery({
    queryKey: ["applications"],
    queryFn: async () => {
      const res = await api.get<Application[]>("/applications")
      return res.data
    },
  })

  const apps = appsQuery.data ?? []
  const needle = query.trim().toLowerCase()
  const filtered = needle
    ? apps.filter(
        (a) =>
          a.name.toLowerCase().includes(needle) ||
          a.description.toLowerCase().includes(needle) ||
          a.client_id.toLowerCase().includes(needle),
      )
    : apps
  const sorted = [...filtered].sort((a, b) => a.name.localeCompare(b.name))

  return (
    <PageContainer>
      <PageHeader
        title="Applications"
        description="Team apps you can sign into through Sentinel."
      />

      <div className="relative mb-6">
        <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          type="search"
          placeholder="Search applications…"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          className="pl-9"
        />
      </div>

      <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
        {appsQuery.isLoading ? (
          Array.from({ length: 6 }).map((_, i) => <Skeleton key={i} className="h-32 rounded-lg" />)
        ) : sorted.length === 0 ? (
          <p className="col-span-full py-12 text-center text-sm text-muted-foreground">
            {needle ? `No applications match "${query}".` : "No applications registered yet."}
          </p>
        ) : (
          sorted.map((app) => <AppCard key={app.id} app={app} />)
        )}
      </div>
    </PageContainer>
  )
}
