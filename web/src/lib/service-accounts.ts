import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"

import { api } from "@/lib/api"
import type { Group } from "@/lib/groups"

// Mirror of core/model/service_account.go::ServiceAccount.
export type ServiceAccount = {
  id: string
  entity_id: string
  application_id: string
  name: string
  created_by: string
  groups: Group[]
  created_at: string
}

// Mirror of core/model/api_key.go::APIKey. Note: `hashed_secret` is
// json:"-" on the backend, so it never appears on the wire.
export type APIKey = {
  id: string
  service_account_id: string
  name: string
  key_id: string
  scope: string
  // null = never expires
  expires_at: string | null
  last_used_at: string | null
  created_at: string
  created_by: string
}

// Returned ONCE on POST /service-accounts/:id/api-keys — `token` is the
// raw sk_<key_id>_<secret> string. Subsequent GETs of the key never
// include it; the user has to copy this on creation or revoke and remint.
export type APIKeyWithToken = {
  key: APIKey
  token: string
}

export function useApplicationServiceAccounts(applicationID: string) {
  return useQuery({
    queryKey: ["application", applicationID, "service-accounts"],
    queryFn: async () => {
      const res = await api.get<ServiceAccount[]>(
        `/applications/${applicationID}/service-accounts`,
      )
      return res.data
    },
    enabled: !!applicationID,
  })
}

export function useCreateServiceAccount(applicationID: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (name: string) => {
      const res = await api.post<ServiceAccount>(
        `/applications/${applicationID}/service-accounts`,
        { name },
      )
      return res.data
    },
    onSuccess: () => {
      qc.invalidateQueries({
        queryKey: ["application", applicationID, "service-accounts"],
      })
    },
  })
}

export function useDeleteServiceAccount(applicationID: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (saID: string) => {
      await api.delete(`/service-accounts/${saID}`)
    },
    onSuccess: () => {
      qc.invalidateQueries({
        queryKey: ["application", applicationID, "service-accounts"],
      })
    },
  })
}

export function useServiceAccountAPIKeys(saID: string) {
  return useQuery({
    queryKey: ["service-account", saID, "api-keys"],
    queryFn: async () => {
      const res = await api.get<APIKey[]>(`/service-accounts/${saID}/api-keys`)
      return res.data
    },
    enabled: !!saID,
  })
}

export type CreateAPIKeyInput = {
  name: string
  // 0 = never expires
  ttl_days: number
  scope: string
}

export function useCreateAPIKey(saID: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (input: CreateAPIKeyInput) => {
      const res = await api.post<APIKeyWithToken>(
        `/service-accounts/${saID}/api-keys`,
        input,
      )
      return res.data
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["service-account", saID, "api-keys"] })
    },
  })
}

export function useRevokeAPIKey(saID: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (keyID: string) => {
      await api.delete(`/service-accounts/${saID}/api-keys/${keyID}`)
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["service-account", saID, "api-keys"] })
    },
  })
}

// TTL_PRESETS is the dropdown shown on the create-key dialog. Values are
// days; 0 maps to "never expires" on the backend.
export const TTL_PRESETS: { label: string; days: number }[] = [
  { label: "30 days", days: 30 },
  { label: "90 days", days: 90 },
  { label: "365 days", days: 365 },
  { label: "Never", days: 0 },
]
