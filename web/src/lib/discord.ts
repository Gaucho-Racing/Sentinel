import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"

import { api } from "@/lib/api"

// Mirror of discord/api/guild.go::discordRole
export type DiscordRole = {
  id: string
  name: string
  color: number
  position: number
  hoist: boolean
  mentionable: boolean
  managed: boolean
}

export function useDiscordRoles() {
  return useQuery({
    queryKey: ["discord", "roles"],
    queryFn: async () => {
      const res = await api.get<DiscordRole[]>("/discord/roles")
      return res.data
    },
    staleTime: 5 * 60 * 1000,
  })
}

export function discordRoleColorHex(color: number): string | null {
  if (!color) return null
  return `#${color.toString(16).padStart(6, "0")}`
}

// Mirror of the future GroupDiscordRoleBinding model. Each binding is an
// AND-group of Discord role IDs; group membership is OR across the bindings.
// So `[{required: [A, B]}, {required: [C]}]` means: any user who has BOTH
// roles A and B, OR has role C.
//
// Storage is per-group in localStorage for now — swapping
// useGroupDiscordBindings/mutations to a real
// `/groups/:id/discord-bindings` endpoint later is a self-contained change.
export type GroupDiscordRoleBinding = {
  id: string
  group_id: string
  discord_role_ids: string[]
  created_at: string
}

function mockKey(groupID: string) {
  return `mock_discord_bindings_${groupID}`
}

function readMockBindings(groupID: string): GroupDiscordRoleBinding[] {
  const raw = localStorage.getItem(mockKey(groupID))
  if (!raw) return []
  try {
    const parsed = JSON.parse(raw) as unknown[]
    return parsed
      .map((entry) => {
        const e = entry as Record<string, unknown>
        // Migrate legacy single-role-per-binding shape on the fly.
        if (typeof e.discord_role_id === "string") {
          return {
            id: (e.id as string) ?? `local_${e.discord_role_id}`,
            group_id: e.group_id as string,
            discord_role_ids: [e.discord_role_id as string],
            created_at: (e.created_at as string) ?? new Date().toISOString(),
          }
        }
        if (Array.isArray(e.discord_role_ids)) {
          return e as unknown as GroupDiscordRoleBinding
        }
        return null
      })
      .filter((b): b is GroupDiscordRoleBinding => b !== null)
  } catch {
    return []
  }
}

function writeMockBindings(groupID: string, bindings: GroupDiscordRoleBinding[]) {
  localStorage.setItem(mockKey(groupID), JSON.stringify(bindings))
}

function newBindingID() {
  return `dbind_local_${Date.now().toString(36)}_${Math.random().toString(36).slice(2, 8)}`
}

export function useGroupDiscordBindings(groupID: string) {
  return useQuery({
    queryKey: ["group", groupID, "discord-bindings"],
    queryFn: () => Promise.resolve(readMockBindings(groupID)),
    enabled: !!groupID,
  })
}

export function useAddGroupDiscordBinding(groupID: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (discordRoleIDs: string[]) => {
      if (discordRoleIDs.length === 0) return readMockBindings(groupID)
      const existing = readMockBindings(groupID)
      const next: GroupDiscordRoleBinding[] = [
        ...existing,
        {
          id: newBindingID(),
          group_id: groupID,
          discord_role_ids: [...discordRoleIDs],
          created_at: new Date().toISOString(),
        },
      ]
      writeMockBindings(groupID, next)
      return next
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["group", groupID, "discord-bindings"] })
    },
  })
}

export function useRemoveGroupDiscordBinding(groupID: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (bindingID: string) => {
      const next = readMockBindings(groupID).filter((b) => b.id !== bindingID)
      writeMockBindings(groupID, next)
      return next
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["group", groupID, "discord-bindings"] })
    },
  })
}
