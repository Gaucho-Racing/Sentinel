package model

type GithubInvite struct {
	Username string `json:"username"`
}

type GithubOrgUser struct {
	Url             string     `json:"url"`
	State           string     `json:"state"`
	Role            string     `json:"role"`
	OrganizationUrl string     `json:"organization_url"`
	User            GithubUser `json:"user"`
	Organization    GithubOrg  `json:"organization"`
}

type GithubUser struct {
	ID         int    `json:"id"`
	Login      string `json:"login"`
	NodeID     string `json:"node_id"`
	AvatarUrl  string `json:"avatar_url"`
	GravatarId string `json:"gravatar_id"`
	Type       string `json:"type"`
	SiteAdmin  bool   `json:"site_admin"`
}

type GithubOrg struct {
	ID               int    `json:"id"`
	Login            string `json:"login"`
	NodeID           string `json:"node_id"`
	Url              string `json:"url"`
	ReposUrl         string `json:"repos_url"`
	EventsUrl        string `json:"events_url"`
	HooksUrl         string `json:"hooks_url"`
	IssuesUrl        string `json:"issues_url"`
	MembersUrl       string `json:"members_url"`
	PublicMembersUrl string `json:"public_members_url"`
	AvatarUrl        string `json:"avatar_url"`
	Description      string `json:"description"`
}
