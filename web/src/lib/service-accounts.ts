import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"

import { api } from "@/lib/api"
import type { Group } from "@/lib/groups"

// Mirror of core/model/jwt.go::Token — the auth_token row.
export type Token = {
  id: string
  entity_id: string
  client_id: string
  scope: string
  expires_at: string
  created_at: string
}

// Mirror of core/model/service_account.go::ServiceAccount with the
// PopulateServiceAccount-injected fields (groups, active_token).
export type ServiceAccount = {
  id: string
  entity_id: string
  application_id: string
  name: string
  scope: string
  ttl_days: number
  created_by: string
  groups: Group[]
  // Null when the SA has no active token row (revoked, never minted,
  // or was deleted out-of-band).
  active_token: Token | null
  created_at: string
}

// Returned on POST /applications/:id/service-accounts and
// POST /service-accounts/:id/rotate. The `token` field is the raw JWT
// — exposed ONCE on create + rotate and never on subsequent reads.
export type ServiceAccountWithToken = {
  service_account: ServiceAccount
  token: string
}

// The scopes the backend's ValidateServiceAccountScope accepts. Kept
// in sync with core/service/service_account.go::ServiceAccountAllowedScopes.
// Read-only by design — *:write scopes route through human-authed flows.
export const SA_ALLOWED_SCOPES = [
  "user:read",
  "groups:read",
  "applications:read",
] as const

export type SAScope = (typeof SA_ALLOWED_SCOPES)[number]

export const SA_SCOPE_DESCRIPTIONS: Record<SAScope, string> = {
  "user:read": "Read user and entity profiles",
  "groups:read": "Read group memberships",
  "applications:read": "Read application details",
}

// TTL_PRESETS is the dropdown shown on the create / rotate dialogs.
// Values are days; 0 = never expires (issued with a ~100-year JWT exp).
export const TTL_PRESETS: { label: string; days: number }[] = [
  { label: "30 days", days: 30 },
  { label: "90 days", days: 90 },
  { label: "365 days", days: 365 },
  { label: "Never", days: 0 },
]

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

export type CreateServiceAccountInput = {
  name: string
  scope: string
  ttl_days: number
}

export function useCreateServiceAccount(applicationID: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (input: CreateServiceAccountInput) => {
      const res = await api.post<ServiceAccountWithToken>(
        `/applications/${applicationID}/service-accounts`,
        input,
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

export function useRotateServiceAccountToken(applicationID: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (saID: string) => {
      const res = await api.post<ServiceAccountWithToken>(
        `/service-accounts/${saID}/rotate`,
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

// isNeverExpires reports whether an expires_at date is far enough in
// the future to render as "Never" rather than a literal date. SA tokens
// with ttl_days=0 are issued with a ~100-year exp; anything past 50
// years from now is effectively non-expiring.
export function isNeverExpires(expiresAt: string | null | undefined): boolean {
  if (!expiresAt) return false
  const t = new Date(expiresAt).getTime()
  const fiftyYears = 50 * 365 * 24 * 60 * 60 * 1000
  return t - Date.now() > fiftyYears
}
