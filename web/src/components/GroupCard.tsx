import { Crown, Users } from "lucide-react"
import { Link } from "react-router-dom"

import { Badge } from "@/components/ui/badge"
import { SOURCE_LABEL, type Group } from "@/lib/groups"

function initial(name: string) {
  return name.slice(0, 1).toUpperCase()
}

export function GroupCard({ group }: { group: Group }) {
  return (
    <Link
      to={`/groups/${group.id}`}
      className="group flex flex-col gap-3 rounded-lg border border-border/60 bg-card p-4 transition-colors hover:bg-muted/40"
    >
      <div className="flex items-start justify-between gap-2">
        <div className="flex size-10 items-center justify-center overflow-hidden rounded-md bg-gradient-to-br from-gr-pink to-gr-purple text-base font-semibold text-white">
          {initial(group.name)}
        </div>
        {group.pending_count > 0 && (
          <Badge
            variant="outline"
            className="border-gr-pink/40 bg-gr-pink/10 text-gr-pink"
          >
            {group.pending_count} pending
          </Badge>
        )}
      </div>
      <div>
        <p className="text-sm font-medium leading-none">{group.name}</p>
        <p className="mt-1 line-clamp-2 text-xs text-muted-foreground">
          {group.description || "No description."}
        </p>
      </div>
      <div className="mt-auto flex items-center justify-between gap-2">
        <div className="flex items-center gap-3 text-xs text-muted-foreground">
          <span className="flex items-center gap-1">
            <Users className="size-3" />
            {group.member_count}
          </span>
          <span className="flex items-center gap-1">
            <Crown className="size-3" />
            {group.owner_count}
          </span>
        </div>
        <div className="flex flex-wrap items-center gap-1">
          {group.allowed_sources?.map((source) => (
            <Badge key={source} variant="outline" className="font-mono text-[10px]">
              {SOURCE_LABEL[source]}
            </Badge>
          ))}
        </div>
      </div>
    </Link>
  )
}
