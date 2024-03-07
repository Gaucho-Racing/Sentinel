package model

type UserSubteam struct {
	UserID string `json:"user_id"`
	RoleID string `json:"role_id"`
}

func (UserSubteam) TableName() string {
	return "user_subteam"
}
