import { AppWindow, IdCard, Pencil, Settings, Users, type LucideIcon } from "lucide-react"

type ScopeMeta = {
  label: string
  description: string
  icon: LucideIcon
}

const SCOPES: Record<string, ScopeMeta> = {
  "user:read": {
    label: "Read your profile",
    description: "See your name, email, entity ID, and basic account info.",
    icon: IdCard,
  },
  "user:write": {
    label: "Update your profile",
    description: "Change your profile fields on your behalf.",
    icon: Pencil,
  },
  "groups:read": {
    label: "Read your group memberships",
    description: "See which groups you belong to and your role in each.",
    icon: Users,
  },
  "applications:read": {
    label: "Read application details",
    description: "See registered applications and their metadata.",
    icon: AppWindow,
  },
  "applications:write": {
    label: "Manage applications",
    description: "Create, update, and delete applications you own.",
    icon: Settings,
  },
}

export type ResolvedScope = ScopeMeta & { key: string; known: boolean }

export function resolveScopes(scopeString: string): ResolvedScope[] {
  return scopeString
    .split(/\s+/)
    .filter(Boolean)
    .map((key) => {
      const meta = SCOPES[key]
      if (meta) return { key, known: true, ...meta }
      return {
        key,
        known: false,
        label: key,
        description: "Unrecognized scope.",
        icon: IdCard,
      }
    })
}
