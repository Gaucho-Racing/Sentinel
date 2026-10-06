package model

import "time"

// OrgPosition is one seat in the team's reporting structure. The positions
// form a tree through ParentID; the chart is that tree rendered top-down.
//
// A position may point at a Sentinel group, in which case that group's live
// membership renders beneath it — so the roster under "Electrical Technical
// Director" stays current through Discord and conditional syncs without anyone
// touching the chart. Positions that are purely people (President, Vice
// President) leave GroupID empty.
//
// One row per seat. Co-leads are two rows sharing a title, which is also how
// an org chart should draw them.
type OrgPosition struct {
	ID    string `json:"id" gorm:"primaryKey"`
	Title string `json:"title"`
	// Empty for a root position. Multiple roots are allowed — an advisor or a
	// faculty sponsor usually doesn't sit under the President.
	ParentID string `json:"parent_id" gorm:"index"`
	// Empty means the seat is vacant, which is worth showing rather than hiding.
	EntityID string `json:"entity_id" gorm:"index"`
	// Empty means no team reports into this position.
	GroupID string `json:"group_id" gorm:"index"`
	// Ordering among siblings; ties fall back to title.
	Rank      int       `json:"rank"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (OrgPosition) TableName() string {
	return "org_position"
}
