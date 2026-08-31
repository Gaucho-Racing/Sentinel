package model

type IdentityApplicationSummary struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	ClientID string `json:"client_id"`
	IconURL  string `json:"icon_url"`
}

type IdentitySummary struct {
	ID          string                      `json:"id"`
	Type        EntityType                  `json:"type"`
	Name        string                      `json:"name"`
	Username    string                      `json:"username,omitempty"`
	AvatarURL   string                      `json:"avatar_url,omitempty"`
	Application *IdentityApplicationSummary `json:"application,omitempty"`
}
