package model

type UserSubteam struct {
	UserID string `gorm:"primaryKey" json:"user_id"`
	RoleID string `gorm:"primaryKey" json:"role_id"`
}

func (UserSubteam) TableName() string {
	return "user_subteam"
}
