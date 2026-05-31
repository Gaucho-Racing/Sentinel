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

// Mirror of discord/model/group_binding.go::GroupDiscordRoleBinding. Each
// binding is an AND-group of Discord role IDs; group membership is OR across
// the bindings. So `[{discord_role_ids: [A, B]}, {discord_role_ids: [C]}]`
// means: any user who has BOTH roles A and B, OR has role C.
export type GroupDiscordRoleBinding = {
  id: string
  group_id: string
  discord_role_ids: string[]
  created_at: string
}

export function useGroupDiscordBindings(groupID: string) {
  return useQuery({
    queryKey: ["group", groupID, "discord-bindings"],
    queryFn: async () => {
      const res = await api.get<GroupDiscordRoleBinding[]>(
        `/discord/role-bindings`,
        { params: { group_id: groupID } },
      )
      return res.data
    },
    enabled: !!groupID,
  })
}

export function useAddGroupDiscordBinding(groupID: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (discordRoleIDs: string[]) => {
      const res = await api.post<GroupDiscordRoleBinding>(
        `/discord/role-bindings`,
        { group_id: groupID, discord_role_ids: discordRoleIDs },
      )
      return res.data
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
      await api.delete(`/discord/role-bindings/${bindingID}`, {
        params: { group_id: groupID },
      })
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["group", groupID, "discord-bindings"] })
    },
  })
}
