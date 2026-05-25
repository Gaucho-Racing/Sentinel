import {
  ArrowLeft,
  Bot,
  Crown,
  Inbox,
  MessageSquare,
  Pencil,
  Search,
  Sparkles,
  UserPlus,
} from "lucide-react"
import { useState } from "react"
import { Link, useParams } from "react-router-dom"

import { OutlineButton } from "@/components/OutlineButton"
import { PageContainer } from "@/components/PageContainer"
import { Avatar, AvatarFallback } from "@/components/ui/avatar"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import {
  getMockGroup,
  MOCK_JOIN_REQUESTS,
  MOCK_LINKED_APPS,
  MOCK_MEMBERS,
  MOCK_OWNERS,
  SOURCE_LABEL,
  syncConfigsFor,
  type GroupSource,
  type MockMember,
  type MockOwner,
  type MockSyncConfig,
} from "@/lib/groups"

const MEMBER_PREVIEW_COUNT = 6

function initials(name: string) {
  return name
    .split(/\s+/)
    .map((p) => p[0])
    .filter(Boolean)
    .slice(0, 2)
    .join("")
    .toUpperCase()
}

function formatDate(iso: string) {
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

function PersonRow({
  name,
  username,
  trailing,
}: {
  name: string
  username: string
  trailing?: React.ReactNode
}) {
  return (
    <div className="flex items-center justify-between gap-3 py-2.5">
      <div className="flex min-w-0 items-center gap-2.5">
        <Avatar className="size-8">
          <AvatarFallback className="text-xs">{initials(name)}</AvatarFallback>
        </Avatar>
        <div className="flex min-w-0 flex-col leading-tight">
          <span className="truncate text-sm">{name}</span>
          <span className="truncate text-xs text-muted-foreground">@{username}</span>
        </div>
      </div>
      {trailing && <div className="flex shrink-0 items-center gap-2">{trailing}</div>}
    </div>
  )
}

function SourcePill({ source }: { source: GroupSource }) {
  return (
    <Badge variant="outline" className="font-mono text-[10px]">
      {SOURCE_LABEL[source]}
    </Badge>
  )
}

function MemberRow({ member }: { member: MockMember }) {
  return (
    <PersonRow
      name={member.display_name}
      username={member.username}
      trailing={
        <>
          <SourcePill source={member.source} />
          <span className="hidden text-xs text-muted-foreground sm:inline">
            joined {formatDate(member.joined_at)}
          </span>
        </>
      }
    />
  )
}

function OwnerRow({ owner }: { owner: MockOwner }) {
  return (
    <PersonRow
      name={owner.display_name}
      username={owner.username}
      trailing={
        <span className="text-xs text-muted-foreground">since {formatDate(owner.added_at)}</span>
      }
    />
  )
}

function SyncConfigBlock({ config }: { config: MockSyncConfig }) {
  if (config.source === "DIRECT") {
    return (
      <div className="flex items-start gap-3 py-3">
        <UserPlus className="mt-0.5 size-4 shrink-0 text-muted-foreground" />
        <div className="min-w-0">
          <p className="text-sm font-medium">Direct invitation</p>
          <p className="mt-0.5 text-xs text-muted-foreground">
            Owners add members manually or approve join requests. Default for human-curated groups.
          </p>
        </div>
      </div>
    )
  }
  if (config.source === "DISCORD") {
    const hex = `#${config.discord_role_color.toString(16).padStart(6, "0")}`
    return (
      <div className="flex items-start gap-3 py-3">
        <Bot className="mt-0.5 size-4 shrink-0 text-muted-foreground" />
        <div className="min-w-0 flex-1">
          <p className="text-sm font-medium">Discord role sync</p>
          <p className="mt-0.5 text-xs text-muted-foreground">
            Members with this Discord role are added automatically; removing the role removes them.
          </p>
          <div className="mt-2 inline-flex items-center gap-2 rounded-md border border-border/60 bg-muted/40 px-2.5 py-1">
            <span
              className="size-2.5 rounded-full border border-border/60"
              style={{ backgroundColor: hex }}
            />
            <span className="font-mono text-xs">@{config.discord_role_name}</span>
          </div>
        </div>
      </div>
    )
  }
  return (
    <div className="flex items-start gap-3 py-3">
      <Sparkles className="mt-0.5 size-4 shrink-0 text-muted-foreground" />
      <div className="min-w-0 flex-1">
        <p className="text-sm font-medium">Conditional rule</p>
        <p className="mt-0.5 text-xs text-muted-foreground">
          Members are added automatically when their profile matches this expression.
        </p>
        <code className="mt-2 block break-all rounded-md border border-border/60 bg-muted/40 px-2.5 py-1.5 font-mono text-xs">
          {config.rule_summary}
        </code>
      </div>
    </div>
  )
}

export default function GroupDetailsPage() {
  const { id } = useParams<{ id: string }>()
  const group = getMockGroup(id)
  const [memberSearch, setMemberSearch] = useState("")

  if (!group) {
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

  const needle = memberSearch.trim().toLowerCase()
  const searching = needle.length > 0
  const matchedMembers = searching
    ? MOCK_MEMBERS.filter(
        (m) =>
          m.display_name.toLowerCase().includes(needle) ||
          m.username.toLowerCase().includes(needle),
      )
    : MOCK_MEMBERS
  const visibleMembers = searching ? matchedMembers : matchedMembers.slice(0, MEMBER_PREVIEW_COUNT)
  const remainingMembers = searching ? 0 : Math.max(0, group.member_count - visibleMembers.length)
  const visibleOwners = MOCK_OWNERS.slice(0, group.owner_count)
  const pending = group.pending_requests > 0 ? MOCK_JOIN_REQUESTS.slice(0, group.pending_requests) : []
  const syncConfigs = syncConfigsFor(group)

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
              {group.allowed_sources.map((source) => (
                <SourcePill key={source} source={source} />
              ))}
            </div>
          </div>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" className="h-10 gap-1.5 rounded-xl px-4 text-sm">
            <Pencil className="size-3.5" />
            Edit
          </Button>
          <OutlineButton type="button" className="w-auto">
            <UserPlus className="size-3.5" />
            Add member
          </OutlineButton>
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
                    <div className="flex min-w-0 items-center gap-2.5">
                      <Avatar className="size-8">
                        <AvatarFallback className="text-xs">
                          {initials(req.requester_name)}
                        </AvatarFallback>
                      </Avatar>
                      <div className="min-w-0">
                        <p className="truncate text-sm font-medium leading-none">
                          {req.requester_name}
                        </p>
                        <p className="truncate text-xs text-muted-foreground">
                          @{req.requester_username} · {relativeTime(req.created_at)}
                        </p>
                      </div>
                    </div>
                    <div className="flex gap-2">
                      <Button variant="ghost" size="sm">
                        Reject
                      </Button>
                      <Button size="sm">Approve</Button>
                    </div>
                  </div>
                  <p className="mt-2.5 text-sm text-muted-foreground">{req.reason}</p>
                  {req.comment_count > 0 && (
                    <div className="mt-2.5 flex items-center gap-1.5 text-xs text-muted-foreground">
                      <MessageSquare className="size-3" />
                      {req.comment_count} {req.comment_count === 1 ? "comment" : "comments"}
                    </div>
                  )}
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
                  {syncConfigs.map((config, i) => (
                    <SyncConfigBlock key={i} config={config} />
                  ))}
                </div>
              </section>

              <div className="space-y-6">
                <section>
                  <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
                    Linked applications
                  </p>
                  <ul className="mt-3 space-y-2">
                    {MOCK_LINKED_APPS.map((app) => (
                      <li
                        key={app.id}
                        className="flex items-center gap-2.5 rounded-md border border-border/60 bg-muted/40 px-3 py-2"
                      >
                        <div className="flex size-7 shrink-0 items-center justify-center overflow-hidden rounded bg-gradient-to-br from-gr-pink to-gr-purple text-xs font-semibold text-white">
                          {app.name.slice(0, 1).toUpperCase()}
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
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Created</span>
                      <span>—</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Last updated</span>
                      <span>—</span>
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
            <ul className="divide-y divide-border/60">
              {visibleOwners.map((o) => (
                <li key={o.entity_id}>
                  <OwnerRow owner={o} />
                </li>
              ))}
            </ul>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex-row items-start justify-between gap-4">
            <div>
              <CardTitle>Members</CardTitle>
              <CardDescription>
                {group.member_count} total ·{" "}
                {MOCK_MEMBERS.filter((m) => m.source === "DIRECT").length} direct ·{" "}
                {MOCK_MEMBERS.filter((m) => m.source === "DISCORD").length} synced
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
            {visibleMembers.length === 0 ? (
              <p className="py-6 text-center text-sm text-muted-foreground">
                No members match "{memberSearch}".
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
