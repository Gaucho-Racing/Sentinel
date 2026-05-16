import { useQuery } from "@tanstack/react-query"
import { ArrowLeft, ExternalLink } from "lucide-react"
import { Link, useParams } from "react-router-dom"

import { OutlineButton } from "@/components/OutlineButton"
import { PageContainer } from "@/components/PageContainer"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { api } from "@/lib/api"
import type { Application } from "@/lib/applications"

function initial(name: string) {
  return name.slice(0, 1).toUpperCase()
}

function formatTime(iso: string) {
  return new Date(iso).toLocaleString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
  })
}

function Row({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="flex flex-col gap-1 border-b border-border/60 py-3 last:border-b-0 sm:flex-row sm:items-center sm:gap-6">
      <span className="text-xs uppercase tracking-wider text-muted-foreground sm:w-40 sm:shrink-0">
        {label}
      </span>
      <div className="min-w-0 flex-1 text-sm">{children}</div>
    </div>
  )
}

export default function ApplicationDetailsPage() {
  const { id } = useParams<{ id: string }>()
  const query = useQuery({
    queryKey: ["application", "id", id],
    queryFn: async () => {
      const res = await api.get<Application>(`/applications/${id}`)
      return res.data
    },
    enabled: !!id,
  })

  if (query.isLoading) {
    return (
      <PageContainer>
        <Skeleton className="mb-4 h-4 w-24" />
        <div className="mb-8 flex items-center gap-4">
          <Skeleton className="size-16 rounded-xl" />
          <div className="space-y-2">
            <Skeleton className="h-6 w-48" />
            <Skeleton className="h-4 w-64" />
          </div>
        </div>
        <Skeleton className="h-64" />
      </PageContainer>
    )
  }

  if (query.isError || !query.data) {
    return (
      <PageContainer>
        <Button asChild variant="ghost" size="sm" className="-ml-2 mb-4 text-muted-foreground">
          <Link to="/applications">
            <ArrowLeft className="mr-1 size-3.5" />
            All applications
          </Link>
        </Button>
        <p className="text-sm text-muted-foreground">Application not found.</p>
      </PageContainer>
    )
  }

  const app = query.data

  return (
    <PageContainer>
      <Button asChild variant="ghost" size="sm" className="-ml-2 mb-4 text-muted-foreground">
        <Link to="/applications">
          <ArrowLeft className="mr-1 size-3.5" />
          All applications
        </Link>
      </Button>

      <header className="mb-8 flex items-start justify-between gap-4">
        <div className="flex items-center gap-4">
          <div className="flex size-16 items-center justify-center overflow-hidden rounded-xl bg-gradient-to-br from-gr-pink to-gr-purple text-2xl font-semibold text-white">
            {app.icon_url ? (
              <img src={app.icon_url} alt={app.name} className="size-full object-cover" />
            ) : (
              initial(app.name)
            )}
          </div>
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">{app.name}</h1>
            <p className="mt-1 text-sm text-muted-foreground">{app.description}</p>
          </div>
        </div>
        {app.launch_url && (
          <OutlineButton
            type="button"
            className="w-auto"
            onClick={() => window.open(app.launch_url, "_blank", "noreferrer")}
          >
            Launch
            <ExternalLink className="ml-1.5 size-3.5" />
          </OutlineButton>
        )}
      </header>

      <section className="rounded-lg border border-border/60 bg-card px-5">
        <Row label="Client ID">
          <code className="font-mono text-xs">{app.client_id}</code>
        </Row>
        <Row label="Launch URL">
          {app.launch_url ? (
            <a
              href={app.launch_url}
              target="_blank"
              rel="noreferrer"
              className="break-all text-foreground hover:text-gr-pink"
            >
              {app.launch_url}
            </a>
          ) : (
            <span className="text-muted-foreground">—</span>
          )}
        </Row>
        <Row label="Redirect URIs">
          {app.redirect_uris.length === 0 ? (
            <span className="text-muted-foreground">—</span>
          ) : (
            <ul className="space-y-1">
              {app.redirect_uris.map((uri) => (
                <li key={uri} className="break-all font-mono text-xs text-muted-foreground">
                  {uri}
                </li>
              ))}
            </ul>
          )}
        </Row>
        <Row label="Registered">{formatTime(app.created_at)}</Row>
      </section>
    </PageContainer>
  )
}
