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

// Mirror of the future GroupDiscordRoleBinding model. While the backend
// isn't built yet the storage is per-group in localStorage so the UX is
// fully exercisable end-to-end; swapping useGroupDiscordBindings/mutations
// to a real `/groups/:id/discord-bindings` endpoint is a self-contained
// change.
export type GroupDiscordRoleBinding = {
  group_id: string
  discord_role_id: string
  created_at: string
}

function mockKey(groupID: string) {
  return `mock_discord_bindings_${groupID}`
}

function readMockBindings(groupID: string): GroupDiscordRoleBinding[] {
  const raw = localStorage.getItem(mockKey(groupID))
  if (!raw) return []
  try {
    return JSON.parse(raw) as GroupDiscordRoleBinding[]
  } catch {
    return []
  }
}

function writeMockBindings(groupID: string, bindings: GroupDiscordRoleBinding[]) {
  localStorage.setItem(mockKey(groupID), JSON.stringify(bindings))
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
    mutationFn: async (discordRoleID: string) => {
      const existing = readMockBindings(groupID)
      if (existing.some((b) => b.discord_role_id === discordRoleID)) return existing
      const next = [
        ...existing,
        {
          group_id: groupID,
          discord_role_id: discordRoleID,
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
    mutationFn: async (discordRoleID: string) => {
      const next = readMockBindings(groupID).filter(
        (b) => b.discord_role_id !== discordRoleID,
      )
      writeMockBindings(groupID, next)
      return next
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["group", groupID, "discord-bindings"] })
    },
  })
}
