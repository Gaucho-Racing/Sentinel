package service

import (
	"strconv"
	"time"

	"github.com/gaucho-racing/sentinel/core/database"
	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/sentinel/core/pkg/logger"
	"github.com/gaucho-racing/ulid-go"
)

// RecordAuditEvent persists an audit row best-effort. Callers invoke it after
// a mutation has already succeeded, so a write failure here must not change
// the request outcome — it's logged and swallowed. Returns nothing for the
// same reason: there is no error the caller should act on.
func RecordAuditEvent(event model.AuditEvent) {
	if event.ID == "" {
		event.ID = ulid.Make().Prefixed("aud")
	}
	if err := database.DB.Create(&event).Error; err != nil {
		logger.SugarLogger.Errorf("Failed to record audit event %s: %v", event.Action, err)
	}
}

// AuditEventsFilter holds the query params accepted by GetAuditEvents. All
// string fields are optional; empty strings are ignored. Mirrors the shape of
// EntityLoginsFilter.
type AuditEventsFilter struct {
	ActorID    string
	Action     string
	TargetType string
	TargetID   string
	Before     string // RFC3339; matches events with created_at < Before
	After      string // RFC3339; matches events with created_at > After
	Limit      string // integer string; unset defaults to 100
}

func GetAuditEvents(filter AuditEventsFilter) ([]model.AuditEvent, error) {
	events := []model.AuditEvent{}
	query := database.DB.Model(&model.AuditEvent{})
	if filter.ActorID != "" {
		query = query.Where("actor_id = ?", filter.ActorID)
	}
	if filter.Action != "" {
		query = query.Where("action = ?", filter.Action)
	}
	if filter.TargetType != "" {
		query = query.Where("target_type = ?", filter.TargetType)
	}
	if filter.TargetID != "" {
		query = query.Where("target_id = ?", filter.TargetID)
	}
	if filter.Before != "" {
		if t, err := time.Parse(time.RFC3339, filter.Before); err == nil {
			query = query.Where("created_at < ?", t)
		}
	}
	if filter.After != "" {
		if t, err := time.Parse(time.RFC3339, filter.After); err == nil {
			query = query.Where("created_at > ?", t)
		}
	}
	query = query.Order("created_at desc")
	limit := 100
	if filter.Limit != "" {
		if n, err := strconv.Atoi(filter.Limit); err == nil && n > 0 {
			limit = n
		}
	}
	query = query.Limit(limit)
	if err := query.Find(&events).Error; err != nil {
		return []model.AuditEvent{}, err
	}
	return events, nil
}

// GetAuditActionSummary returns the count of audit events per action over the
// trailing window, most frequent first. Powers the audit breakdown chart.
func GetAuditActionSummary(days int) ([]CategoryCount, error) {
	if days <= 0 {
		days = 30
	}
	start := time.Now().AddDate(0, 0, -days)
	out := []CategoryCount{}
	sql := `
		SELECT action AS label, COUNT(*) AS count
		FROM audit_event
		WHERE created_at >= ?
		GROUP BY action
		ORDER BY count DESC
	`
	if err := database.DB.Raw(sql, start).Scan(&out).Error; err != nil {
		return []CategoryCount{}, err
	}
	return out, nil
}
