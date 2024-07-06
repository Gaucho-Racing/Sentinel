package model

type WikiArrayResponse[T any] struct {
	Data []T `json:"data"`
}

type WikiUserCreate struct {
	Name           string `json:"name"`
	Email          string `json:"email"`
	Roles          []int  `json:"roles"`
	ExternalAuthID string `json:"external_auth_id"`
	Password       string `json:"password"`
	SendInvite     bool   `json:"send_invite"`
}

type WikiUser struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Email          string `json:"email"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
	ExternalAuthID string `json:"external_auth_id"`
	Slug           string `json:"slug"`
	LastActivityAt string `json:"last_activity_at"`
	ProfileURL     string `json:"profile_url"`
	EditURL        string `json:"edit_url"`
	AvatarURL      string `json:"avatar_url"`
}

type WikiRole int

const (
	WikiRoleAdmin  WikiRole = 1
	WikiRoleDevOps WikiRole = 6
	WikiRoleEditor WikiRole = 2
	WikiRoleLead   WikiRole = 5
	WikiRolePublic WikiRole = 4
	WikiRoleViewer WikiRole = 3
)

/*
{
			"id": 1,
			"name": "GR Admin",
			"email": "admin@gauchoracing.com",
			"created_at": "2024-01-08T22:25:05.000000Z",
			"updated_at": "2024-01-08T22:59:47.000000Z",
			"external_auth_id": "",
			"slug": "gr-admin",
			"last_activity_at": "2024-07-06T07:00:59.000000Z",
			"profile_url": "https:\/\/wiki.gauchoracing.com\/user\/gr-admin",
			"edit_url": "https:\/\/wiki.gauchoracing.com\/settings\/users\/1",
			"avatar_url": "https:\/\/wiki.gauchoracing.com\/uploads\/images\/user\/2024-01\/thumbs-50-50\/O1tgkEgkCZ4df2Wv-gr-logo-blank.png"
		}
*/
