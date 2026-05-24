import { useQuery } from "@tanstack/react-query"
import { ArrowRight, ChevronRight } from "lucide-react"
import { useMemo } from "react"
import { Link } from "react-router-dom"

import { PageContainer, PageHeader } from "@/components/PageContainer"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { api } from "@/lib/api"

type DiscordRole = {
  id: string
  name: string
  color: number
  position: number
  hoist: boolean
  mentionable: boolean
  managed: boolean
}

type DiscordChannel = {
  id: string
  name: string
  type: number
  position: number
  parent_id: string
  topic: string
  nsfw: boolean
}

const CHANNEL_TYPE_LABELS: Record<number, string> = {
  0: "text",
  2: "voice",
  4: "category",
  5: "announcement",
  10: "news-thread",
  11: "public-thread",
  12: "private-thread",
  13: "stage",
  15: "forum",
  16: "media",
}

function roleColorHex(color: number): string | null {
  if (!color) return null
  return `#${color.toString(16).padStart(6, "0")}`
}

const CHANNEL_TYPE_CATEGORY = 4

type OrganizedChannels = {
  topLevel: DiscordChannel[]
  categories: DiscordChannel[]
  childrenByParent: Map<string, DiscordChannel[]>
}

function organizeChannels(channels: DiscordChannel[]): OrganizedChannels {
  const categories = channels
    .filter((c) => c.type === CHANNEL_TYPE_CATEGORY)
    .sort((a, b) => a.position - b.position)
  const categoryIds = new Set(categories.map((c) => c.id))

  const topLevel: DiscordChannel[] = []
  const childrenByParent = new Map<string, DiscordChannel[]>()

  for (const c of channels) {
    if (c.type === CHANNEL_TYPE_CATEGORY) continue
    // Channel is top-level if it has no parent, or its parent isn't a category
    // we saw (orphaned — shouldn't happen, but defensive).
    if (!c.parent_id || !categoryIds.has(c.parent_id)) {
      topLevel.push(c)
      continue
    }
    const list = childrenByParent.get(c.parent_id) ?? []
    list.push(c)
    childrenByParent.set(c.parent_id, list)
  }

  topLevel.sort((a, b) => a.position - b.position)
  for (const list of childrenByParent.values()) {
    list.sort((a, b) => a.position - b.position)
  }

  return { topLevel, categories, childrenByParent }
}

function ChannelRow({ channel, indent }: { channel: DiscordChannel; indent?: boolean }) {
  const typeLabel = CHANNEL_TYPE_LABELS[channel.type] ?? `type ${channel.type}`
  return (
    <li
      className={`flex items-center justify-between gap-4 py-3 pr-6 ${
        indent ? "pl-12" : "pl-6"
      }`}
    >
      <div className="flex min-w-0 items-center gap-3">
        <Badge variant="outline" className="shrink-0 font-mono">
          {typeLabel}
        </Badge>
        <div className="min-w-0">
          <p className="truncate text-sm font-medium leading-none">{channel.name}</p>
          <p className="mt-1 font-mono text-xs text-muted-foreground">{channel.id}</p>
        </div>
      </div>
      <div className="flex shrink-0 items-center gap-1.5">
        {channel.nsfw && <Badge variant="destructive">nsfw</Badge>}
        <span className="ml-2 font-mono text-xs text-muted-foreground">
          pos {channel.position}
        </span>
      </div>
    </li>
  )
}

type LinkGroup = {
  title: string
  description: string
  links: Array<{ to: string; label: string; note?: string }>
}

const SAMPLE_AUTHORIZE = new URLSearchParams({
  client_id: "sentinel",
  redirect_uri: "http://localhost:3000/auth/callback",
  scope: "user:read groups:read",
}).toString()

