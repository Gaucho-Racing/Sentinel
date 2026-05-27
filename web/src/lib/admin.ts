import { useQuery } from "@tanstack/react-query"

import { api } from "@/lib/api"
import { loadSession } from "@/lib/auth"
import type { GroupMember } from "@/lib/groups"

// Fixed ID of the global Admins group on the backend (service.AdminsGroupID).
export const ADMINS_GROUP_ID = "grp_01kqs3w6h82xkdnft94vpj7qrm"

// Returns the list of entity IDs in the Admins group plus a helper boolean
// for the current session. Members are cached for 5 minutes since admin
// composition changes rarely.
export function useAdmins() {
  const session = loadSession()
  const entityId = session?.entityId

  const query = useQuery({
    queryKey: ["admins"],
    queryFn: async () => {
      const res = await api.get<GroupMember[]>(`/groups/${ADMINS_GROUP_ID}/members`)
      return res.data
    },
    staleTime: 5 * 60 * 1000,
  })

  const adminIds = new Set((query.data ?? []).map((m) => m.entity_id))
  const isAdmin = !!entityId && adminIds.has(entityId)

  return {
    adminIds,
    isAdmin,
    isLoading: query.isLoading,
  }
}
