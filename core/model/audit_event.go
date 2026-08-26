package model

import "time"

// AuditAction enumerates the mutating actions Sentinel records to the audit
// trail. Values are stable strings — they're persisted and queried by the
// analytics layer, so renaming one is a breaking change.
type AuditAction string

const (
	AuditActionApplicationCreated        AuditAction = "APPLICATION_CREATED"
	AuditActionApplicationDeleted        AuditAction = "APPLICATION_DELETED"
	AuditActionApplicationSecretRevealed AuditAction = "APPLICATION_SECRET_REVEALED"
	AuditActionGroupMemberAdded          AuditAction = "GROUP_MEMBER_ADDED"
	AuditActionGroupMemberRemoved        AuditAction = "GROUP_MEMBER_REMOVED"
	AuditActionJoinRequestApproved       AuditAction = "JOIN_REQUEST_APPROVED"
	AuditActionJoinRequestRejected       AuditAction = "JOIN_REQUEST_REJECTED"
)

// AuditEvent is one recorded administrative action. Rows are written
// best-effort by the API layer (a failed write is logged, never surfaced to
// the caller) and read back by the analytics endpoints for the audit/security
// views. Metadata carries action-specific context (target name, affected
// entity, source, etc.) without needing a column per action.
type AuditEvent struct {
	ID         string    `json:"id" gorm:"primaryKey"`
	ActorID    string    `json:"actor_id" gorm:"index"`
	Action     string    `json:"action" gorm:"index"`
	TargetType string    `json:"target_type" gorm:"index"`
	TargetID   string    `json:"target_id" gorm:"index"`
	IPAddress  string    `json:"ip_address"`
	Metadata   JSONMap   `json:"metadata" gorm:"type:jsonb"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime;index"`
}

func (AuditEvent) TableName() string {
	return "audit_event"
}
