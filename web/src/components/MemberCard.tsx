import { useQuery } from "@tanstack/react-query"
import { Link } from "react-router-dom"

import { DiscordIcon, GithubIcon } from "@/components/icons/socials"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { api } from "@/lib/api"
import type { Entity } from "@/lib/auth"
import { academicSummary, userInitials, userName, type Member } from "@/lib/users"
import { cn } from "@/lib/utils"

export function MemberAvatar({
  member,
  className,
  fallbackClassName,
}: {
  member: Member
  className?: string
  fallbackClassName?: string
}) {
  return (
    <Avatar className={cn("size-10", className)}>
      <AvatarImage src={member.avatar_url || undefined} alt={userName(member)} />
      <AvatarFallback
        className={cn(
          "bg-gradient-to-br from-gr-pink to-gr-purple text-sm font-semibold text-white",
          fallbackClassName,
        )}
      >
        {userInitials(member)}
      </AvatarFallback>
    </Avatar>
  )
}

export function MemberLinkedAccounts({ entityID }: { entityID: string }) {
  const entityQuery = useQuery({
    queryKey: ["entity", entityID],
    queryFn: async ({ signal }) =>
      (await api.get<Entity>(`/entities/${entityID}`, { signal })).data,
    staleTime: 5 * 60 * 1000,
  })

  if (entityQuery.isPending) {
    return <p className="text-xs text-muted-foreground">Loading accounts…</p>
  }
  if (entityQuery.isError) {
    return <p className="text-xs text-muted-foreground">Accounts unavailable</p>
  }

  const accounts = entityQuery.data.external_auths ?? []
  const github = accounts.find((account) => account.provider === "GITHUB")
  const discord = accounts.find((account) => account.provider === "DISCORD")

  if (!github && !discord) {
    return <p className="text-xs text-muted-foreground">No GitHub or Discord accounts</p>
  }

  return (
    <div className="space-y-1 text-xs">
      {github && (
        <p className="flex items-start gap-1.5">
          <GithubIcon className="mt-0.5 size-3.5 shrink-0" />
          <span className="min-w-0 break-all">
            <span className="sr-only">GitHub: </span>
            {github.metadata?.username || `ID: ${github.external_id}`}
          </span>
        </p>
      )}
      {discord && (
        <p className="flex items-start gap-1.5">
          <DiscordIcon className="mt-0.5 size-3.5 shrink-0" />
          <span className="min-w-0 break-all">
            <span className="sr-only">Discord: </span>
            {discord.metadata?.username || `ID: ${discord.external_id}`}
          </span>
        </p>
      )}
    </div>
  )
}

export function MemberCard({ member }: { member: Member }) {
  const academic = academicSummary(member)
  return (
    <Link
      to={`/members/${member.entity_id}`}
      className="group flex flex-col gap-3 rounded-lg border border-border/60 bg-card p-4 transition-colors hover:bg-muted/40"
    >
      <div className="flex items-start gap-3">
        <MemberAvatar member={member} />
        <div className="min-w-0 flex-1">
          <p className="truncate text-sm font-medium leading-none">{userName(member)}</p>
          <p className="mt-1 truncate text-xs text-muted-foreground">
            {member.username ? `@${member.username}` : member.entity_id}
          </p>
        </div>
      </div>

      {academic && <p className="truncate text-xs text-muted-foreground">{academic}</p>}

      <div className="min-w-0">
        <p className="mb-1 text-xs text-muted-foreground">Email</p>
        <p className="break-all text-sm">{member.email || "No email"}</p>
      </div>

      <div className="mt-auto min-w-0">
        <p className="mb-1 text-xs text-muted-foreground">Linked accounts</p>
        <MemberLinkedAccounts entityID={member.entity_id} />
      </div>
    </Link>
  )
}
