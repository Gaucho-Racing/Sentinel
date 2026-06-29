import { useQuery, useQueryClient } from "@tanstack/react-query"
import {
  ArrowLeft,
  Bot,
  Check,
  Crown,
  Hourglass,
  Inbox,
  Mail,
  Pencil,
  Search,
  Shield,
  Sparkles,
  Trash2,
  UserPlus,
} from "lucide-react"
import { useMemo, useState } from "react"
import { Link, useParams } from "react-router-dom"
import { toast } from "sonner"

import { EntityChip } from "@/components/EntityChip"
import { PageContainer } from "@/components/PageContainer"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardAction, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Skeleton } from "@/components/ui/skeleton"
import { Textarea } from "@/components/ui/textarea"
import { useAdmins } from "@/lib/admin"
import { api } from "@/lib/api"
import type { Application } from "@/lib/applications"
import { loadSession } from "@/lib/auth"
import {
  useGroupConditionalBindings,
  type GroupConditionalBinding,
} from "@/lib/conditional"
import {
  discordRoleColorHex,
  useDiscordRoles,
  useGroupDiscordBindings,
  type DiscordRole,
  type GroupDiscordRoleBinding,
} from "@/lib/discord"
import {
  addCustom,
  addPreset,
  DURATION_PRESETS,
  formatAbsoluteDate,
  formatDurationBetween,
  formatExpiresIn,
  MAX_BY_UNIT,
  type DurationPreset,
  type DurationUnit,
} from "@/lib/duration"
import { fuzzyFilter } from "@/lib/fuzzy"
import { useGroupGoogleBinding } from "@/lib/google"
import {
  SOURCE_LABEL,
  type Group,
  type GroupJoinRequest,
  type GroupMember,
  type GroupOwner,
  type GroupSource,
} from "@/lib/groups"

import { AddGroupPersonDialog } from "./AddGroupPersonDialog"
import { ReviewRequestDialog } from "./ReviewRequestDialog"

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

// h-10 OutlineButton-style display pill: colored ring + bg-card interior +
// colored text. Brand tone uses the gradient; gold/green use solid accents.
function OutlinePill({
  icon: Icon,
  label,
  tone = "brand",
}: {
  icon: React.ComponentType<{ className?: string }>
  label: string
  tone?: "brand" | "gold" | "green"
}) {
  const ring = {
    brand: "bg-gradient-to-r from-gr-pink to-gr-purple",
    gold: "bg-amber-400",
    green: "bg-emerald-500",
  }[tone]
  const accent = {
    brand: "text-gr-pink",
    gold: "text-amber-400",
    green: "text-emerald-500",
  }[tone]
  return (
    <span className={`inline-flex h-10 items-center rounded-xl ${ring} p-0.5 text-sm font-medium`}>
      <span className="inline-flex h-full items-center gap-1.5 rounded-[10px] bg-card px-4">
        <Icon className={`size-3.5 ${accent}`} />
        {tone === "brand" ? (
          <span className="bg-gradient-to-r from-gr-pink to-gr-purple bg-clip-text text-transparent">
            {label}
          </span>
        ) : (
          <span className={accent}>{label}</span>
        )}
      </span>
    </span>
  )
}

function MemberRow({ member }: { member: GroupMember }) {
  return (
    <div className="flex items-center justify-between gap-3 py-2.5">
      <EntityChip entityId={member.entity_id} />
      <div className="flex shrink-0 items-center gap-2">
        {member.source && <SourcePill source={member.source as GroupSource} />}
        {member.has_expiration ? (
          <span className="hidden text-xs text-muted-foreground sm:inline">
            expires {formatAbsoluteDate(member.expires_at)}
          </span>
        ) : (
          <span className="hidden text-xs text-muted-foreground sm:inline">
            joined {formatDate(member.joined_at)}
          </span>
        )}
      </div>
    </div>
  )
}

function OwnerRow({
  entityId,
  since,
  isAdmin,
}: {
  entityId: string
  since?: string
  isAdmin: boolean
}) {
  return (
    <div className="flex items-center justify-between gap-3 py-2.5">
      <EntityChip entityId={entityId} />
      <div className="flex items-center gap-2">
        {isAdmin && (
          <span className="inline-flex h-6 items-center rounded-md bg-gradient-to-r from-gr-pink to-gr-purple p-px text-xs font-medium">
            <span className="inline-flex h-full items-center gap-1 rounded-[7px] bg-card px-2">
              <Shield className="size-3 text-gr-pink" />
              <span className="bg-gradient-to-r from-gr-pink to-gr-purple bg-clip-text text-transparent">
                admin
              </span>
            </span>
          </span>
        )}
        {since && (
          <span className="text-xs text-muted-foreground">since {formatDate(since)}</span>
        )}
      </div>
    </div>
  )
}

