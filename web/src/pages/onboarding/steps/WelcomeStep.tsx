import { ShieldCheck } from "lucide-react"

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import type { DiscordIdentity } from "@/pages/onboarding/types"

type WelcomeStepProps = {
  identity: DiscordIdentity
}

function initials(name: string) {
  return name
    .split(" ")
    .map((part) => part[0])
    .filter(Boolean)
    .slice(0, 2)
    .join("")
    .toUpperCase()
}

export function WelcomeStep({ identity }: WelcomeStepProps) {
  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <h2 className="text-lg font-semibold tracking-tight">Welcome to Gaucho Racing</h2>
        <p className="text-sm text-muted-foreground">
          You verified through Discord. We'll set up your Sentinel account in a few short
          steps so you can sign in to all the team's tools.
        </p>
      </div>

      <div className="flex items-center gap-3 rounded-xl border border-border/60 bg-muted/30 px-4 py-3">
        <Avatar className="size-10">
          <AvatarImage src={identity.avatarUrl} alt={identity.globalName} />
          <AvatarFallback>{initials(identity.globalName)}</AvatarFallback>
        </Avatar>
        <div className="min-w-0 flex-1">
          <p className="truncate text-sm font-medium leading-none">{identity.globalName}</p>
          <p className="mt-1 truncate text-xs text-muted-foreground">@{identity.username}</p>
        </div>
        <div className="flex items-center gap-1.5 text-xs text-emerald-500">
          <ShieldCheck className="size-3.5" />
          Verified
        </div>
      </div>
    </div>
  )
}
