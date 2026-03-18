package model

import "time"

type User struct {
	ID                    string    `json:"id" gorm:"primaryKey"`
	EntityID              string    `json:"entity_id" gorm:"index"`
	Username              string    `json:"username"`
	FirstName             string    `json:"first_name"`
	LastName              string    `json:"last_name"`
	Email                 string    `json:"email" gorm:"-"`
	PhoneNumber           string    `json:"phone_number" gorm:"-"`
	Gender                string    `json:"gender"`
	Birthday              time.Time `json:"birthday"`
	GraduateLevel         string    `json:"graduate_level"`
	GraduationYear        int       `json:"graduation_year"`
	Major                 string    `json:"major"`
	ShirtSize             string    `json:"shirt_size"`
	JacketSize            string    `json:"jacket_size"`
	SAERegistrationNumber string    `json:"sae_registration_number"`
	AvatarURL             string    `json:"avatar_url"`
	UpdatedAt             time.Time `json:"updated_at"`
	CreatedAt             time.Time `json:"created_at"`
}

func (User) TableName() string {
	return "user"
}
