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

// SAMLConfig mirrors core's model.SAMLServiceProvider — the SAML relying-party
// registration attached to an application. `entity_id` is the SP's SAML
// entityID (issuer); `acs_url` is its Assertion Consumer Service. Provide
// `metadata_xml` instead to have the IdP derive the ACS and signing cert from
// the SP's published metadata.
export type SAMLConfig = {
  application_id: string
  entity_id: string
  acs_url: string
  name_id_format: string
  certificate_pem: string
  want_authn_requests_signed: boolean
  metadata_xml: string
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