function SyncConfigBlock({
  source,
  discordBindings,
  discordRoles,
  conditionalBindings,
  groupNamesByID,
}: {
  source: GroupSource
  discordBindings?: GroupDiscordRoleBinding[]
  discordRoles?: DiscordRole[]
  conditionalBindings?: GroupConditionalBinding[]
  groupNamesByID?: Record<string, string>
}) {
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
    const bindings = discordBindings ?? []
    return (
      <div className="flex items-start gap-3 py-3">
        <Bot className="mt-0.5 size-4 shrink-0 text-muted-foreground" />
        <div className="min-w-0 flex-1">
          <p className="text-sm font-medium">Discord role sync</p>
          <p className="mt-0.5 text-xs text-muted-foreground">
            Members of any of the role sets below are added automatically.
          </p>
          {bindings.length === 0 ? (
            <p className="mt-2 text-xs italic text-muted-foreground">
              No role bindings configured yet.
            </p>
          ) : (
            <ul className="mt-2 space-y-1.5">
              {bindings.map((binding) => (
                <li key={binding.id} className="flex flex-wrap items-center gap-1.5">
                  {binding.discord_role_ids.map((roleID, idx) => {
                    const role = discordRoles?.find((r) => r.id === roleID)
                    const hex = role ? discordRoleColorHex(role.color) : null
                    return (
                      <span key={roleID} className="flex items-center gap-1.5">
                        {idx > 0 && (
                          <span className="text-xs font-medium text-muted-foreground">
                            AND
                          </span>
                        )}
                        <span className="inline-flex items-center gap-1.5 rounded-md border border-border/60 bg-muted/40 px-2 py-0.5">
                          <span
                            className="size-2 shrink-0 rounded-full border border-border/60"
                            style={{ backgroundColor: hex ?? "transparent" }}
                          />
                          <span className="text-sm">
                            {role ? `@${role.name}` : (
                              <code className="font-mono text-xs text-muted-foreground">
                                {roleID}
                              </code>
                            )}
                          </span>
                        </span>
                      </span>
                    )
                  })}
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>
    )
  }
  // CONDITIONAL
  const bindings = conditionalBindings ?? []
  return (
    <div className="flex items-start gap-3 py-3">
      <Sparkles className="mt-0.5 size-4 shrink-0 text-muted-foreground" />
      <div className="min-w-0 flex-1">
        <p className="text-sm font-medium">Conditional rule</p>
        <p className="mt-0.5 text-xs text-muted-foreground">
          Members of any of the group sets below are added automatically.
        </p>
        {bindings.length === 0 ? (
          <p className="mt-2 text-xs italic text-muted-foreground">
            No conditional bindings configured yet.
          </p>
        ) : (
          <ul className="mt-2 space-y-1.5">
            {bindings.map((binding) => (
              <li key={binding.id} className="flex flex-wrap items-center gap-1.5">
                {binding.required_group_ids.map((groupID, idx) => {
                  const name = groupNamesByID?.[groupID]
                  return (
                    <span key={groupID} className="flex items-center gap-1.5">
                      {idx > 0 && (
                        <span className="text-xs font-medium text-muted-foreground">
                          AND
                        </span>
                      )}
                      <span className="inline-flex items-center gap-1.5 rounded-md border border-border/60 bg-muted/40 px-2 py-0.5">
                        <span className="text-sm">
                          {name ?? (
                            <code className="font-mono text-xs text-muted-foreground">
                              {groupID}
                            </code>
                          )}
                        </span>
                      </span>
                    </span>
                  )
                })}
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  )
}

export default function GroupDetailsPage() {
  const { id } = useParams<{ id: string }>()
  const qc = useQueryClient()
  const myEntityID = loadSession()?.entityId ?? ""
  const { adminIds, isAdmin } = useAdmins()
  const [memberSearch, setMemberSearch] = useState("")
  const [reviewTarget, setReviewTarget] = useState<{
    request: GroupJoinRequest
    action: "approve" | "reject"
  } | null>(null)
  const [joinOpen, setJoinOpen] = useState(false)
  const [joinReason, setJoinReason] = useState("")
  const [submittingJoin, setSubmittingJoin] = useState(false)
  const [joinDuration, setJoinDuration] = useState<DurationPreset | "custom">("1mo")
  const [customAmount, setCustomAmount] = useState(7)
  const [customUnit, setCustomUnit] = useState<DurationUnit>("days")
  const [cancelling, setCancelling] = useState(false)
  const [cancelConfirmOpen, setCancelConfirmOpen] = useState(false)
  const [addPersonOpen, setAddPersonOpen] = useState(false)
  const [addPersonMode, setAddPersonMode] = useState<"member" | "owner">("member")

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

  const discordBindingsQuery = useGroupDiscordBindings(id ?? "")
  const discordRolesQuery = useDiscordRoles()
  const conditionalBindingsQuery = useGroupConditionalBindings(id ?? "")
  const googleBindingQuery = useGroupGoogleBinding(id ?? "")
  // Fetch ALL groups once so we can resolve required_group_ids → names for
  // the conditional-binding chips. Cheap query for typical org scale.
  const allGroupsQuery = useQuery({
    queryKey: ["groups"],
    queryFn: async () => {
      const res = await api.get<Group[]>("/groups")
      return res.data
    },
  })
  const groupNamesByID = useMemo(() => {
    const m: Record<string, string> = {}
    for (const g of allGroupsQuery.data ?? []) m[g.id] = g.name
    return m
  }, [allGroupsQuery.data])


  async function submitJoinRequest() {
    if (!id || !myEntityID || submittingJoin) return

    let expires: Date
    if (joinDuration === "custom") {
      if (
        !Number.isFinite(customAmount) ||
        customAmount <= 0 ||
        customAmount > MAX_BY_UNIT[customUnit]
      ) {
        toast.error(`Pick between 1 and ${MAX_BY_UNIT[customUnit]} ${customUnit}.`)
        return
      }
      expires = addCustom(new Date(), customAmount, customUnit)
    } else {
      expires = addPreset(new Date(), joinDuration)
    }

    setSubmittingJoin(true)
    try {
      const res = await api.post<GroupJoinRequest>(`/groups/${id}/requests`, {
        entity_id: myEntityID,
        has_expiration: true,
        expires_at: expires.toISOString(),
      })
      const reason = joinReason.trim()
      if (reason) {
        await api.post(`/groups/${id}/requests/${res.data.id}/comments`, {
          entity_id: myEntityID,
          comment: reason,
        })
      }
      qc.invalidateQueries({ queryKey: ["group", id, "requests"] })
      qc.invalidateQueries({ queryKey: ["group", id] })
      toast.success("Request submitted")
      setJoinOpen(false)
      setJoinReason("")
    } catch (err: unknown) {
      const message =
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        "Couldn't submit request."
      toast.error(message)
    } finally {
      setSubmittingJoin(false)
    }
  }

  async function cancelMyRequest(requestID: string) {
    if (!id || cancelling) return
    setCancelling(true)
    try {
      await api.delete(`/groups/${id}/requests/${requestID}`)
      qc.invalidateQueries({ queryKey: ["group", id, "requests"] })
      qc.invalidateQueries({ queryKey: ["group", id] })
      toast.success("Request cancelled")
      setCancelConfirmOpen(false)
    } catch (err: unknown) {
      const message =
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        "Couldn't cancel request."
      toast.error(message)
    } finally {
      setCancelling(false)
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
  const allPending = requests.filter((r) => r.status === "PENDING")

  const isExplicitOwner = !!myEntityID && owners.some((o) => o.entity_id === myEntityID)
  const isOwner = isExplicitOwner || isAdmin
  const isMember = !!myEntityID && members.some((m) => m.entity_id === myEntityID)
  const myPending = myEntityID
    ? allPending.find((r) => r.entity_id === myEntityID)
    : undefined
  // Joining is request-based and approvals mint a DIRECT member, so the button
  // only makes sense for groups that allow the DIRECT source. Ownership is
  // tracked separately from membership, so an owner who isn't a member can
  // still request to join.
  const canRequestToJoin =
    (group.allowed_sources?.includes("DIRECT") ?? false) && !isMember && !myPending
  // Hide the viewer's own request from the owner inbox; they manage it from
  // their "Request pending" dialog instead.
  const pending = myEntityID
    ? allPending.filter((r) => r.entity_id !== myEntityID)
    : allPending

  // Owners list = explicit owners + admins (who get owner-equivalent rights).
  // Dedupe by entity_id, prefer the explicit-owner entry so we keep the
  // "since X" date when both apply.
  const ownerRows = (() => {
    const seen = new Set<string>()
    const rows: { entityId: string; since?: string; isAdmin: boolean }[] = []
    for (const o of owners) {
      seen.add(o.entity_id)
      rows.push({
        entityId: o.entity_id,
        since: o.created_at,
        isAdmin: adminIds.has(o.entity_id),
      })
    }
    for (const adminId of adminIds) {
      if (seen.has(adminId)) continue
      rows.push({ entityId: adminId, isAdmin: true })
    }
    return rows
  })()

  const needle = memberSearch.trim()
  const searching = needle.length > 0
  const matchedMembers = searching
    ? fuzzyFilter(members, needle, (m) => [m.entity_id])
    : members
  const visibleMembers = searching
    ? matchedMembers
    : matchedMembers.slice(0, MEMBER_PREVIEW_COUNT)
  const remainingMembers = searching
    ? 0
    : Math.max(0, members.length - visibleMembers.length)

  const directCount = members.filter((m) => m.source === "DIRECT").length
  const syncedCount = members.filter((m) => m.source === "DISCORD" || m.source === "CONDITIONAL").length
  const directMembershipsEnabled = group.allowed_sources?.includes("DIRECT") ?? false

  function openAddPerson(mode: "member" | "owner") {
    setAddPersonMode(mode)
    setAddPersonOpen(true)
  }

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
        <div className="flex flex-wrap items-center gap-2">
          {isOwner && <OutlinePill icon={Crown} label="Owner" tone="gold" />}
          {isMember && <OutlinePill icon={Check} label="Member" tone="green" />}
          {myPending && (
            <Button
              asChild
              variant="outline"
              className="h-10 gap-1.5 rounded-xl px-4 text-sm"
            >
              <Link to={`/groups/${group.id}/requests/${myPending.id}`}>
                <Hourglass className="size-3.5" />
                Request pending
              </Link>
            </Button>
          )}
          {canRequestToJoin && (
            <Button
              className="h-10 gap-1.5 rounded-xl px-4 text-sm"
              onClick={() => setJoinOpen(true)}
            >
              <UserPlus className="size-3.5" />
              Request to join
            </Button>
          )}
          {isOwner && (
            <Button asChild variant="outline" className="h-10 gap-1.5 rounded-xl px-4 text-sm">
              <Link to={`/groups/${group.id}/edit`}>
                <Pencil className="size-3.5" />
                Edit
              </Link>
            </Button>
          )}
        </div>
      </header>

      <div className="space-y-4">
        {myPending && (() => {
          const reason = myPending.comments?.find((c) => c.entity_id === myEntityID)?.comment
          return (
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Hourglass className="size-4 text-gr-pink" />
                  Your request is pending
                </CardTitle>
                <CardDescription>
                  Submitted {relativeTime(myPending.created_at)}
                  {myPending.has_expiration && (
                    <> · {formatDurationBetween(myPending.created_at, myPending.expires_at)} of access</>
                  )}
                  . Owners haven't reviewed it yet.
                </CardDescription>
                {myPending.has_expiration && (
                  <CardAction className="text-right">
                    <p className="text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
                      Would expire
                    </p>
                    <p className="mt-0.5 text-base font-semibold">
                      {formatAbsoluteDate(myPending.expires_at)}
                    </p>
                    <p className="text-xs text-muted-foreground">
                      {formatExpiresIn(myPending.expires_at)}
                    </p>
                  </CardAction>
                )}
              </CardHeader>
              <CardContent className="space-y-4">
                {reason && (
                  <div>
                    <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
                      Your note
                    </p>
                    <p className="mt-2 rounded-md border border-border/60 bg-muted/30 p-3 text-sm text-muted-foreground">
                      {reason}
                    </p>
                  </div>
                )}
                <div className="flex flex-wrap justify-end gap-2">
                  <Button
                    type="button"
                    variant="destructive"
                    disabled={cancelling}
                    onClick={() => setCancelConfirmOpen(true)}
                  >
                    Cancel request
                  </Button>
                  <Button asChild variant="outline">
                    <Link to={`/groups/${group.id}/requests/${myPending.id}`}>
                      View details
                    </Link>
                  </Button>
                </div>
              </CardContent>
            </Card>
          )
        })()}

        {isOwner && pending.length > 0 && (
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
              <CardAction>
                <Button asChild variant="ghost" size="sm" className="text-muted-foreground">
                  <Link to={`/groups/${group.id}/requests`}>View all</Link>
                </Button>
              </CardAction>
            </CardHeader>
            <CardContent className="space-y-3">
              {pending.map((req) => {
                const reason = req.comments?.find((c) => c.entity_id === req.entity_id)?.comment
                return (
                  <Link
                    key={req.id}
                    to={`/groups/${group.id}/requests/${req.id}`}
                    className="block rounded-lg border border-border/60 bg-muted/30 p-3 transition-colors hover:bg-muted/50"
                  >
                    <div className="flex items-start justify-between gap-3">
                      <div className="min-w-0 flex-1">
                        <EntityChip entityId={req.entity_id} />
                        <p className="mt-2 text-xs text-muted-foreground">
                          Requested {relativeTime(req.created_at)}
                          {req.has_expiration && (
                            <>
                              {" · "}
                              {formatDurationBetween(req.created_at, req.expires_at)} of access
                            </>
                          )}
                        </p>
                        {reason && (
                          <p className="mt-2 rounded-md border border-border/60 bg-background/60 p-2.5 text-sm text-muted-foreground">
                            {reason}
                          </p>
                        )}
                      </div>
                      <div className="flex gap-2" onClick={(e) => e.preventDefault()}>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => setReviewTarget({ request: req, action: "reject" })}
                        >
                          Reject
                        </Button>
                        <Button
                          size="sm"
                          onClick={() => setReviewTarget({ request: req, action: "approve" })}
                        >
                          Approve
                        </Button>
                      </div>
                    </div>
                  </Link>
                )
              })}
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
                      <SyncConfigBlock
                        key={source}
                        source={source}
                        discordBindings={discordBindingsQuery.data}
                        discordRoles={discordRolesQuery.data}
                        conditionalBindings={conditionalBindingsQuery.data}
                        groupNamesByID={groupNamesByID}
                      />
                    ))
                  ) : (
                    <p className="py-3 text-sm text-muted-foreground">
                      No sources configured yet.
                    </p>
                  )}
                </div>
                <Link
                  to={`/groups/${group.id}/requests`}
                  className="mt-4 flex items-center justify-between gap-3 rounded-md border border-border/60 bg-muted/30 px-3 py-2.5 text-sm transition-colors hover:bg-muted/50"
                >
                  <div className="flex items-center gap-2 text-muted-foreground">
                    <Inbox className="size-4" />
                    <span>
                      <span className="font-medium text-foreground">{group.pending_count}</span>{" "}
                      pending {group.pending_count === 1 ? "request" : "requests"}
                    </span>
                  </div>
                  <span className="text-xs text-muted-foreground">View all →</span>
                </Link>
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

                {googleBindingQuery.data && (
                  <section className="border-t border-border/60 pt-6">
                    <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
                      Google Group
                    </p>
                    <p className="mt-1 text-xs text-muted-foreground">
                      Members are mirrored into this Google Group.
                    </p>
                    <div className="mt-3 flex items-center gap-2.5 rounded-md border border-border/60 bg-muted/40 px-3 py-2">
                      <Mail className="size-4 shrink-0 text-muted-foreground" />
                      <span className="truncate font-mono text-sm">
                        {googleBindingQuery.data.google_group_email}
                      </span>
                    </div>
                  </section>
                )}

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
              Can edit the group, manage members, and approve requests. Global admins inherit these permissions.
            </CardDescription>
            {isOwner && (
              <CardAction>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  className="gap-1.5"
                  onClick={() => openAddPerson("owner")}
                >
                  <Crown className="size-3.5" />
                  Add owner
                </Button>
              </CardAction>
            )}
          </CardHeader>
          <CardContent>
            {ownersQuery.isLoading ? (
              <Skeleton className="h-10 w-full" />
            ) : ownerRows.length === 0 ? (
              <p className="text-sm text-muted-foreground">No owners assigned.</p>
            ) : (
              <ul className="divide-y divide-border/60">
                {ownerRows.map((row) => (
                  <li key={row.entityId}>
                    <OwnerRow
                      entityId={row.entityId}
                      since={row.since}
                      isAdmin={row.isAdmin}
                    />
                  </li>
                ))}
              </ul>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Members</CardTitle>
            <CardDescription>
              {members.length} total · {directCount} direct · {syncedCount} synced
            </CardDescription>
            {isOwner && (
              <CardAction>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  className="gap-1.5"
                  disabled={!directMembershipsEnabled}
                  title={
                    directMembershipsEnabled
                      ? undefined
                      : "Enable direct source before adding direct members"
                  }
                  onClick={() => openAddPerson("member")}
                >
                  <UserPlus className="size-3.5" />
                  Add member
                </Button>
              </CardAction>
            )}
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

      <Dialog
        open={joinOpen}
        onOpenChange={(open) => {
          if (!submittingJoin) setJoinOpen(open)
        }}
      >
        <DialogContent className="gap-5 sm:max-w-md">
          <DialogHeader className="gap-3">
            <div className="flex size-10 items-center justify-center rounded-xl bg-gradient-to-br from-gr-pink to-gr-purple text-white">
              <UserPlus className="size-5" />
            </div>
            <DialogTitle>Request to join {group.name}</DialogTitle>
            <DialogDescription>
              Owners will review your request. Direct memberships are time-boxed — pick how long you need access.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-2">
            <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
              Membership duration
            </p>
            <div className="flex flex-wrap gap-2 pt-1">
              {DURATION_PRESETS.map((p) => {
                const active = joinDuration === p.value
                return (
                  <button
                    key={p.value}
                    type="button"
                    onClick={() => setJoinDuration(p.value)}
                    className={
                      "rounded-full border px-3 py-1.5 text-sm transition-colors " +
                      (active
                        ? "border-gr-pink/40 bg-gr-pink/10 text-foreground"
                        : "border-border/60 bg-muted/30 text-muted-foreground hover:bg-muted/50")
                    }
                  >
                    {p.label}
                  </button>
                )
              })}
              <button
                type="button"
                onClick={() => setJoinDuration("custom")}
                className={
                  "rounded-full border px-3 py-1.5 text-sm transition-colors " +
                  (joinDuration === "custom"
                    ? "border-gr-pink/40 bg-gr-pink/10 text-foreground"
                    : "border-border/60 bg-muted/30 text-muted-foreground hover:bg-muted/50")
                }
              >
                Custom
              </button>
            </div>
            {joinDuration === "custom" && (
              <div className="flex items-center gap-2 pt-2">
                <Input
                  type="number"
                  min={1}
                  max={MAX_BY_UNIT[customUnit]}
                  value={customAmount}
                  onChange={(e) => setCustomAmount(parseInt(e.target.value, 10) || 0)}
                  className="w-24"
                />
                <Select
                  value={customUnit}
                  onValueChange={(v) => setCustomUnit(v as DurationUnit)}
                >
                  <SelectTrigger className="w-32">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="hours">Hours</SelectItem>
                    <SelectItem value="days">Days</SelectItem>
                    <SelectItem value="weeks">Weeks</SelectItem>
                    <SelectItem value="months">Months</SelectItem>
                  </SelectContent>
                </Select>
                <p className="text-xs text-muted-foreground">
                  Max {MAX_BY_UNIT[customUnit]}
                </p>
              </div>
            )}
          </div>

          <div className="space-y-2">
            <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
              Note
            </p>
            <Textarea
              placeholder="Why do you want to join? (optional)"
              value={joinReason}
              onChange={(e) => setJoinReason(e.target.value)}
              rows={3}
            />
          </div>

          <div className="flex justify-end gap-2 pt-1">
            <Button
              type="button"
              variant="ghost"
              disabled={submittingJoin}
              onClick={() => setJoinOpen(false)}
            >
              Cancel
            </Button>
            <Button type="button" disabled={submittingJoin} onClick={submitJoinRequest}>
              {submittingJoin ? "Submitting…" : "Submit request"}
            </Button>
          </div>
        </DialogContent>
      </Dialog>

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
              disabled={cancelling || !myPending}
              onClick={() => myPending && cancelMyRequest(myPending.id)}
            >
              {cancelling ? "Cancelling…" : "Cancel request"}
            </Button>
          </div>
        </DialogContent>
      </Dialog>

      <ReviewRequestDialog
        open={reviewTarget !== null}
        onOpenChange={(o) => {
          if (!o) setReviewTarget(null)
        }}
        groupID={id ?? ""}
        request={reviewTarget?.request ?? null}
        action={reviewTarget?.action ?? "approve"}
        reviewerEntityID={myEntityID}
      />

      {addPersonOpen && (
        <AddGroupPersonDialog
          open={addPersonOpen}
          onOpenChange={setAddPersonOpen}
          groupID={id ?? ""}
          groupName={group.name}
          initialMode={addPersonMode}
          existingMemberEntityIDs={members.map((member) => member.entity_id)}
          existingOwnerEntityIDs={owners.map((owner) => owner.entity_id)}
        />
      )}
    </PageContainer>
  )
}
