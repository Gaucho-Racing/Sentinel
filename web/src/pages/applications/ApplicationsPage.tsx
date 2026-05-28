import { useQuery } from "@tanstack/react-query"
import { Plus, Search } from "lucide-react"
import { useState } from "react"
import { Link } from "react-router-dom"

import { AppCard } from "@/components/AppCard"
import { PageContainer, PageHeader } from "@/components/PageContainer"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Skeleton } from "@/components/ui/skeleton"
import { api } from "@/lib/api"
import type { Application } from "@/lib/applications"
import { fuzzyFilter } from "@/lib/fuzzy"

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
  const needle = query.trim()
  const sorted = needle
    ? fuzzyFilter(apps, needle, (a) => [a.name, a.description, a.client_id])
    : [...apps].sort((a, b) => a.name.localeCompare(b.name))

  return (
    <PageContainer>
      <div className="mb-6 flex items-start justify-between gap-4">
        <PageHeader
          title="Applications"
          description="Team apps you can sign into through Sentinel."
        />
        <Button asChild>
          <Link to="/applications/new">
            <Plus className="mr-1 size-3.5" />
            New application
          </Link>
        </Button>
      </div>

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