const GROUPS: LinkGroup[] = [
  {
    title: "Dashboard",
    description: "Pages reachable from the sidebar.",
    links: [
      { to: "/", label: "Home" },
      { to: "/applications", label: "Applications" },
      { to: "/groups", label: "Groups" },
      { to: "/analytics", label: "Analytics" },
      { to: "/settings", label: "Settings" },
    ],
  },
  {
    title: "Auth flow",
    description: "Full-bleed pages outside the dashboard shell.",
    links: [
      { to: "/auth/login", label: "Login" },
      {
        to: `/oauth/authorize?${SAMPLE_AUTHORIZE}`,
        label: "OAuth authorize",
        note: "with sample client_id, redirect_uri, scope",
      },
      {
        to: "/onboard?token=mock_verify_token",
        label: "Onboarding",
        note: "Discord !verify → DM'd link, multi-step account setup",
      },
      { to: "/onboard", label: "Onboarding (no token)", note: "invalid invite fallback" },
    ],
  },
  {
    title: "Edge cases",
    description: "States that aren't part of the normal nav.",
    links: [{ to: "/this-route-does-not-exist", label: "404" }],
  },
]

function LinkRow({ to, label, note }: { to: string; label: string; note?: string }) {
  return (
    <Link
      to={to}
      className="flex items-center justify-between gap-4 px-6 py-3 transition-colors hover:bg-muted/40"
    >
      <div>
        <p className="text-sm font-medium leading-none">{label}</p>
        {note && <p className="mt-1 text-xs text-muted-foreground">{note}</p>}
      </div>
      <div className="flex items-center gap-3 text-xs text-muted-foreground">
        <span className="font-mono">{to}</span>
        <ArrowRight className="size-3.5" />
      </div>
    </Link>
  )
}

function CollapsibleCard({
  title,
  description,
  count,
  children,
}: {
  title: string
  description: React.ReactNode
  count?: number
  children: React.ReactNode
}) {
  return (
    <Card className="overflow-hidden p-0">
      <details className="group/details">
        <summary className="flex cursor-pointer list-none items-center gap-3 px-6 py-4 transition-colors hover:bg-muted/40 [&::-webkit-details-marker]:hidden">
          <ChevronRight className="size-4 shrink-0 text-muted-foreground transition-transform group-open/details:rotate-90" />
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2">
              <p className="text-base font-semibold leading-none">{title}</p>
              {typeof count === "number" && (
                <span className="font-mono text-xs text-muted-foreground">({count})</span>
              )}
            </div>
            <p className="mt-1 text-sm text-muted-foreground">{description}</p>
          </div>
        </summary>
        <div className="border-t border-border">{children}</div>
      </details>
    </Card>
  )
}

function DiscordRolesCard() {
  const query = useQuery({
    queryKey: ["debug", "discord", "roles"],
    queryFn: async () => {
      const res = await api.get<DiscordRole[]>("/discord/roles")
      return res.data
    },
  })

  // Backend returns roles ascending by position (Discord API ordering); reverse
  // here so the top of the role hierarchy (admin-style roles) shows first,
  // matching how Discord's own UI presents the role list.
  const ordered = query.data ? [...query.data].reverse() : undefined

  return (
    <CollapsibleCard
      title="Discord roles"
      description={
        <>
          Live from <code className="font-mono">GET /discord/roles</code>, highest position first.
        </>
      }
      count={ordered?.length}
    >
      {query.isLoading && (
        <div className="space-y-2 px-6 py-4">
          <Skeleton className="h-6 w-full" />
          <Skeleton className="h-6 w-full" />
          <Skeleton className="h-6 w-3/4" />
        </div>
      )}
      {query.isError && (
        <p className="px-6 py-4 text-sm text-destructive">
          Failed to fetch roles: {(query.error as Error).message}
        </p>
      )}
      {ordered && ordered.length === 0 && (
        <p className="px-6 py-4 text-sm text-muted-foreground">No roles returned.</p>
      )}
      {ordered && ordered.length > 0 && (
        <ul className="divide-y divide-border">
          {ordered.map((role) => {
            const hex = roleColorHex(role.color)
            return (
              <li
                key={role.id}
                className="flex items-center justify-between gap-4 px-6 py-3"
              >
                <div className="flex min-w-0 items-center gap-3">
                  <span
                    className="size-3 shrink-0 rounded-full border border-border/60"
                    style={{ backgroundColor: hex ?? "transparent" }}
                    title={hex ?? "no color"}
                  />
                  <div className="min-w-0">
                    <p className="truncate text-sm font-medium leading-none">{role.name}</p>
                    <p className="mt-1 font-mono text-xs text-muted-foreground">{role.id}</p>
                  </div>
                </div>
                <div className="flex shrink-0 items-center gap-1.5">
                  {role.managed && <Badge variant="secondary">managed</Badge>}
                  {role.hoist && <Badge variant="outline">hoist</Badge>}
                  {role.mentionable && <Badge variant="outline">mention</Badge>}
                  <span className="ml-2 font-mono text-xs text-muted-foreground">
                    pos {role.position}
                  </span>
                </div>
              </li>
            )
          })}
        </ul>
      )}
    </CollapsibleCard>
  )
}

