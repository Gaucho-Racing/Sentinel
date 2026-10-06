import { useQuery } from "@tanstack/react-query"

import { api } from "@/lib/api"

import type { Group } from "./groups"

// Application API shape — mirror of core's model.Application JSON.
// owner_id is the entity_id of the creator (USER or SERVICE_ACCOUNT entity).
export type Application = {
  id: string
  owner_id: string
  name: string
  description: string
  client_id: string
  icon_url: string
  launch_url: string
  redirect_uris: string[]
  updated_at: string
  created_at: string
}

// GroupWithLink is what `GET /applications/:id/groups` returns — a Group
// enriched with the `required` flag from its application_group link.
// `required` gates OAuth access: if any linked group on the app has it set
// true, the user must be in at least one of those required groups to obtain
// a token. Non-required links still flow into the token's groups claim.
export type GroupWithLink = Group & { required: boolean }

// ApplicationWithLink is the inverse — what `GET /groups/:id/applications`
// returns. Application + the link's `required` flag inline.
export type ApplicationWithLink = Application & { required: boolean }

// Mirror of core's model.ApplicationGroup — the raw link row, without the
// joined group/application records. `GET /applications/groups` returns the
// whole table so a consumer holding both lists can resolve the names itself.
export type ApplicationGroupLink = {
  application_id: string
  group_id: string
  required: boolean
  created_at: string
}

export function useAllApplicationGroupLinks(enabled = true) {
  return useQuery({
    queryKey: ["application-group-links"],
    queryFn: async () => {
      const res = await api.get<ApplicationGroupLink[]>("/applications/groups")
      return res.data
    },
    enabled,
  })
}

export function useApplications(enabled = true) {
  return useQuery({
    queryKey: ["applications"],
    queryFn: async () => {
      const res = await api.get<Application[]>("/applications")
      return res.data
    },
    enabled,
  })
}

// Substitutions chosen to demonstrate that `*` is greedy and matches dots and
// slashes — the two characters that make wildcard redirect URIs dangerous
// (host confusion, path takeover). Order: innocuous → concerning.
const WILDCARD_EXAMPLE_SUBSTITUTIONS = ["app", "beta.staging", "evil.com/path"]

export function redirectURIWildcardExamples(pattern: string): string[] {
  if (!pattern.includes("*")) return []
  return WILDCARD_EXAMPLE_SUBSTITUTIONS.map((sub) => pattern.replaceAll("*", sub))
}

// Team convention: an application whose name ends in `-dev` is a development
// or local-testing client rather than something people sign into day to day.
// There is no flag on the record for this, so the name is the only signal.
const DEV_APPLICATION_SUFFIX = "-dev"

export function isDevApplication(app: Application): boolean {
  return app.name.trim().toLowerCase().endsWith(DEV_APPLICATION_SUFFIX)
}
