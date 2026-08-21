package service

import (
	"time"

	"github.com/gaucho-racing/sentinel/core/database"
	"github.com/gaucho-racing/sentinel/core/model"
)

// CategoryCount is a generic label/value pair used by the breakdown charts
// (grad year, major, auth method, audit action, etc.).
type CategoryCount struct {
	Label string `json:"label"`
	Count int64  `json:"count"`
}

// AnalyticsOverview is the set of headline KPIs shown as stat cards at the top
// of the analytics hub.
type AnalyticsOverview struct {
	TotalUsers           int64 `json:"total_users"`
	TotalServiceAccounts int64 `json:"total_service_accounts"`
	TotalApplications    int64 `json:"total_applications"`
	TotalGroups          int64 `json:"total_groups"`
	NewUsers30d          int64 `json:"new_users_30d"`
	Logins24h            int64 `json:"logins_24h"`
	Logins7d             int64 `json:"logins_7d"`
	Logins30d            int64 `json:"logins_30d"`
	ActiveUsers7d        int64 `json:"active_users_7d"`
	ActiveUsers30d       int64 `json:"active_users_30d"`
	PendingJoinRequests  int64 `json:"pending_join_requests"`
	AuditEvents7d        int64 `json:"audit_events_7d"`
}

func GetAnalyticsOverview() (AnalyticsOverview, error) {
	var o AnalyticsOverview
	now := time.Now()
	db := database.DB

	db.Model(&model.User{}).Count(&o.TotalUsers)
	db.Model(&model.Entity{}).Where("type = ?", model.EntityTypeServiceAccount).Count(&o.TotalServiceAccounts)
	db.Model(&model.Application{}).Count(&o.TotalApplications)
	db.Model(&model.Group{}).Count(&o.TotalGroups)
	db.Model(&model.User{}).Where("created_at > ?", now.AddDate(0, 0, -30)).Count(&o.NewUsers30d)

	db.Model(&model.EntityLogin{}).Where("created_at > ?", now.Add(-24*time.Hour)).Count(&o.Logins24h)
	db.Model(&model.EntityLogin{}).Where("created_at > ?", now.AddDate(0, 0, -7)).Count(&o.Logins7d)
	db.Model(&model.EntityLogin{}).Where("created_at > ?", now.AddDate(0, 0, -30)).Count(&o.Logins30d)
	db.Model(&model.EntityLogin{}).Where("created_at > ?", now.AddDate(0, 0, -7)).Distinct("entity_id").Count(&o.ActiveUsers7d)
	db.Model(&model.EntityLogin{}).Where("created_at > ?", now.AddDate(0, 0, -30)).Distinct("entity_id").Count(&o.ActiveUsers30d)

	db.Model(&model.GroupJoinRequest{}).Where("status = ?", model.GroupJoinRequestStatusPending).Count(&o.PendingJoinRequests)
	db.Model(&model.AuditEvent{}).Where("created_at > ?", now.AddDate(0, 0, -7)).Count(&o.AuditEvents7d)

	return o, nil
}

// LoginPoint is one day in the sign-in trend series.
type LoginPoint struct {
	Date        string `json:"date"` // YYYY-MM-DD (UTC)
	Logins      int64  `json:"logins"`
	UniqueUsers int64  `json:"unique_users"`
}

// GetLoginTimeSeries returns per-day login counts and distinct-user counts for
// the trailing `days` window, gap-filled so every calendar day is present
// (charts render a continuous axis without client-side interpolation).
func GetLoginTimeSeries(days int) ([]LoginPoint, error) {
	if days <= 0 {
		days = 30
	}
	now := time.Now().UTC()
	startDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -(days - 1))

	type row struct {
		Day         string
		Logins      int64
		UniqueUsers int64
	}
	rows := []row{}
	sql := `
		SELECT TO_CHAR(created_at, 'YYYY-MM-DD') AS day,
		       COUNT(*) AS logins,
		       COUNT(DISTINCT entity_id) AS unique_users
		FROM entity_login
		WHERE created_at >= ?
		GROUP BY day
		ORDER BY day
	`
	if err := database.DB.Raw(sql, startDay).Scan(&rows).Error; err != nil {
		return []LoginPoint{}, err
	}

	byDay := map[string]row{}
	for _, r := range rows {
		byDay[r.Day] = r
	}
	out := make([]LoginPoint, 0, days)
	for i := 0; i < days; i++ {
		d := startDay.AddDate(0, 0, i).Format("2006-01-02")
		if r, ok := byDay[d]; ok {
			out = append(out, LoginPoint{Date: d, Logins: r.Logins, UniqueUsers: r.UniqueUsers})
		} else {
			out = append(out, LoginPoint{Date: d})
		}
	}
	return out, nil
}

// HeatmapCell is one (weekday, hour) bucket of sign-in volume. Weekday is
// 0=Sunday..6=Saturday, hour is 0..23, both in UTC.
type HeatmapCell struct {
	Weekday int   `json:"weekday"`
	Hour    int   `json:"hour"`
	Count   int64 `json:"count"`
}

