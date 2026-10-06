import { ArrowLeft, ChevronRight, Hash } from "lucide-react"
import { useMemo } from "react"
import { Link } from "react-router-dom"

import { PageContainer, PageHeader } from "@/components/PageContainer"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { useAdmins } from "@/lib/admin"
import {
  discordRoleColorHex,
  useDiscordChannels,
  useDiscordRoles,
  type DiscordChannel,
} from "@/lib/discord"

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

  const childrenByParent = new Map<string, DiscordChannel[]>()
  const topLevel: DiscordChannel[] = []
  for (const channel of channels) {
    if (channel.type === CHANNEL_TYPE_CATEGORY) continue
    if (!channel.parent_id) {
      topLevel.push(channel)
      continue
    }
    const siblings = childrenByParent.get(channel.parent_id)
    if (siblings) siblings.push(channel)
    else childrenByParent.set(channel.parent_id, [channel])
  }
  topLevel.sort((a, b) => a.position - b.position)
  for (const siblings of childrenByParent.values()) {
    siblings.sort((a, b) => a.position - b.position)
  }

  return { topLevel, categories, childrenByParent }
}

function ChannelRow({ channel, indent }: { channel: DiscordChannel; indent?: boolean }) {
  return (
    <li
      className={`flex items-center justify-between gap-4 py-3 pr-6 ${indent ? "pl-12" : "pl-6"}`}
    >
      <div className="flex min-w-0 items-center gap-3">
        <Badge variant="outline" className="font-mono text-[10px]">
          {CHANNEL_TYPE_LABELS[channel.type] ?? channel.type}
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
  const query = useDiscordRoles()

  // Backend returns roles ascending by position (Discord API ordering); reverse
  // here so the top of the role hierarchy (admin-style roles) shows first,
  // matching how Discord's own UI presents the role list.
  const ordered = query.data ? [...query.data].reverse() : undefined

  return (
    <CollapsibleCard
      title="Roles"
      description="Guild roles, highest position first. Bind these to groups from a group's edit page."
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
            const hex = discordRoleColorHex(role.color)
            return (
              <li key={role.id} className="flex items-center justify-between gap-4 px-6 py-3">
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
  const query = useDiscordChannels()

  const organized = useMemo(
    () => (query.data ? organizeChannels(query.data) : null),
    [query.data],
  )

  const isEmpty =
    organized && organized.topLevel.length === 0 && organized.categories.length === 0

  return (
    <CollapsibleCard
      title="Channels"
      description="Guild channels, grouped by category."
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

export default function DiscordDirectoryPage() {
  const { isAdmin, isLoading } = useAdmins()

  return (
    <PageContainer>
      <Button asChild variant="ghost" size="sm" className="-ml-2 mb-4 text-muted-foreground">
        <Link to="/settings">
          <ArrowLeft className="size-4" />
          Settings
        </Link>
      </Button>
      <PageHeader
        title="Discord directory"
        description="Roles and channels as Sentinel currently sees them in the guild."
      />

      {isLoading ? (
        <div className="space-y-4">
          <Skeleton className="h-20 rounded-lg" />
          <Skeleton className="h-20 rounded-lg" />
        </div>
      ) : isAdmin ? (
        <div className="space-y-4">
          <DiscordRolesCard />
          <DiscordChannelsCard />
        </div>
      ) : (
        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <Hash className="size-4 text-muted-foreground" />
              <CardTitle>Restricted</CardTitle>
            </div>
            <CardDescription>The Discord directory is available to admins.</CardDescription>
          </CardHeader>
          <CardContent className="text-sm text-muted-foreground">
            Group owners can still pick Discord roles from their own group's edit page.
          </CardContent>
        </Card>
      )}
    </PageContainer>
  )
}
