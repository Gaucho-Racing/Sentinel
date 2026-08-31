export type SCIMSyncStatus = "NEVER" | "QUEUED" | "RUNNING" | "SUCCEEDED" | "FAILED"
export type SCIMSyncInterval = "5m" | "15m" | "30m" | "1h" | "6h" | "24h"
export type SCIMSyncRunStatus = "QUEUED" | "RUNNING" | "SUCCEEDED" | "FAILED"

export type SCIMConfiguration = {
  application_id: string
  endpoint: string
  token_configured: boolean
  token_expires_at: string | null
  enabled: boolean
  sync_interval: SCIMSyncInterval
  next_sync_at: string | null
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
  skipped_users: {
    entity_id: string
    username: string
    groups: string[]
    reason: string
  }[]
  validation_errors: string[]
  completed_at: string
}

export type SCIMSyncRun = {
  id: string
  application_id: string
  trigger: "MANUAL" | "SCHEDULED"
  status: SCIMSyncRunStatus
  error: string
  requested_at: string
  started_at: string | null
  completed_at: string | null
  result: SCIMSyncResult
}
