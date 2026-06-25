import { useQuery } from "@tanstack/react-query"

import { api } from "./api"

// Mirror of google/model/group_binding.go::GroupGoogleBinding. A 1:1 mapping of
// a Sentinel group to the Google Group its membership is mirrored into.
export type GroupGoogleBinding = {
  id: string
  group_id: string
  google_group_email: string
  created_at: string
}

// useGroupGoogleBinding returns the single binding for a group, or null. The
// list endpoint returns an array (0 or 1 rows) since the mapping is 1:1.
export function useGroupGoogleBinding(groupID: string) {
  return useQuery({
    queryKey: ["group", groupID, "google-binding"],
    queryFn: async () => {
      const res = await api.get<GroupGoogleBinding[]>(`/google/group-bindings`, {
        params: { group_id: groupID },
      })
      return res.data[0] ?? null
    },
    enabled: !!groupID,
  })
}