function DiscordChannelsCard() {
  const query = useQuery({
    queryKey: ["debug", "discord", "channels"],
    queryFn: async () => {
      const res = await api.get<DiscordChannel[]>("/discord/channels")
      return res.data
    },
  })

  const organized = useMemo(
    () => (query.data ? organizeChannels(query.data) : null),
    [query.data],
  )

  const isEmpty =
    organized && organized.topLevel.length === 0 && organized.categories.length === 0

  return (
    <CollapsibleCard
      title="Discord channels"
      description={
        <>
          Live from <code className="font-mono">GET /discord/channels</code>, grouped by category.
        </>
      }
      count={query.data?.length}
    >
      {query.isLoading && (
        <div className="space-y-2 px-6 py-4">
          <Skeleton className="h-6 w-full" />
          <Skeleton className="h-6 w-full" />
          <Skeleton className="h-6 w-3/4" />
        </div>
      )}
      {query.isError && (
        <p className="px-6 py-4 text-sm text-destructive">
          Failed to fetch channels: {(query.error as Error).message}
        </p>
      )}
      {isEmpty && (
        <p className="px-6 py-4 text-sm text-muted-foreground">No channels returned.</p>
      )}
      {organized && !isEmpty && (
        <div className="divide-y divide-border">
          {organized.topLevel.length > 0 && (
            <ul className="divide-y divide-border">
              {organized.topLevel.map((c) => (
                <ChannelRow key={c.id} channel={c} />
              ))}
            </ul>
          )}
          {organized.categories.map((category) => {
            const children = organized.childrenByParent.get(category.id) ?? []
            return (
              <section key={category.id}>
                <div className="flex items-center justify-between gap-4 bg-muted/40 px-6 py-2">
                  <div className="flex min-w-0 items-center gap-2">
                    <p className="truncate text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                      {category.name}
                    </p>
                    <p className="font-mono text-xs text-muted-foreground/70">{category.id}</p>
                  </div>
                  <span className="font-mono text-xs text-muted-foreground/70">
                    pos {category.position}
                  </span>
                </div>
                {children.length === 0 ? (
                  <p className="px-12 py-2 text-xs italic text-muted-foreground/70">
                    empty category
                  </p>
                ) : (
                  <ul className="divide-y divide-border">
                    {children.map((c) => (
                      <ChannelRow key={c.id} channel={c} indent />
                    ))}
                  </ul>
                )}
              </section>
            )
          })}
        </div>
      )}
    </CollapsibleCard>
  )
}

export default function DebugPage() {
  return (
    <PageContainer>
      <PageHeader
        title="Debug"
        description="Index of every page in the app, for design and QA."
      />
      <div className="space-y-6">
        {GROUPS.map((group) => (
          <Card key={group.title}>
            <CardHeader>
              <CardTitle>{group.title}</CardTitle>
              <CardDescription>{group.description}</CardDescription>
            </CardHeader>
            <CardContent className="p-0">
              <ul className="divide-y divide-border">
                {group.links.map((link) => (
                  <li key={link.to}>
                    <LinkRow {...link} />
                  </li>
                ))}
              </ul>
            </CardContent>
          </Card>
        ))}
        <DiscordRolesCard />
        <DiscordChannelsCard />
      </div>
    </PageContainer>
  )
}
