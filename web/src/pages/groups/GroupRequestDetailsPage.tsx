import { useQuery, useQueryClient } from "@tanstack/react-query"
import { ArrowLeft, CheckCircle2, Send, Trash2, UserPlus, XCircle } from "lucide-react"
import { useState } from "react"
import { Link, useNavigate, useParams } from "react-router-dom"
import { toast } from "sonner"

import { EntityChip } from "@/components/EntityChip"
import { PageContainer } from "@/components/PageContainer"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Skeleton } from "@/components/ui/skeleton"
import { Textarea } from "@/components/ui/textarea"
import { useAdmins } from "@/lib/admin"
import { api } from "@/lib/api"
import { loadSession } from "@/lib/auth"
import {
  formatAbsoluteDate,
  formatDurationBetween,
  formatExpiresIn,
} from "@/lib/duration"
import type { Group, GroupJoinRequest, GroupJoinRequestStatus, GroupOwner } from "@/lib/groups"

import { ReviewRequestDialog } from "./ReviewRequestDialog"

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

function formatTimestamp(iso: string) {
  if (!iso) return ""
  return new Date(iso).toLocaleString(undefined, {
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  })
}

type TimelineEntry =
  | {
      kind: "system"
      key: string
      icon: typeof UserPlus
      iconClass: string
      entityID: string
      text: string
      time: string
    }
  | {
      kind: "comment"
      key: string
      entityID: string
      body: string
      time: string
    }

