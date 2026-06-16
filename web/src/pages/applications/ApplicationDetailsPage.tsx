import { useQuery } from "@tanstack/react-query"
import { ArrowLeft, Copy, ExternalLink, Eye, EyeOff, Pencil } from "lucide-react"
import { useState } from "react"
import { Link, useParams } from "react-router-dom"
import { toast } from "sonner"

import { OutlineButton } from "@/components/OutlineButton"
import { PageContainer } from "@/components/PageContainer"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { useAdmins } from "@/lib/admin"
import { api } from "@/lib/api"
import type { Application, GroupWithLink } from "@/lib/applications"
import { loadSession, type Entity } from "@/lib/auth"

import { ServiceAccountsCard } from "./ServiceAccountsCard"

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

function Field({
  label,
  children,
}: {
  label: string
  children: React.ReactNode
}) {
  return (
    <div>
      <p className="text-xs uppercase tracking-wider text-muted-foreground">{label}</p>
      <div className="mt-1.5">{children}</div>
    </div>
  )
}

function CopyableMono({
  value,
  label,
  className,
}: {
  value: string
  label: string
  className?: string
}) {
  return (
    <div className="flex items-center gap-1 rounded-md border border-border/60 bg-muted/40 px-2.5 py-1.5">
      <code className={`flex-1 truncate font-mono text-xs ${className ?? ""}`}>{value}</code>
      <Button
        variant="ghost"
        size="icon-sm"
        onClick={() => {
          navigator.clipboard.writeText(value)
          toast.success(`${label} copied`)
        }}
      >
        <Copy className="size-3.5" />
      </Button>
    </div>
  )
}

