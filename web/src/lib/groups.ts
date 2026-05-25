// Group surface shape — mock for now until backend wiring lands. Types here
// will become the on-the-wire shape from /groups, /groups/:id, etc.

export type GroupSource = "DIRECT" | "DISCORD" | "CONDITIONAL"

export type GroupSummary = {
  id: string
  name: string
  description: string
  member_count: number
  owner_count: number
  allowed_sources: GroupSource[]
  pending_requests: number
  created_by: string
}

export type MockMember = {
  entity_id: string
  display_name: string
  username: string
  source: GroupSource
  added_by_name: string | null
  joined_at: string
}

export type MockOwner = {
  entity_id: string
  display_name: string
  username: string
  added_at: string
}

export type MockJoinRequest = {
  id: string
  requester_name: string
  requester_username: string
  reason: string
  comment_count: number
  created_at: string
}

export type MockLinkedApp = {
  id: string
  name: string
  client_id: string
}

export type MockSyncConfig =
  | { source: "DIRECT" }
  | { source: "DISCORD"; discord_role_name: string; discord_role_color: number }
  | { source: "CONDITIONAL"; rule_summary: string }

export const SOURCE_LABEL: Record<GroupSource, string> = {
  DIRECT: "direct",
  DISCORD: "discord",
  CONDITIONAL: "conditional",
}

export const MOCK_GROUPS: GroupSummary[] = [
  {
    id: "grp_officers",
    name: "Officers",
    description: "Team leadership — President, VPs, Captains, and Safety.",
    member_count: 8,
    owner_count: 3,
    allowed_sources: ["DIRECT"],
    pending_requests: 0,
    created_by: "ent_002",
  },
  {
    id: "grp_suspension",
    name: "Suspension Team",
    description: "Geometry, dampers, kinematics. Synced with the Suspension Discord role.",
    member_count: 12,
    owner_count: 2,
    allowed_sources: ["DIRECT", "DISCORD"],
    pending_requests: 1,
    created_by: "ent_002",
  },
  {
    id: "grp_aero",
    name: "Aero Team",
    description: "CFD, bodywork, undertray. Synced with the Aero Discord role.",
    member_count: 15,
    owner_count: 2,
    allowed_sources: ["DIRECT", "DISCORD"],
    pending_requests: 3,
    created_by: "ent_002",
  },
  {
    id: "grp_powertrain",
    name: "Powertrain Team",
    description: "Engine, drivetrain, fuel system. Synced with the Powertrain Discord role.",
    member_count: 10,
    owner_count: 2,
    allowed_sources: ["DIRECT", "DISCORD"],
    pending_requests: 0,
    created_by: "ent_002",
  },
  {
    id: "grp_electrical",
    name: "Electrical Team",
    description: "Wiring harness, sensors, data acquisition.",
    member_count: 9,
    owner_count: 2,
    allowed_sources: ["DIRECT", "DISCORD"],
    pending_requests: 2,
    created_by: "ent_003",
  },
  {
    id: "grp_all_members",
    name: "All Members",
    description: "Auto-populated from active membership — every onboarded member.",
    member_count: 89,
    owner_count: 1,
    allowed_sources: ["CONDITIONAL"],
    pending_requests: 0,
    created_by: "ent_002",
  },
  {
    id: "grp_alumni",
    name: "Alumni",
    description: "Past members. Auto-populated based on graduation year.",
    member_count: 47,
    owner_count: 1,
    allowed_sources: ["CONDITIONAL"],
    pending_requests: 0,
    created_by: "ent_002",
  },
  {
    id: "grp_mentors",
    name: "Mentors",
    description: "Industry mentors and alumni advisors with project oversight.",
    member_count: 6,
    owner_count: 1,
    allowed_sources: ["DIRECT"],
    pending_requests: 2,
    created_by: "ent_003",
  },
  {
    id: "grp_faculty",
    name: "Faculty Advisors",
    description: "UCSB faculty with formal advisory roles.",
    member_count: 3,
    owner_count: 1,
    allowed_sources: ["DIRECT"],
    pending_requests: 0,
    created_by: "ent_002",
  },
]

