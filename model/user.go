package model

import "time"

type User struct {
	ID          string     `gorm:"primaryKey" json:"id"`
	Username    string     `json:"username"`
	FirstName   string     `json:"first_name"`
	LastName    string     `json:"last_name"`
	Email       string     `json:"email"`
	PhoneNumber string     `json:"phone_number"`
	ShirtSize   string     `json:"shirt_size"`
	JacketSize  string     `json:"jacket_size"`
	AvatarURL   string     `json:"avatar_url"`
	Verified    bool       `json:"verified"`
	Subteams    []Subteam  `gorm:"-" json:"subteams"`
	Roles       []UserRole `gorm:"-" json:"roles"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

func (User) TableName() string {
	return "user"
}

func (user User) String() string {
	return "(" + user.ID + ")" + " " + user.FirstName + " " + user.LastName + " [" + user.Email + "]"
}