export default function GroupRequestDetailsPage() {
  const { id, requestID } = useParams<{ id: string; requestID: string }>()
  const navigate = useNavigate()
  const qc = useQueryClient()
  const myEntityID = loadSession()?.entityId ?? ""
  const { isAdmin } = useAdmins()

  const [composing, setComposing] = useState("")
  const [posting, setPosting] = useState(false)
  const [reviewAction, setReviewAction] = useState<"approve" | "reject" | null>(null)
  const [cancelling, setCancelling] = useState(false)
  const [cancelConfirmOpen, setCancelConfirmOpen] = useState(false)

  const groupQuery = useQuery({
    queryKey: ["group", id],
    queryFn: async () => {
      const res = await api.get<Group>(`/groups/${id}`)
      return res.data
    },
    enabled: !!id,
  })

  const requestQuery = useQuery({
    queryKey: ["group", id, "requests", requestID],
    queryFn: async () => {
      const res = await api.get<GroupJoinRequest>(`/groups/${id}/requests/${requestID}`)
      return res.data
    },
    enabled: !!id && !!requestID,
  })

  const ownersQuery = useQuery({
    queryKey: ["group", id, "owners"],
    queryFn: async () => {
      const res = await api.get<GroupOwner[]>(`/groups/${id}/owners`)
      return res.data
    },
    enabled: !!id,
  })

  async function postComment() {
    const body = composing.trim()
    if (!body || !id || !requestID || posting) return
    setPosting(true)
    try {
      await api.post(`/groups/${id}/requests/${requestID}/comments`, {
        entity_id: myEntityID,
        comment: body,
      })
      qc.invalidateQueries({ queryKey: ["group", id, "requests", requestID] })
      qc.invalidateQueries({ queryKey: ["group", id, "requests"] })
      setComposing("")
    } catch (err: unknown) {
      const message =
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        "Couldn't post comment."
      toast.error(message)
    } finally {
      setPosting(false)
    }
  }

  async function cancelRequest() {
    if (!id || !requestID || cancelling) return
    setCancelling(true)
    try {
      await api.delete(`/groups/${id}/requests/${requestID}`)
      qc.invalidateQueries({ queryKey: ["group", id, "requests"] })
      qc.invalidateQueries({ queryKey: ["group", id] })
      toast.success("Request cancelled")
      navigate(`/groups/${id}`)
    } catch (err: unknown) {
      const message =
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        "Couldn't cancel request."
      toast.error(message)
      setCancelling(false)
    }
  }

  if (requestQuery.isLoading || groupQuery.isLoading) {
    return (
      <PageContainer>
        <Skeleton className="mb-4 h-4 w-24" />
        <Skeleton className="mb-8 h-12 w-64" />
        <Skeleton className="h-64" />
      </PageContainer>
    )
  }

  if (requestQuery.isError || !requestQuery.data || !groupQuery.data) {
    return (
      <PageContainer>
        <Button asChild variant="ghost" size="sm" className="-ml-2 mb-4 text-muted-foreground">
          <Link to={`/groups/${id}/requests`}>
            <ArrowLeft className="mr-1 size-3.5" />
            All requests
          </Link>
        </Button>
        <p className="text-sm text-muted-foreground">Request not found.</p>
      </PageContainer>
    )
  }

  const group = groupQuery.data
  const request = requestQuery.data
  const owners = ownersQuery.data ?? []
  const isOwner = (!!myEntityID && owners.some((o) => o.entity_id === myEntityID)) || isAdmin
  const isRequester = !!myEntityID && request.entity_id === myEntityID
  const isPending = request.status === "PENDING"
  const canComment = isPending && (isOwner || isRequester)
  const status = request.status as GroupJoinRequestStatus

  // Build the timeline: synthesized creation entry, comments in order, optional review entry.
  const timeline: TimelineEntry[] = []
  const submittedText = request.has_expiration
    ? `submitted this request for ${formatDurationBetween(request.created_at, request.expires_at)} of access`
    : "submitted this request"
  timeline.push({
    kind: "system",
    key: "created",
    icon: UserPlus,
    iconClass: "text-muted-foreground",
    entityID: request.entity_id,
    text: submittedText,
    time: request.created_at,
  })
  const sortedComments = [...(request.comments ?? [])].sort(
    (a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime(),
  )
  for (const c of sortedComments) {
    timeline.push({
      kind: "comment",
      key: c.id,
      entityID: c.entity_id,
      body: c.comment,
      time: c.created_at,
    })
  }
  if (!isPending && request.reviewed_by) {
    timeline.push({
      kind: "system",
      key: "reviewed",
      icon: request.status === "APPROVED" ? CheckCircle2 : XCircle,
      iconClass: request.status === "APPROVED" ? "text-green-500" : "text-destructive",
      entityID: request.reviewed_by,
      text: request.status === "APPROVED" ? "approved this request" : "rejected this request",
      time: request.reviewed_at,
    })
  }

  return (
    <PageContainer>
      <Button asChild variant="ghost" size="sm" className="-ml-2 mb-4 text-muted-foreground">
        <Link to={`/groups/${id}/requests`}>
          <ArrowLeft className="mr-1 size-3.5" />
          All requests
        </Link>
      </Button>

      <header className="mb-8 flex flex-wrap items-start justify-between gap-4">
        <div className="min-w-0">
          <div className="flex items-center gap-3">
            <h1 className="text-2xl font-semibold tracking-tight">Join request</h1>
            <Badge variant="outline" className={STATUS_BADGE[status]}>
              {status.toLowerCase()}
            </Badge>
          </div>
          <p className="mt-1 text-sm text-muted-foreground">
            For <Link to={`/groups/${id}`} className="text-foreground hover:text-gr-pink">{group.name}</Link>.
          </p>
          {request.has_expiration && (
            <p className="mt-1 text-sm text-muted-foreground">
              {isPending ? "Would expire" : "Expires"}{" "}
              <span className="text-foreground">{formatAbsoluteDate(request.expires_at)}</span>{" "}
              ({formatExpiresIn(request.expires_at)})
            </p>
          )}
        </div>
        {isPending && (
          <div className="flex gap-2">
            {isOwner && (
              <>
                <Button variant="ghost" onClick={() => setReviewAction("reject")}>
                  Reject
                </Button>
                <Button onClick={() => setReviewAction("approve")}>Approve</Button>
              </>
            )}
            {isRequester && !isOwner && (
              <Button
                variant="destructive"
                disabled={cancelling}
                onClick={() => setCancelConfirmOpen(true)}
              >
                Cancel request
              </Button>
            )}
          </div>
        )}
      </header>

      <Card>
        <CardContent className="space-y-5">
          <ol className="space-y-5">
            {timeline.map((entry) => (
              <li key={entry.key}>
                {entry.kind === "system" ? (
                  <div className="flex items-center gap-2 text-sm text-muted-foreground">
                    <entry.icon className={"size-4 shrink-0 " + entry.iconClass} />
                    <EntityChip entityId={entry.entityID} size="sm" />
                    <span>{entry.text}</span>
                    <span className="ml-auto whitespace-nowrap text-xs" title={formatTimestamp(entry.time)}>
                      {relativeTime(entry.time)}
                    </span>
                  </div>
                ) : (
                  <div className="flex items-start gap-3">
                    <div className="pt-1">
                      <EntityChip entityId={entry.entityID} size="sm" />
                    </div>
                    <div className="flex-1 rounded-md border border-border/60 bg-muted/30 p-3">
                      <div className="flex items-baseline justify-between gap-3">
                        <span className="text-xs text-muted-foreground" title={formatTimestamp(entry.time)}>
                          {relativeTime(entry.time)}
                        </span>
                      </div>
                      <p className="mt-1 whitespace-pre-wrap text-sm">{entry.body}</p>
                    </div>
                  </div>
                )}
              </li>
            ))}
          </ol>

          {canComment && (
            <form
              onSubmit={(e) => {
                e.preventDefault()
                postComment()
              }}
              className="space-y-2 border-t border-border/60 pt-5"
            >
              <Textarea
                placeholder="Add a comment…"
                value={composing}
                onChange={(e) => setComposing(e.target.value)}
                rows={3}
              />
              <div className="flex justify-end">
                <Button type="submit" disabled={!composing.trim() || posting}>
                  <Send className="mr-1 size-3.5" />
                  {posting ? "Posting…" : "Post comment"}
                </Button>
              </div>
            </form>
          )}

          {!canComment && !isPending && (
            <p className="border-t border-border/60 pt-5 text-xs text-muted-foreground">
              Commenting closed because this request has been {status.toLowerCase()}.
            </p>
          )}
        </CardContent>
      </Card>

      <Dialog
        open={cancelConfirmOpen}
        onOpenChange={(open) => {
          if (!cancelling) setCancelConfirmOpen(open)
        }}
      >
        <DialogContent className="gap-5 sm:max-w-md">
          <DialogHeader className="gap-3">
            <div className="flex size-10 items-center justify-center rounded-xl bg-destructive/10 text-destructive">
              <Trash2 className="size-5" />
            </div>
            <DialogTitle>Cancel your request?</DialogTitle>
            <DialogDescription>
              This permanently removes your request to join {group.name}, along with any
              comments on it. You can submit a new one later.
            </DialogDescription>
          </DialogHeader>

          <div className="flex justify-end gap-2 pt-1">
            <Button
              type="button"
              variant="ghost"
              disabled={cancelling}
              onClick={() => setCancelConfirmOpen(false)}
            >
              Keep request
            </Button>
            <Button
              type="button"
              variant="destructive"
              disabled={cancelling}
              onClick={cancelRequest}
            >
              {cancelling ? "Cancelling…" : "Cancel request"}
            </Button>
          </div>
        </DialogContent>
      </Dialog>

      <ReviewRequestDialog
        open={reviewAction !== null}
        onOpenChange={(o) => {
          if (!o) setReviewAction(null)
        }}
        groupID={id ?? ""}
        request={request}
        action={reviewAction ?? "approve"}
        reviewerEntityID={myEntityID}
      />
    </PageContainer>
  )
}