export const MOCK_MEMBERS: MockMember[] = [
  {
    entity_id: "ent_001",
    display_name: "Alex Chen",
    username: "achen",
    source: "DIRECT",
    added_by_name: "Priya Patel",
    joined_at: "2025-09-12T18:32:00Z",
  },
  {
    entity_id: "ent_002",
    display_name: "Priya Patel",
    username: "ppatel",
    source: "DIRECT",
    added_by_name: "Marcus Johnson",
    joined_at: "2025-08-04T14:11:00Z",
  },
  {
    entity_id: "ent_003",
    display_name: "Marcus Johnson",
    username: "mjohnson",
    source: "DISCORD",
    added_by_name: null,
    joined_at: "2025-09-22T03:08:00Z",
  },
  {
    entity_id: "ent_004",
    display_name: "Sofia Rodriguez",
    username: "srodriguez",
    source: "DISCORD",
    added_by_name: null,
    joined_at: "2025-10-01T20:45:00Z",
  },
  {
    entity_id: "ent_005",
    display_name: "Liam Kim",
    username: "lkim",
    source: "DIRECT",
    added_by_name: "Priya Patel",
    joined_at: "2025-10-14T16:20:00Z",
  },
  {
    entity_id: "ent_006",
    display_name: "Emily Zhang",
    username: "ezhang",
    source: "DISCORD",
    added_by_name: null,
    joined_at: "2026-01-09T11:55:00Z",
  },
  {
    entity_id: "ent_007",
    display_name: "Jordan Williams",
    username: "jwilliams",
    source: "DISCORD",
    added_by_name: null,
    joined_at: "2026-02-18T09:01:00Z",
  },
  {
    entity_id: "ent_008",
    display_name: "Aisha Khan",
    username: "akhan",
    source: "DIRECT",
    added_by_name: "Marcus Johnson",
    joined_at: "2026-03-05T22:14:00Z",
  },
]

export const MOCK_OWNERS: MockOwner[] = [
  {
    entity_id: "ent_002",
    display_name: "Priya Patel",
    username: "ppatel",
    added_at: "2025-08-04T14:11:00Z",
  },
  {
    entity_id: "ent_003",
    display_name: "Marcus Johnson",
    username: "mjohnson",
    added_at: "2025-08-04T14:11:00Z",
  },
]

export const MOCK_JOIN_REQUESTS: MockJoinRequest[] = [
  {
    id: "grp_req_01",
    requester_name: "Ryan O'Connor",
    requester_username: "roconnor",
    reason: "Joining the suspension subteam this quarter — VP Wilson asked me to request access.",
    comment_count: 2,
    created_at: "2026-05-22T17:43:00Z",
  },
  {
    id: "grp_req_02",
    requester_name: "Maya Singh",
    requester_username: "msingh",
    reason: "Working on the damper test rig — need access to the team Drive and Trackside.",
    comment_count: 0,
    created_at: "2026-05-24T08:12:00Z",
  },
]

export const MOCK_LINKED_APPS: MockLinkedApp[] = [
  { id: "app_trackside", name: "Trackside", client_id: "trackside" },
  { id: "app_telemetry", name: "Telemetry Vault", client_id: "telemetry-vault" },
  { id: "app_parts", name: "Parts Hub", client_id: "parts-hub" },
]

export function getMockGroup(id: string | undefined): GroupSummary | undefined {
  if (!id) return undefined
  return MOCK_GROUPS.find((g) => g.id === id)
}

export function getMockPerson(entityID: string): { display_name: string; username: string } | undefined {
  const member = MOCK_MEMBERS.find((m) => m.entity_id === entityID)
  if (member) return { display_name: member.display_name, username: member.username }
  const owner = MOCK_OWNERS.find((o) => o.entity_id === entityID)
  if (owner) return { display_name: owner.display_name, username: owner.username }
  return undefined
}

export function syncConfigsFor(group: GroupSummary): MockSyncConfig[] {
  const configs: MockSyncConfig[] = []
  if (group.allowed_sources.includes("DIRECT")) {
    configs.push({ source: "DIRECT" })
  }
  if (group.allowed_sources.includes("DISCORD")) {
    configs.push({
      source: "DISCORD",
      discord_role_name: group.name,
      discord_role_color: 0x8412fc,
    })
  }
  if (group.allowed_sources.includes("CONDITIONAL")) {
    const ruleSummary =
      group.id === "grp_alumni"
        ? "graduation_year < current_year AND onboarded = true"
        : group.id === "grp_all_members"
          ? "onboarded = true AND initial_role IN ('member', 'officer')"
          : "custom rule defined for this group"
    configs.push({ source: "CONDITIONAL", rule_summary: ruleSummary })
  }
  return configs
}
