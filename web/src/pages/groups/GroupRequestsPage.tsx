import { useQuery } from "@tanstack/react-query"
import { ArrowLeft, ChevronRight } from "lucide-react"
import { useState } from "react"
import { Link, useParams } from "react-router-dom"

import { EntityChip } from "@/components/EntityChip"
import { PageContainer } from "@/components/PageContainer"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { api } from "@/lib/api"
import {
  formatAbsoluteDate,
  formatDurationBetween,
  formatExpiresIn,
} from "@/lib/duration"
import type { Group, GroupJoinRequest, GroupJoinRequestStatus } from "@/lib/groups"

type Filter = "ALL" | GroupJoinRequestStatus

const FILTERS: Filter[] = ["PENDING", "APPROVED", "REJECTED", "ALL"]
const FILTER_LABEL: Record<Filter, string> = {
  ALL: "All",
  PENDING: "Pending",
  APPROVED: "Approved",
  REJECTED: "Rejected",
}

const STATUS_BADGE: Record<GroupJoinRequestStatus, string> = {
  PENDING: "border-gr-pink/40 bg-gr-pink/10 text-gr-pink",
  APPROVED: "border-green-500/40 bg-green-500/10 text-green-500",
  REJECTED: "border-destructive/40 bg-destructive/10 text-destructive",
}

function relativeTime(iso: string) {
  if (!iso) return ""
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

export default function GroupRequestsPage() {
  const { id } = useParams<{ id: string }>()
  const [filter, setFilter] = useState<Filter>("PENDING")

  const groupQuery = useQuery({
    queryKey: ["group", id],
    queryFn: async () => {
      const res = await api.get<Group>(`/groups/${id}`)
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

  if (groupQuery.isLoading || requestsQuery.isLoading) {
    return (
      <PageContainer>
        <Skeleton className="mb-4 h-4 w-24" />
        <Skeleton className="mb-8 h-8 w-64" />
        <Skeleton className="h-48" />
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
  const requests = requestsQuery.data ?? []

  const counts: Record<Filter, number> = {
    ALL: requests.length,
    PENDING: requests.filter((r) => r.status === "PENDING").length,
    APPROVED: requests.filter((r) => r.status === "APPROVED").length,
    REJECTED: requests.filter((r) => r.status === "REJECTED").length,
  }

  const filtered =
    filter === "ALL" ? requests : requests.filter((r) => r.status === filter)
  const sorted = [...filtered].sort(
    (a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime(),
  )

  return (
    <PageContainer>
      <Button asChild variant="ghost" size="sm" className="-ml-2 mb-4 text-muted-foreground">
        <Link to={`/groups/${id}`}>
          <ArrowLeft className="mr-1 size-3.5" />
          Back to {group.name}
        </Link>
      </Button>

      <div className="mb-8">
        <h1 className="text-2xl font-semibold tracking-tight">Join requests</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          All requests to join {group.name}.
        </p>
      </div>

      <div className="mb-4 flex flex-wrap gap-2">
        {FILTERS.map((f) => {
          const active = filter === f
          return (
            <button
              key={f}
              type="button"
              onClick={() => setFilter(f)}
              className={
                "flex items-center gap-1.5 rounded-full border px-3 py-1 text-xs transition-colors " +
                (active
                  ? "border-foreground/30 bg-foreground/10 text-foreground"
                  : "border-border/60 bg-muted/30 text-muted-foreground hover:bg-muted/50")
              }
            >
              {FILTER_LABEL[f]}
              <span className="font-mono">{counts[f]}</span>
            </button>
          )
        })}
      </div>

      {sorted.length === 0 ? (
        <Card>
          <CardContent className="py-12 text-center text-sm text-muted-foreground">
            No {filter === "ALL" ? "" : FILTER_LABEL[filter].toLowerCase() + " "}requests.
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-2">
          {sorted.map((req) => {
            const reason = req.comments?.find((c) => c.entity_id === req.entity_id)?.comment
            const status = req.status as GroupJoinRequestStatus
            return (
              <Link
                key={req.id}
                to={`/groups/${id}/requests/${req.id}`}
                className="flex items-start justify-between gap-4 rounded-lg border border-border/60 bg-card p-4 transition-colors hover:bg-muted/40"
              >
                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-3">
                    <EntityChip entityId={req.entity_id} />
                    <Badge variant="outline" className={STATUS_BADGE[status]}>
                      {status.toLowerCase()}
                    </Badge>
                  </div>
                  <p className="mt-2 text-xs text-muted-foreground">
                    {req.status === "PENDING"
                      ? `Requested ${relativeTime(req.created_at)}`
                      : `Reviewed ${relativeTime(req.reviewed_at)}`}
                    {req.has_expiration && (
                      <span> · {formatDurationBetween(req.created_at, req.expires_at)} of access</span>
                    )}
                    {req.comments && req.comments.length > 0 && (
                      <span> · {req.comments.length} {req.comments.length === 1 ? "comment" : "comments"}</span>
                    )}
                  </p>
                  {reason && (
                    <p className="mt-2 line-clamp-2 text-sm text-muted-foreground">{reason}</p>
                  )}
                </div>
                {req.has_expiration && status !== "REJECTED" && (
                  <div className="shrink-0 text-right">
                    <p className="text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
                      {status === "PENDING" ? "Would expire" : "Expires"}
                    </p>
                    <p className="mt-0.5 text-sm font-semibold">
                      {formatAbsoluteDate(req.expires_at)}
                    </p>
                    <p className="text-xs text-muted-foreground">
                      {formatExpiresIn(req.expires_at)}
                    </p>
                  </div>
                )}
                <ChevronRight className="mt-1 size-4 shrink-0 text-muted-foreground" />
              </Link>
            )
          })}
        </div>
      )}
    </PageContainer>
  )
}
