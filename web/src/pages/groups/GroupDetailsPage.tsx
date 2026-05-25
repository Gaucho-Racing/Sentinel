import { useQuery, useQueryClient } from "@tanstack/react-query"
import {
  ArrowLeft,
  Bot,
  Crown,
  Inbox,
  Pencil,
  Search,
  Sparkles,
  UserPlus,
} from "lucide-react"
import { useState } from "react"
import { Link, useParams } from "react-router-dom"
import { toast } from "sonner"

import { EntityChip } from "@/components/EntityChip"
import { PageContainer } from "@/components/PageContainer"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Skeleton } from "@/components/ui/skeleton"
import { api } from "@/lib/api"
import type { Application } from "@/lib/applications"
import {
  SOURCE_LABEL,
  type Group,
  type GroupJoinRequest,
  type GroupMember,
  type GroupOwner,
  type GroupSource,
} from "@/lib/groups"

const MEMBER_PREVIEW_COUNT = 6

function formatDate(iso: string) {
  if (!iso) return "—"
  return new Date(iso).toLocaleDateString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
  })
}

function relativeTime(iso: string) {
  const ms = Date.now() - new Date(iso).getTime()
  const minutes = Math.floor(ms / 60_000)
  if (minutes < 1) return "just now"
  if (minutes < 60) return `${minutes}m ago`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}h ago`
  const days = Math.floor(hours / 24)
  if (days < 30) return `${days}d ago`
  const months = Math.floor(days / 30)
  return `${months}mo ago`
}

function SourcePill({ source }: { source: GroupSource }) {
  return (
    <Badge variant="outline" className="font-mono text-[10px]">
      {SOURCE_LABEL[source]}
    </Badge>
  )
}

function MemberRow({ member }: { member: GroupMember }) {
  return (
    <div className="flex items-center justify-between gap-3 py-2.5">
      <EntityChip entityId={member.entity_id} />
      <div className="flex shrink-0 items-center gap-2">
        {member.source && <SourcePill source={member.source as GroupSource} />}
        <span className="hidden text-xs text-muted-foreground sm:inline">
          joined {formatDate(member.joined_at)}
        </span>
      </div>
    </div>
  )
}

function OwnerRow({ owner }: { owner: GroupOwner }) {
  return (
    <div className="flex items-center justify-between gap-3 py-2.5">
      <EntityChip entityId={owner.entity_id} />
      <span className="text-xs text-muted-foreground">
        since {formatDate(owner.created_at)}
      </span>
    </div>
  )
}

function SyncConfigBlock({ source }: { source: GroupSource }) {
  if (source === "DIRECT") {
    return (
      <div className="flex items-start gap-3 py-3">
        <UserPlus className="mt-0.5 size-4 shrink-0 text-muted-foreground" />
        <div className="min-w-0">
          <p className="text-sm font-medium">Direct invitation</p>
          <p className="mt-0.5 text-xs text-muted-foreground">
            Owners add members manually or approve join requests.
          </p>
        </div>
      </div>
    )
  }
  if (source === "DISCORD") {
    return (
      <div className="flex items-start gap-3 py-3">
        <Bot className="mt-0.5 size-4 shrink-0 text-muted-foreground" />
        <div className="min-w-0">
          <p className="text-sm font-medium">Discord role sync</p>
          <p className="mt-0.5 text-xs text-muted-foreground">
            Members with a linked Discord role are added automatically. Role binding not yet configurable.
          </p>
        </div>
      </div>
    )
  }
  return (
    <div className="flex items-start gap-3 py-3">
      <Sparkles className="mt-0.5 size-4 shrink-0 text-muted-foreground" />
      <div className="min-w-0">
        <p className="text-sm font-medium">Conditional rule</p>
        <p className="mt-0.5 text-xs text-muted-foreground">
          Members are auto-populated by a rule against entity profiles. Rule editor not yet built.
        </p>
      </div>
    </div>
  )
}

export default function GroupDetailsPage() {
  const { id } = useParams<{ id: string }>()
  const qc = useQueryClient()
  const [memberSearch, setMemberSearch] = useState("")
  const [reviewing, setReviewing] = useState<string | null>(null)

  const groupQuery = useQuery({
    queryKey: ["group", id],
    queryFn: async () => {
      const res = await api.get<Group>(`/groups/${id}`)
      return res.data
    },
    enabled: !!id,
  })

  const membersQuery = useQuery({
    queryKey: ["group", id, "members"],
    queryFn: async () => {
      const res = await api.get<GroupMember[]>(`/groups/${id}/members`)
      return res.data
    },
    enabled: !!id,
  })

  const ownersQuery = useQuery({
    queryKey: ["group", id, "owners"],
    queryFn: async () => {
      const res = await api.get<GroupOwner[]>(`/groups/${id}/owners`)
      return res.data
    },
    enabled: !!id,
  })

  const requestsQuery = useQuery({
    queryKey: ["group", id, "requests"],
    queryFn: async () => {
      const res = await api.get<GroupJoinRequest[]>(`/groups/${id}/requests`)
      return res.data
    },
    enabled: !!id,
  })

  const appsQuery = useQuery({
    queryKey: ["group", id, "applications"],
    queryFn: async () => {
      const res = await api.get<Application[]>(`/groups/${id}/applications`)
      return res.data
    },
    enabled: !!id,
  })

  async function reviewRequest(requestID: string, action: "approve" | "reject") {
    if (reviewing) return
    setReviewing(requestID)
    try {
      await api.post(`/groups/${id}/requests/${requestID}/${action}`, {
        reviewed_by: "",
      })
      qc.invalidateQueries({ queryKey: ["group", id, "requests"] })
      qc.invalidateQueries({ queryKey: ["group", id, "members"] })
      qc.invalidateQueries({ queryKey: ["group", id] })
      toast.success(`Request ${action === "approve" ? "approved" : "rejected"}`)
    } catch (err: unknown) {
      const message =
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        `Couldn't ${action} request.`
      toast.error(message)
    } finally {
      setReviewing(null)
    }
  }

  if (groupQuery.isLoading) {
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

  if (groupQuery.isError || !groupQuery.data) {
    return (
      <PageContainer>
        <Button asChild variant="ghost" size="sm" className="-ml-2 mb-4 text-muted-foreground">
          <Link to="/groups">
            <ArrowLeft className="mr-1 size-3.5" />
            All groups
          </Link>
        </Button>
        <p className="text-sm text-muted-foreground">Group not found.</p>
      </PageContainer>
    )
  }

  const group = groupQuery.data
  const members = membersQuery.data ?? []
  const owners = ownersQuery.data ?? []
  const requests = requestsQuery.data ?? []
  const apps = appsQuery.data ?? []
  const pending = requests.filter((r) => r.status === "PENDING")

  const needle = memberSearch.trim().toLowerCase()
  const searching = needle.length > 0
  const matchedMembers = searching
    ? members.filter((m) => m.entity_id.toLowerCase().includes(needle))
    : members
  const visibleMembers = searching
    ? matchedMembers
    : matchedMembers.slice(0, MEMBER_PREVIEW_COUNT)
  const remainingMembers = searching
    ? 0
    : Math.max(0, members.length - visibleMembers.length)

  const directCount = members.filter((m) => m.source === "DIRECT").length
  const syncedCount = members.filter((m) => m.source === "DISCORD" || m.source === "CONDITIONAL").length

  return (
    <PageContainer>
      <Button asChild variant="ghost" size="sm" className="-ml-2 mb-4 text-muted-foreground">
        <Link to="/groups">
          <ArrowLeft className="mr-1 size-3.5" />
          All groups
        </Link>
      </Button>

      <header className="mb-8 flex flex-wrap items-start justify-between gap-4">
        <div className="flex items-center gap-4">
          <div className="flex size-16 items-center justify-center overflow-hidden rounded-xl bg-gradient-to-br from-gr-pink to-gr-purple text-2xl font-semibold text-white">
            {group.name.slice(0, 1).toUpperCase()}
          </div>
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">{group.name}</h1>
            <p className="mt-1 max-w-prose text-sm text-muted-foreground">{group.description}</p>
            <div className="mt-2 flex flex-wrap items-center gap-2">
              {group.allowed_sources?.map((source) => (
                <SourcePill key={source} source={source} />
              ))}
            </div>
          </div>
        </div>
        <div className="flex gap-2">
          <Button asChild variant="outline" className="h-10 gap-1.5 rounded-xl px-4 text-sm">
            <Link to={`/groups/${group.id}/edit`}>
              <Pencil className="size-3.5" />
              Edit
            </Link>
          </Button>
        </div>
      </header>

      <div className="space-y-4">
        {pending.length > 0 && (
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Inbox className="size-4 text-gr-pink" />
                Pending join requests
                <Badge
                  variant="outline"
                  className="ml-1 border-gr-pink/40 bg-gr-pink/10 text-gr-pink"
                >
                  {pending.length}
                </Badge>
              </CardTitle>
              <CardDescription>
                Awaiting owner review. Approving creates a member with <code className="font-mono">source=DIRECT</code>.
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-3">
              {pending.map((req) => (
                <div
                  key={req.id}
                  className="rounded-lg border border-border/60 bg-muted/30 p-3"
                >
                  <div className="flex items-start justify-between gap-3">
                    <div className="min-w-0 flex-1">
                      <EntityChip entityId={req.entity_id} />
                      <p className="mt-2 text-xs text-muted-foreground">
                        Requested {relativeTime(req.created_at)}
                      </p>
                    </div>
                    <div className="flex gap-2">
                      <Button
                        variant="ghost"
                        size="sm"
                        disabled={reviewing === req.id}
                        onClick={() => reviewRequest(req.id, "reject")}
                      >
                        Reject
                      </Button>
                      <Button
                        size="sm"
                        disabled={reviewing === req.id}
                        onClick={() => reviewRequest(req.id, "approve")}
                      >
                        Approve
                      </Button>
                    </div>
                  </div>
                </div>
              ))}
            </CardContent>
          </Card>
        )}

        <Card>
          <CardHeader>
            <CardTitle>Group details</CardTitle>
            <CardDescription>
              How members are added, what this group unlocks, and identifiers.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
              <section className="lg:col-span-2 lg:border-r lg:border-border/60 lg:pr-6">
                <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
                  How members are added
                </p>
                <div className="mt-2 divide-y divide-border/60">
                  {group.allowed_sources && group.allowed_sources.length > 0 ? (
                    group.allowed_sources.map((source) => (
                      <SyncConfigBlock key={source} source={source} />
                    ))
                  ) : (
                    <p className="py-3 text-sm text-muted-foreground">
                      No sources configured yet.
                    </p>
                  )}
                </div>
              </section>

              <div className="space-y-6">
                <section>
                  <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
                    Linked applications
                  </p>
                  {apps.length === 0 ? (
                    <p className="mt-3 text-sm text-muted-foreground">No linked applications.</p>
                  ) : (
                    <ul className="mt-3 space-y-2">
                      {apps.map((app) => (
                        <li
                          key={app.id}
                          className="flex items-center gap-2.5 rounded-md border border-border/60 bg-muted/40 px-3 py-2"
                        >
                          <div className="flex size-7 shrink-0 items-center justify-center overflow-hidden rounded bg-gradient-to-br from-gr-pink to-gr-purple text-xs font-semibold text-white">
                            {app.icon_url ? (
                              <img
                                src={app.icon_url}
                                alt={app.name}
                                className="size-full object-cover"
                              />
                            ) : (
                              app.name.slice(0, 1).toUpperCase()
                            )}
                          </div>
                          <div className="min-w-0 flex-1 leading-tight">
                            <p className="truncate text-sm">{app.name}</p>
                            <p className="truncate font-mono text-xs text-muted-foreground">
                              {app.client_id}
                            </p>
                          </div>
                        </li>
                      ))}
                    </ul>
                  )}
                </section>

                <section className="border-t border-border/60 pt-6">
                  <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
                    Metadata
                  </p>
                  <div className="mt-3 space-y-3 text-sm">
                    <div className="flex items-center justify-between gap-3">
                      <span className="text-muted-foreground">Group ID</span>
                      <code className="truncate font-mono text-xs">{group.id}</code>
                    </div>
                    <div className="flex items-center justify-between gap-3">
                      <span className="text-muted-foreground">Created by</span>
                      {group.created_by ? (
                        <EntityChip entityId={group.created_by} size="sm" />
                      ) : (
                        <span className="text-sm text-muted-foreground">—</span>
                      )}
                    </div>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Created</span>
                      <span>{formatDate(group.created_at)}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Last updated</span>
                      <span>{formatDate(group.updated_at)}</span>
                    </div>
                  </div>
                </section>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Crown className="size-4 text-muted-foreground" />
              Owners
            </CardTitle>
            <CardDescription>
              Can edit the group, manage members, and approve requests.
            </CardDescription>
          </CardHeader>
          <CardContent>
            {ownersQuery.isLoading ? (
              <Skeleton className="h-10 w-full" />
            ) : owners.length === 0 ? (
              <p className="text-sm text-muted-foreground">No owners assigned.</p>
            ) : (
              <ul className="divide-y divide-border/60">
                {owners.map((o) => (
                  <li key={o.entity_id}>
                    <OwnerRow owner={o} />
                  </li>
                ))}
              </ul>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex-row items-start justify-between gap-4">
            <div>
              <CardTitle>Members</CardTitle>
              <CardDescription>
                {members.length} total · {directCount} direct · {syncedCount} synced
              </CardDescription>
            </div>
          </CardHeader>
          <CardContent>
            <div className="relative mb-4">
              <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                type="search"
                placeholder="Search members…"
                value={memberSearch}
                onChange={(e) => setMemberSearch(e.target.value)}
                className="pl-9"
              />
            </div>
            {membersQuery.isLoading ? (
              <Skeleton className="h-10 w-full" />
            ) : visibleMembers.length === 0 ? (
              <p className="py-6 text-center text-sm text-muted-foreground">
                {searching ? `No members match "${memberSearch}".` : "No members yet."}
              </p>
            ) : (
              <ul className="divide-y divide-border/60">
                {visibleMembers.map((m) => (
                  <li key={m.entity_id}>
                    <MemberRow member={m} />
                  </li>
                ))}
              </ul>
            )}
            {remainingMembers > 0 && (
              <p className="mt-3 text-xs text-muted-foreground">
                + {remainingMembers} more not shown.
              </p>
            )}
          </CardContent>
        </Card>
      </div>
    </PageContainer>
  )
}
