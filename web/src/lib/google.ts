import { useQuery } from "@tanstack/react-query"

import { api } from "./api"

// Mirror of google/model/group_binding.go::GroupGoogleBinding. A 1:1 mapping of
// a Sentinel group to the Google Group its membership is mirrored into.
export type GroupGoogleBinding = {
  id: string
  group_id: string
  google_group_email: string
  status: "active" | "pending"
  operation_id?: string
  last_sync_error?: string
  created_at: string
  updated_at: string
}

export type GoogleGroupMemberSnapshot = {
  email: string
  role: "OWNER" | "MANAGER" | "MEMBER"
}

export type GoogleGroupSnapshot = {
  requested_email: string
  id?: string
  email: string
  name?: string
  exists: boolean
  members: GoogleGroupMemberSnapshot[]
}

export type GoogleBindingConfirmations = {
  overwrite_requested_group_members: boolean
  delete_previous_group: boolean
}

export type GoogleBindingPreflight = {
  group_id: string
  requested_email: string
  binding_changed: boolean
  current_binding: GroupGoogleBinding | null
  previous_group: GoogleGroupSnapshot | null
  requested_group: GoogleGroupSnapshot | null
  required_confirmation: GoogleBindingConfirmations
}

export async function preflightGoogleBinding(
  groupID: string,
  googleGroupEmail: string,
) {
  const res = await api.post<GoogleBindingPreflight>(
    "/google/group-bindings/preflight",
    {
      group_id: groupID,
      google_group_email: googleGroupEmail,
    },
  )
  return res.data
}

export async function applyGoogleBinding(
  groupID: string,
  googleGroupEmail: string,
  expectedCurrentBindingID: string,
  confirmations: GoogleBindingConfirmations,
) {
  const res = await api.put<{
    binding: GroupGoogleBinding | null
    operation_id?: string
    status: "queued" | "unchanged"
  }>(
    "/google/group-bindings",
    {
      group_id: groupID,
      google_group_email: googleGroupEmail,
      expected_current_binding_id: expectedCurrentBindingID,
      confirm_overwrite_requested_group_members:
        confirmations.overwrite_requested_group_members,
      confirm_delete_previous_group: confirmations.delete_previous_group,
    },
  )
  return res.data
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