func GetLoginHeatmap(days int) ([]HeatmapCell, error) {
	if days <= 0 {
		days = 90
	}
	start := time.Now().AddDate(0, 0, -days)
	cells := []HeatmapCell{}
	sql := `
		SELECT CAST(EXTRACT(DOW FROM created_at) AS INTEGER) AS weekday,
		       CAST(EXTRACT(HOUR FROM created_at) AS INTEGER) AS hour,
		       COUNT(*) AS count
		FROM entity_login
		WHERE created_at >= ?
		GROUP BY weekday, hour
	`
	if err := database.DB.Raw(sql, start).Scan(&cells).Error; err != nil {
		return []HeatmapCell{}, err
	}
	return cells, nil
}

// TopApplication ranks an app by sign-in volume over the window. Name/IconURL
// are left-joined so logins against a deleted or unknown client_id still show
// (falling back to the raw client_id as the label).
type TopApplication struct {
	ClientID    string `json:"client_id"`
	Name        string `json:"name"`
	IconURL     string `json:"icon_url"`
	Logins      int64  `json:"logins"`
	UniqueUsers int64  `json:"unique_users"`
}

func GetTopApplications(days int, limit int) ([]TopApplication, error) {
	if days <= 0 {
		days = 30
	}
	if limit <= 0 {
		limit = 10
	}
	start := time.Now().AddDate(0, 0, -days)
	apps := []TopApplication{}
	sql := `
		SELECT l.client_id AS client_id,
		       COALESCE(NULLIF(a.name, ''), l.client_id) AS name,
		       COALESCE(a.icon_url, '') AS icon_url,
		       COUNT(*) AS logins,
		       COUNT(DISTINCT l.entity_id) AS unique_users
		FROM entity_login l
		LEFT JOIN application a ON a.client_id = l.client_id
		WHERE l.created_at >= ?
		GROUP BY l.client_id, a.name, a.icon_url
		ORDER BY logins DESC
		LIMIT ?
	`
	if err := database.DB.Raw(sql, start, limit).Scan(&apps).Error; err != nil {
		return []TopApplication{}, err
	}
	return apps, nil
}

// UserGrowthPoint is one month of member growth: new signups that month plus
// the running total of all members through that month.
type UserGrowthPoint struct {
	Date       string `json:"date"` // YYYY-MM (UTC)
	NewUsers   int64  `json:"new_users"`
	Cumulative int64  `json:"cumulative"`
}

func GetUserGrowth(months int) ([]UserGrowthPoint, error) {
	if months <= 0 {
		months = 12
	}
	now := time.Now().UTC()
	startMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, -(months - 1), 0)

	// Members created before the window form the cumulative baseline so the
	// running total reflects the whole roster, not just the visible range.
	var baseline int64
	database.DB.Model(&model.User{}).Where("created_at < ?", startMonth).Count(&baseline)

	type row struct {
		Month    string
		NewUsers int64
	}
	rows := []row{}
	sql := `
		SELECT TO_CHAR(created_at, 'YYYY-MM') AS month, COUNT(*) AS new_users
		FROM "user"
		WHERE created_at >= ?
		GROUP BY month
		ORDER BY month
	`
	if err := database.DB.Raw(sql, startMonth).Scan(&rows).Error; err != nil {
		return []UserGrowthPoint{}, err
	}

	byMonth := map[string]int64{}
	for _, r := range rows {
		byMonth[r.Month] = r.NewUsers
	}
	out := make([]UserGrowthPoint, 0, months)
	cum := baseline
	for i := 0; i < months; i++ {
		m := startMonth.AddDate(0, i, 0).Format("2006-01")
		n := byMonth[m]
		cum += n
		out = append(out, UserGrowthPoint{Date: m, NewUsers: n, Cumulative: cum})
	}
	return out, nil
}

// MemberDemographics bundles the member breakdowns the officer-facing view
// cares about into a single response.
type MemberDemographics struct {
	ByGradYear      []CategoryCount `json:"by_grad_year"`
	ByMajor         []CategoryCount `json:"by_major"`
	ByGraduateLevel []CategoryCount `json:"by_graduate_level"`
}

func GetMemberDemographics(majorLimit int) (MemberDemographics, error) {
	if majorLimit <= 0 {
		majorLimit = 10
	}
	var d MemberDemographics

	d.ByGradYear = []CategoryCount{}
	if err := database.DB.Raw(`
		SELECT CAST(graduation_year AS TEXT) AS label, COUNT(*) AS count
		FROM "user"
		WHERE graduation_year > 0
		GROUP BY graduation_year
		ORDER BY graduation_year
	`).Scan(&d.ByGradYear).Error; err != nil {
		return MemberDemographics{}, err
	}

	d.ByMajor = []CategoryCount{}
	if err := database.DB.Raw(`
		SELECT major AS label, COUNT(*) AS count
		FROM "user"
		WHERE major <> ''
		GROUP BY major
		ORDER BY count DESC
		LIMIT ?
	`, majorLimit).Scan(&d.ByMajor).Error; err != nil {
		return MemberDemographics{}, err
	}

	d.ByGraduateLevel = []CategoryCount{}
	if err := database.DB.Raw(`
		SELECT graduate_level AS label, COUNT(*) AS count
		FROM "user"
		WHERE graduate_level <> ''
		GROUP BY graduate_level
		ORDER BY count DESC
	`).Scan(&d.ByGraduateLevel).Error; err != nil {
		return MemberDemographics{}, err
	}

	return d, nil
}

