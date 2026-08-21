import { useQuery } from "@tanstack/react-query"

import { api } from "@/lib/api"

// Types mirror the JSON shapes returned by the core /analytics/* endpoints.

export type AnalyticsOverview = {
  total_users: number
  total_service_accounts: number
  total_applications: number
  total_groups: number
  new_users_30d: number
  logins_24h: number
  logins_7d: number
  logins_30d: number
  active_users_7d: number
  active_users_30d: number
  pending_join_requests: number
  audit_events_7d: number
}

export type LoginPoint = {
  date: string
  logins: number
  unique_users: number
}

export type HeatmapCell = {
  weekday: number
  hour: number
  count: number
}

export type TopApplication = {
  client_id: string
  name: string
  icon_url: string
  logins: number
  unique_users: number
}

export type UserGrowthPoint = {
  date: string
  new_users: number
  cumulative: number
}

export type CategoryCount = {
  label: string
  count: number
}

export type MemberDemographics = {
  by_grad_year: CategoryCount[]
  by_major: CategoryCount[]
  by_graduate_level: CategoryCount[]
}

export type AuthMethodBreakdown = {
  email: number
  phone: number
  discord: number
  google: number
  github: number
}

export type GroupMembership = {
  group_id: string
  name: string
  member_count: number
  direct: number
  conditional: number
  discord: number
}

export type JoinRequestFunnel = {
  pending: number
  approved: number
  rejected: number
  median_decision_hours: number
}

export type AuditEvent = {
  id: string
  actor_id: string
  action: string
  target_type: string
  target_id: string
  ip_address: string
  metadata: Record<string, unknown> | null
  created_at: string
}

const STALE = 5 * 60 * 1000

export function useAnalyticsOverview(enabled = true) {
  return useQuery({
    queryKey: ["analytics", "overview"],
    queryFn: async () => (await api.get<AnalyticsOverview>("/analytics/overview")).data,
    staleTime: STALE,
    enabled,
  })
}

export function useLoginTimeSeries(days: number, enabled = true) {
  return useQuery({
    queryKey: ["analytics", "logins", "timeseries", days],
    queryFn: async () =>
      (await api.get<LoginPoint[]>("/analytics/logins/timeseries", { params: { days } })).data,
    staleTime: STALE,
    enabled,
  })
}

export function useLoginHeatmap(days: number, enabled = true) {
  return useQuery({
    queryKey: ["analytics", "logins", "heatmap", days],
    queryFn: async () =>
      (await api.get<HeatmapCell[]>("/analytics/logins/heatmap", { params: { days } })).data,
    staleTime: STALE,
    enabled,
  })
}

export function useTopApplications(days: number, limit = 10, enabled = true) {
  return useQuery({
    queryKey: ["analytics", "applications", "top", days, limit],
    queryFn: async () =>
      (await api.get<TopApplication[]>("/analytics/applications/top", { params: { days, limit } }))
        .data,
    staleTime: STALE,
    enabled,
  })
}

export function useUserGrowth(months: number, enabled = true) {
  return useQuery({
    queryKey: ["analytics", "users", "growth", months],
    queryFn: async () =>
      (await api.get<UserGrowthPoint[]>("/analytics/users/growth", { params: { months } })).data,
    staleTime: STALE,
    enabled,
  })
}

export function useMemberDemographics(enabled = true) {
  return useQuery({
    queryKey: ["analytics", "members", "demographics"],
    queryFn: async () =>
      (await api.get<MemberDemographics>("/analytics/members/demographics")).data,
    staleTime: STALE,
    enabled,
  })
}

export function useAuthMethods(enabled = true) {
  return useQuery({
    queryKey: ["analytics", "auth-methods"],
    queryFn: async () => (await api.get<AuthMethodBreakdown>("/analytics/auth-methods")).data,
    staleTime: STALE,
    enabled,
  })
}

export function useGroupMembership(enabled = true) {
  return useQuery({
    queryKey: ["analytics", "groups", "membership"],
    queryFn: async () =>
      (await api.get<GroupMembership[]>("/analytics/groups/membership")).data,
    staleTime: STALE,
    enabled,
  })
}

export function useJoinRequestFunnel(days: number, enabled = true) {
  return useQuery({
    queryKey: ["analytics", "groups", "join-requests", days],
    queryFn: async () =>
      (await api.get<JoinRequestFunnel>("/analytics/groups/join-requests", { params: { days } }))
        .data,
    staleTime: STALE,
    enabled,
  })
}

export function useAuditEvents(
  filters: { action?: string; limit?: number } = {},
  enabled = true,
) {
  return useQuery({
    queryKey: ["analytics", "audit", filters],
    queryFn: async () =>
      (await api.get<AuditEvent[]>("/analytics/audit", { params: filters })).data,
    staleTime: 60 * 1000,
    enabled,
  })
}

export function useAuditSummary(days: number, enabled = true) {
  return useQuery({
    queryKey: ["analytics", "audit", "summary", days],
    queryFn: async () =>
      (await api.get<CategoryCount[]>("/analytics/audit/summary", { params: { days } })).data,
    staleTime: STALE,
    enabled,
  })
}
