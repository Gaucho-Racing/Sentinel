import { GraduationCap, ShieldCheck, UserPlus, Users } from "lucide-react"

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { cn } from "@/lib/utils"
import type { DiscordIdentity, OnboardingRole, StepProps } from "@/pages/onboarding/types"

type WelcomeStepProps = StepProps & {
  identity: DiscordIdentity
}

const ROLE_OPTIONS: {
  value: OnboardingRole
  label: string
  description: string
  Icon: typeof Users
}[] = [
  {
    value: "member",
    label: "Current member",
    description: "Active student on the team",
    Icon: Users,
  },
  {
    value: "alumni",
    label: "Alumni",
    description: "Graduated from the team",
    Icon: GraduationCap,
  },
  {
    value: "guest",
    label: "Guest",
    description: "Mentor, sponsor, or other",
    Icon: UserPlus,
  },
]

function initials(name: string) {
  return name
    .split(" ")
    .map((part) => part[0])
    .filter(Boolean)
    .slice(0, 2)
    .join("")
    .toUpperCase()
}

export function WelcomeStep({ identity, data, update }: WelcomeStepProps) {
  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <h2 className="text-lg font-semibold tracking-tight">Welcome to Gaucho Racing</h2>
        <p className="text-sm text-muted-foreground">
          You verified through Discord. Tell us how you'll be joining the team so we can
          tailor the rest of setup.
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

      <div className="space-y-2">
        <p className="text-sm font-medium">I'm joining as a…</p>
        <div className="grid gap-2">
          {ROLE_OPTIONS.map(({ value, label, description, Icon }) => {
            const selected = data.role === value
            return (
              <button
                key={value}
                type="button"
                onClick={() => update({ role: value })}
                aria-pressed={selected}
                className={cn(
                  "flex items-center gap-3 rounded-xl border px-4 py-3 text-left transition-colors",
                  selected
                    ? "border-gr-pink/60 bg-gr-pink/5"
                    : "border-border/60 hover:border-foreground/60 hover:bg-muted/40",
                )}
              >
                <div
                  className={cn(
                    "flex size-9 items-center justify-center rounded-lg",
                    selected
                      ? "bg-gradient-to-br from-gr-pink to-gr-purple text-white"
                      : "bg-muted text-muted-foreground",
                  )}
                >
                  <Icon className="size-4" />
                </div>
                <div className="min-w-0 flex-1">
                  <p className="text-sm font-medium leading-none">{label}</p>
                  <p className="mt-1 text-xs text-muted-foreground">{description}</p>
                </div>
              </button>
            )
          })}
        </div>
      </div>
    </div>
  )
}
