import { Link } from "react-router-dom"

import { DiscordIcon, GithubIcon } from "@/components/icons/socials"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Badge } from "@/components/ui/badge"
import { academicSummary, userInitials, userName, type LinkedAccount, type Member } from "@/lib/users"
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

export function MemberAccounts({ accounts }: { accounts?: LinkedAccount[] }) {
  if (!accounts?.length) {
    return <span className="text-xs text-muted-foreground">No linked accounts</span>
  }
  return (
    <div className="flex min-w-0 flex-wrap gap-1.5">
      {accounts.map((account) => {
        const Icon = account.provider === "GITHUB" ? GithubIcon : DiscordIcon
        const provider = account.provider === "GITHUB" ? "GitHub" : "Discord"
        const identity = account.username || account.external_id || "Linked"
        return (
          <Badge key={account.provider} variant="outline" className="max-w-full gap-1 text-xs" title={`${provider}: ${identity}`}>
            <Icon className="size-3 shrink-0" aria-hidden="true" />
            <span className="sr-only">{provider}: </span>
            <span className="truncate">{identity}</span>
          </Badge>
        )
      })}
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
        <p className="text-xs text-muted-foreground">Email</p>
        <p className="break-all text-sm">{member.email || "Not provided"}</p>
      </div>
      <div className="mt-auto min-w-0">
        <p className="mb-1 text-xs text-muted-foreground">Linked accounts</p>
        <MemberAccounts accounts={member.linked_accounts} />
      </div>
    </Link>
  )
}