// AuthMethodBreakdown counts distinct entities holding each authentication
// method. An entity can appear in more than one bucket (multi-auth), which is
// exactly the signal this view surfaces.
type AuthMethodBreakdown struct {
	Email   int64 `json:"email"`
	Phone   int64 `json:"phone"`
	Discord int64 `json:"discord"`
	Google  int64 `json:"google"`
	GitHub  int64 `json:"github"`
}

func GetAuthMethodBreakdown() (AuthMethodBreakdown, error) {
	var b AuthMethodBreakdown
	db := database.DB
	db.Model(&model.EntityEmail{}).Distinct("entity_id").Count(&b.Email)
	db.Model(&model.EntityPhone{}).Distinct("entity_id").Count(&b.Phone)
	db.Model(&model.EntityExternalAuth{}).Where("provider = ?", model.ExternalAuthProviderDiscord).Distinct("entity_id").Count(&b.Discord)
	db.Model(&model.EntityExternalAuth{}).Where("provider = ?", model.ExternalAuthProviderGoogle).Distinct("entity_id").Count(&b.Google)
	db.Model(&model.EntityExternalAuth{}).Where("provider = ?", model.ExternalAuthProviderGitHub).Distinct("entity_id").Count(&b.GitHub)
	return b, nil
}

// GroupMembership is per-group membership sizing split by member source. Powers
// the stacked group breakdown (DIRECT vs CONDITIONAL vs DISCORD).
type GroupMembership struct {
	GroupID     string `json:"group_id"`
	Name        string `json:"name"`
	MemberCount int64  `json:"member_count"`
	Direct      int64  `json:"direct"`
	Conditional int64  `json:"conditional"`
	Discord     int64  `json:"discord"`
}

func GetGroupMembershipBreakdown() ([]GroupMembership, error) {
	out := []GroupMembership{}
	sql := `
		SELECT g.id AS group_id,
		       g.name AS name,
		       COUNT(m.entity_id) AS member_count,
		       COUNT(m.entity_id) FILTER (WHERE m.source = 'DIRECT') AS direct,
		       COUNT(m.entity_id) FILTER (WHERE m.source = 'CONDITIONAL') AS conditional,
		       COUNT(m.entity_id) FILTER (WHERE m.source = 'DISCORD') AS discord
		FROM "group" g
		LEFT JOIN group_member m ON m.group_id = g.id
		GROUP BY g.id, g.name
		ORDER BY member_count DESC
	`
	if err := database.DB.Raw(sql).Scan(&out).Error; err != nil {
		return []GroupMembership{}, err
	}
	return out, nil
}

// JoinRequestFunnel summarizes join-request throughput over the window. Pending
// is point-in-time (the current backlog); approved/rejected are counted within
// the window; the median decision time is in hours.
type JoinRequestFunnel struct {
	Pending             int64   `json:"pending"`
	Approved            int64   `json:"approved"`
	Rejected            int64   `json:"rejected"`
	MedianDecisionHours float64 `json:"median_decision_hours"`
}

func GetJoinRequestFunnel(days int) (JoinRequestFunnel, error) {
	if days <= 0 {
		days = 90
	}
	start := time.Now().AddDate(0, 0, -days)
	var f JoinRequestFunnel
	db := database.DB
	db.Model(&model.GroupJoinRequest{}).Where("status = ?", model.GroupJoinRequestStatusPending).Count(&f.Pending)
	db.Model(&model.GroupJoinRequest{}).Where("status = ? AND created_at >= ?", model.GroupJoinRequestStatusApproved, start).Count(&f.Approved)
	db.Model(&model.GroupJoinRequest{}).Where("status = ? AND created_at >= ?", model.GroupJoinRequestStatusRejected, start).Count(&f.Rejected)

	var median float64
	sql := `
		SELECT COALESCE(
			percentile_cont(0.5) WITHIN GROUP (
				ORDER BY EXTRACT(EPOCH FROM (reviewed_at - created_at)) / 3600.0
			), 0)
		FROM group_join_request
		WHERE status IN ('APPROVED', 'REJECTED')
		  AND reviewed_at > created_at
		  AND created_at >= ?
	`
	if err := database.DB.Raw(sql, start).Row().Scan(&median); err != nil {
		return f, err
	}
	f.MedianDecisionHours = median
	return f, nil
}
