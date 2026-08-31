export type SCIMSyncStatus = "NEVER" | "RUNNING" | "SUCCEEDED" | "FAILED"

export type SCIMConfiguration = {
  application_id: string
  endpoint: string
  token_configured: boolean
  token_expires_at: string | null
  enabled: boolean
  last_sync_at: string | null
  last_sync_status: SCIMSyncStatus
  last_sync_error: string
  updated_at: string
  created_at: string
}

export type SCIMSyncResult = {
  dry_run: boolean
  desired_users: number
  desired_groups: number
  users_created: number
  users_updated: number
  users_deactivated: number
  groups_created: number
  groups_updated: number
  memberships_added: number
  memberships_removed: number
  validation_errors: string[]
  completed_at: string
}