export default function ApplicationDetailsPage() {
  const { id } = useParams<{ id: string }>()
  const [secretVisible, setSecretVisible] = useState(false)

  const query = useQuery({
    queryKey: ["application", "id", id],
    queryFn: async () => {
      const res = await api.get<Application>(`/applications/${id}`)
      return res.data
    },
    enabled: !!id,
  })

  const secretQuery = useQuery({
    queryKey: ["application", "id", id, "secret"],
    queryFn: async () => {
      const res = await api.get<{ client_secret: string }>(`/applications/${id}/secret`)
      return res.data.client_secret
    },
    enabled: !!id && secretVisible,
    staleTime: 5 * 60 * 1000,
  })

  const linkedGroupsQuery = useQuery({
    queryKey: ["application", "id", id, "groups"],
    queryFn: async () => {
      const res = await api.get<GroupWithLink[]>(`/applications/${id}/groups`)
      return res.data
    },
    enabled: !!id,
  })

  const ownerId = query.data?.owner_id
  const myEntityID = loadSession()?.entityId
  const { isAdmin } = useAdmins()
  // Owners and admins can manage service accounts on this app. The
  // backend gates the same way (requireAppOwnerOrAdmin), so non-owners
  // would just see 403s on every action — hide the section entirely
  // rather than render a broken UI.
  const canManageServiceAccounts =
    !!ownerId && (ownerId === myEntityID || isAdmin)
  const ownerQuery = useQuery({
    queryKey: ["entity", ownerId],
    queryFn: async () => {
      const res = await api.get<Entity>(`/entities/${ownerId}`)
      return res.data
    },
    enabled: !!ownerId,
    staleTime: 5 * 60 * 1000,
  })

  function renderOwner() {
    if (!ownerId) return <span className="text-sm text-muted-foreground">—</span>
    if (ownerQuery.isLoading) {
      return <span className="text-sm text-muted-foreground">Loading…</span>
    }
    const e = ownerQuery.data
    if (!e) return <span className="font-mono text-xs text-muted-foreground">{ownerId}</span>
    if (e.type === "USER" && e.user) {
      const full = `${e.user.first_name} ${e.user.last_name}`.trim()
      const name = full || e.user.username || ownerId
      const initials = (e.user.first_name?.[0] ?? "") + (e.user.last_name?.[0] ?? "")
      return (
        <div className="flex items-center gap-2.5">
          <Avatar className="size-8">
            <AvatarImage src={e.user.avatar_url} alt={name} />
            <AvatarFallback className="text-xs">
              {initials.toUpperCase() || name.slice(0, 1).toUpperCase()}
            </AvatarFallback>
          </Avatar>
          <div className="flex flex-col leading-tight">
            <span className="text-sm">{name}</span>
            {e.user.username && (
              <span className="text-xs text-muted-foreground">@{e.user.username}</span>
            )}
          </div>
        </div>
      )
    }
    if (e.type === "SERVICE_ACCOUNT" && e.service_account) {
      const name = e.service_account.name || ownerId
      const initials = name
        .split(/\s+/)
        .map((p) => p[0])
        .filter(Boolean)
        .slice(0, 2)
        .join("")
        .toUpperCase()
      return (
        <div className="flex items-center gap-2.5">
          <Avatar className="size-8">
            <AvatarFallback className="bg-gradient-to-br from-gr-pink to-gr-purple text-xs text-white">
              {initials || name.slice(0, 1).toUpperCase()}
            </AvatarFallback>
          </Avatar>
          <div className="flex flex-col leading-tight">
            <span className="text-sm">{name}</span>
            <span className="text-xs text-muted-foreground">Service account</span>
          </div>
        </div>
      )
    }
    return <span className="font-mono text-xs text-muted-foreground">{ownerId}</span>
  }

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
  const maskedSecret = "•".repeat(48)

  return (
    <PageContainer>
      <Button asChild variant="ghost" size="sm" className="-ml-2 mb-4 text-muted-foreground">
        <Link to="/applications">
          <ArrowLeft className="mr-1 size-3.5" />
          All applications
        </Link>
      </Button>

      <header className="mb-8 flex flex-wrap items-start justify-between gap-4">
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
        <div className="flex gap-2">
          <Button
            asChild
            variant="outline"
            className="h-10 gap-1.5 rounded-xl px-4 text-sm"
          >
            <Link to={`/applications/${app.id}/edit`}>
              <Pencil className="size-3.5" />
              Edit
            </Link>
          </Button>
          {app.launch_url && (
            <OutlineButton
              type="button"
              className="w-auto"
              onClick={() => window.open(app.launch_url, "_blank", "noreferrer")}
            >
              Launch
              <ExternalLink className="size-3.5" />
            </OutlineButton>
          )}
        </div>
      </header>

      <div className="space-y-4">
        <Card>
          <CardHeader>
            <CardTitle>OAuth credentials</CardTitle>
          </CardHeader>
          <CardContent className="space-y-5">
            <Field label="Client ID">
              <CopyableMono value={app.client_id} label="Client ID" />
            </Field>
            <Field label="Client Secret">
              <div className="flex items-center gap-1 rounded-md border border-border/60 bg-muted/40 px-2.5 py-1.5">
                <code className="flex-1 truncate font-mono text-xs">
                  {secretVisible ? (secretQuery.data ?? "Loading…") : maskedSecret}
                </code>
                <Button
                  variant="ghost"
                  size="icon-sm"
                  onClick={() => setSecretVisible((v) => !v)}
                >
                  {secretVisible ? <EyeOff className="size-3.5" /> : <Eye className="size-3.5" />}
                </Button>
                <Button
                  variant="ghost"
                  size="icon-sm"
                  disabled={!secretQuery.data}
                  onClick={() => {
                    if (!secretQuery.data) return
                    navigator.clipboard.writeText(secretQuery.data)
                    toast.success("Client secret copied")
                  }}
                >
                  <Copy className="size-3.5" />
                </Button>
              </div>
            </Field>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Redirect URIs</CardTitle>
          </CardHeader>
          <CardContent>
            {app.redirect_uris.length === 0 ? (
              <p className="text-sm text-muted-foreground">
                No redirect URIs registered.{" "}
                <Link
                  to={`/applications/${app.id}/edit`}
                  className="text-foreground hover:text-gr-pink"
                >
                  Add one
                </Link>
                .
              </p>
            ) : (
              <ul className="space-y-2">
                {app.redirect_uris.map((uri) => (
                  <li key={uri}>
                    <CopyableMono value={uri} label="Redirect URI" />
                  </li>
                ))}
              </ul>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Linked groups</CardTitle>
          </CardHeader>
          <CardContent>
            {!linkedGroupsQuery.data ? (
              <Skeleton className="h-10 w-full" />
            ) : linkedGroupsQuery.data.length === 0 ? (
              <p className="text-sm text-muted-foreground">
                No groups linked.{" "}
                <Link
                  to={`/applications/${app.id}/edit`}
                  className="text-foreground hover:text-gr-pink"
                >
                  Link one
                </Link>
                .
              </p>
            ) : (
              <ul className="space-y-2">
                {linkedGroupsQuery.data.map((g) => (
                  <li
                    key={g.id}
                    className="flex items-center gap-2 rounded-md border border-border/60 bg-muted/40 px-2.5 py-1.5"
                  >
                    <Link
                      to={`/groups/${g.id}`}
                      className="flex-1 truncate text-sm hover:text-gr-pink"
                    >
                      {g.name}
                    </Link>
                    <Badge variant={g.required ? "default" : "outline"}>
                      {g.required ? "Required" : "Optional"}
                    </Badge>
                  </li>
                ))}
              </ul>
            )}
          </CardContent>
        </Card>

        {canManageServiceAccounts && app && (
          <ServiceAccountsCard applicationID={app.id} />
        )}

        <Card>
          <CardHeader>
            <CardTitle>Metadata</CardTitle>
          </CardHeader>
          <CardContent className="space-y-5">
            <Field label="Launch URL">
              {app.launch_url ? (
                <a
                  href={app.launch_url}
                  target="_blank"
                  rel="noreferrer"
                  className="break-all text-sm text-foreground hover:text-gr-pink"
                >
                  {app.launch_url}
                </a>
              ) : (
                <span className="text-sm text-muted-foreground">—</span>
              )}
            </Field>
            <Field label="Created by">{renderOwner()}</Field>
            <Field label="Registered">
              <span className="text-sm">{formatTime(app.created_at)}</span>
            </Field>
            <Field label="Last updated">
              <span className="text-sm">{formatTime(app.updated_at)}</span>
            </Field>
          </CardContent>
        </Card>
      </div>
    </PageContainer>
  )
}
