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

// Substitutions chosen to demonstrate that `*` is greedy and matches dots and
// slashes — the two characters that make wildcard redirect URIs dangerous
// (host confusion, path takeover). Order: innocuous → concerning.
const WILDCARD_EXAMPLE_SUBSTITUTIONS = ["app", "beta.staging", "evil.com/path"]

export function redirectURIWildcardExamples(pattern: string): string[] {
  if (!pattern.includes("*")) return []
  return WILDCARD_EXAMPLE_SUBSTITUTIONS.map((sub) => pattern.replaceAll("*", sub))
}
