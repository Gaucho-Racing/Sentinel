// Group API shapes — mirror of core's GORM models. `*_count` fields are
// populated by PopulateGroup on the backend.

export type GroupSource = "DIRECT" | "DISCORD" | "CONDITIONAL"

export const ALL_GROUP_SOURCES: GroupSource[] = ["DIRECT", "DISCORD", "CONDITIONAL"]

export const SOURCE_LABEL: Record<GroupSource, string> = {
  DIRECT: "direct",
  DISCORD: "discord",
  CONDITIONAL: "conditional",
}

export type Group = {
  id: string
  name: string
  description: string
  allowed_sources: GroupSource[]
  created_by: string
  created_at: string
  updated_at: string
  member_count: number
  owner_count: number
  pending_count: number
}

export type GroupMember = {
  group_id: string
  entity_id: string
  source: GroupSource | ""
  added_by: string
  has_expiration: boolean
  expires_at: string
  joined_at: string
}

export type GroupOwner = {
  group_id: string
  entity_id: string
  added_by: string
  created_at: string
}

export type GroupJoinRequestStatus = "PENDING" | "APPROVED" | "REJECTED"

export type GroupJoinRequestComment = {
  id: string
  request_id: string
  entity_id: string
  comment: string
  created_at: string
}

export type GroupJoinRequest = {
  id: string
  group_id: string
  entity_id: string
  status: GroupJoinRequestStatus
  reviewed_by: string
  reviewed_at: string
  has_expiration: boolean
  expires_at: string
  created_at: string
  comments?: GroupJoinRequestComment[]
}
