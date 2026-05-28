import { useQuery } from "@tanstack/react-query"

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Skeleton } from "@/components/ui/skeleton"
import { api } from "@/lib/api"
import type { Entity } from "@/lib/auth"

type EntityChipSize = "sm" | "md"

const AVATAR_CLASS: Record<EntityChipSize, string> = {
  sm: "size-6",
  md: "size-8",
}

const FALLBACK_TEXT_CLASS: Record<EntityChipSize, string> = {
  sm: "text-[10px]",
  md: "text-xs",
}

export function entityInitials(name: string) {
  return name
    .split(/\s+/)
    .map((p) => p[0])
    .filter(Boolean)
    .slice(0, 2)
    .join("")
    .toUpperCase()
}

export type ResolvedPerson = {
  name: string
  username: string | null
  avatarUrl?: string
  isServiceAccount: boolean
}

export function useEntity(entityId: string) {
  return useQuery({
    queryKey: ["entity", entityId],
    queryFn: async () => {
      const res = await api.get<Entity>(`/entities/${entityId}`)
      return res.data
    },
    enabled: !!entityId,
    staleTime: 5 * 60 * 1000,
  })
}

export function resolveEntity(entity: Entity): ResolvedPerson | null {
  if (entity.type === "USER" && entity.user) {
    const full = `${entity.user.first_name} ${entity.user.last_name}`.trim()
    return {
      name: full || entity.user.username || entity.id,
      username: entity.user.username || null,
      avatarUrl: entity.user.avatar_url,
      isServiceAccount: false,
    }
  }
  if (entity.type === "SERVICE_ACCOUNT" && entity.service_account) {
    return {
      name: entity.service_account.name || entity.id,
      username: null,
      isServiceAccount: true,
    }
  }
  return null
}

export function EntityChip({
  entityId,
  size = "md",
}: {
  entityId: string
  size?: EntityChipSize
}) {
  const query = useEntity(entityId)

  if (!entityId) {
    return <span className="text-sm text-muted-foreground">—</span>
  }

  if (query.isLoading) {
    return (
      <div className="flex items-center gap-2.5">
        <Skeleton className={AVATAR_CLASS[size] + " rounded-full"} />
        <div className="space-y-1">
          <Skeleton className="h-3.5 w-24" />
          <Skeleton className="h-3 w-16" />
        </div>
      </div>
    )
  }

  if (!query.data) {
    return <code className="font-mono text-xs text-muted-foreground">{entityId}</code>
  }

  const person = resolveEntity(query.data)
  if (!person) {
    return <code className="font-mono text-xs text-muted-foreground">{entityId}</code>
  }

  return (
    <div className="flex min-w-0 items-center gap-2.5">
      <Avatar className={AVATAR_CLASS[size]}>
        {person.avatarUrl && <AvatarImage src={person.avatarUrl} alt={person.name} />}
        <AvatarFallback
          className={
            FALLBACK_TEXT_CLASS[size] +
            (person.isServiceAccount
              ? " bg-gradient-to-br from-gr-pink to-gr-purple text-white"
              : "")
          }
        >
          {entityInitials(person.name) || person.name.slice(0, 1).toUpperCase()}
        </AvatarFallback>
      </Avatar>
      <div className="flex min-w-0 flex-col leading-tight">
        <span className="truncate text-sm">{person.name}</span>
        {person.username ? (
          <span className="truncate text-xs text-muted-foreground">@{person.username}</span>
        ) : person.isServiceAccount ? (
          <span className="truncate text-xs text-muted-foreground">Service account</span>
        ) : null}
      </div>
    </div>
  )
}
