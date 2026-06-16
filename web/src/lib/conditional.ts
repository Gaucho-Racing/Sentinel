import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"

import { api } from "@/lib/api"

// Mirror of core/model/conditional_binding.go::GroupConditionalBinding.
// Each binding is an AND-group of required Sentinel group IDs; group
// membership is OR across the bindings on the same parent. So
// `[{required_group_ids: [A, B]}, {required_group_ids: [C]}]` means:
// any entity that's in BOTH A and B, OR is in C.
export type GroupConditionalBinding = {
  id: string
  group_id: string
  required_group_ids: string[]
  created_at: string
}

export function useGroupConditionalBindings(groupID: string) {
  return useQuery({
    queryKey: ["group", groupID, "conditional-bindings"],
    queryFn: async () => {
      const res = await api.get<GroupConditionalBinding[]>(
        `/groups/${groupID}/conditional-bindings`,
      )
      return res.data
    },
    enabled: !!groupID,
  })
}

export function useAddGroupConditionalBinding(groupID: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (requiredGroupIDs: string[]) => {
      const res = await api.post<GroupConditionalBinding>(
        `/groups/${groupID}/conditional-bindings`,
        { required_group_ids: requiredGroupIDs },
      )
      return res.data
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["group", groupID, "conditional-bindings"] })
      // Memberships might have changed as part of the sweep that fires on
      // binding create. Invalidate the parent group's members + any
      // affected group's members would be too aggressive; just invalidate
      // this group's so the count refreshes when the page settles.
      qc.invalidateQueries({ queryKey: ["group", groupID, "members"] })
    },
  })
}

export function useRemoveGroupConditionalBinding(groupID: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (bindingID: string) => {
      await api.delete(`/groups/${groupID}/conditional-bindings/${bindingID}`)
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["group", groupID, "conditional-bindings"] })
      qc.invalidateQueries({ queryKey: ["group", groupID, "members"] })
    },
  })
}
