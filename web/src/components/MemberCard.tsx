import { Link } from "react-router-dom"

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Badge } from "@/components/ui/badge"
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

export function MemberCard({ member, groups }: { member: Member; groups: string[] }) {
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

      <div className="mt-auto flex flex-wrap items-center gap-1">
        {groups.slice(0, 3).map((name) => (
          <Badge key={name} variant="outline" className="h-5 max-w-full text-[10px]">
            <span className="truncate">{name}</span>
          </Badge>
        ))}
        {groups.length > 3 && (
          <span className="text-[10px] text-muted-foreground">+{groups.length - 3}</span>
        )}
      </div>
    </Link>
  )
}
