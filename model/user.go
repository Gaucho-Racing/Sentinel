package model

import "time"

type UserInfo struct {
	Sub            string   `json:"sub,omitempty"`
	Name           string   `json:"name,omitempty"`
	GivenName      string   `json:"given_name,omitempty"`
	FamilyName     string   `json:"family_name,omitempty"`
	Profile        string   `json:"profile,omitempty"`
	Picture        string   `json:"picture,omitempty"`
	EmailVerified  bool     `json:"email_verified,omitempty"`
	BookstackRoles []string `json:"bookstack_roles,omitempty"`
	User
}

type User struct {
	ID                    string    `gorm:"primaryKey" json:"id"`
	Username              string    `json:"username"`
	FirstName             string    `json:"first_name"`
	LastName              string    `json:"last_name"`
	Email                 string    `json:"email"`
	PhoneNumber           string    `json:"phone_number"`
	Gender                string    `json:"gender"`
	Birthday              string    `json:"birthday"`
	GraduateLevel         string    `json:"graduate_level"`
	GraduationYear        int       `json:"graduation_year"`
	Major                 string    `json:"major"`
	ShirtSize             string    `json:"shirt_size"`
	JacketSize            string    `json:"jacket_size"`
	SAERegistrationNumber string    `json:"sae_registration_number"`
	AvatarURL             string    `json:"avatar_url"`
	Verified              bool      `json:"verified"`
	Subteams              []Subteam `gorm:"-" json:"subteams"`
	Roles                 []string  `gorm:"-" json:"roles"`
	UpdatedAt             time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	CreatedAt             time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (User) TableName() string {
	return "user"
}

func (user User) String() string {
	return "(" + user.ID + ")" + " " + user.FirstName + " " + user.LastName + " [" + user.Email + "]"
}

func (user User) GetHighestRole() string {
	if user.IsAdmin() {
		return "d_admin"
	}
	if user.IsOfficer() {
		return "d_officer"
	}
	if user.IsLead() {
		return "d_lead"
	}
	if user.IsSpecialAdvisor() {
		return "d_special_advisor"
	}
	if user.IsTeamMember() {
		return "d_team_member"
	}
	if user.IsMember() {
		return "d_member"
	}
	if user.IsAlumni() {
		return "d_alumni"
	}
	return ""
}

func (user User) HasRole(role string) bool {
	for _, r := range user.Roles {
		if r == role {
			return true
		}
	}
	return false
}

func (user User) IsAdmin() bool {
	return user.HasRole("d_admin")
}

func (user User) IsOfficer() bool {
	return user.HasRole("d_officer")
}

func (user User) IsLead() bool {
	return user.HasRole("d_lead")
}

func (user User) IsSpecialAdvisor() bool {
	return user.HasRole("d_special_advisor")
}

func (user User) IsInnerCircle() bool {
	return user.IsAdmin() || user.IsOfficer() || user.IsLead() || user.IsSpecialAdvisor()
}

func (user User) IsTeamMember() bool {
	return user.HasRole("d_team_member")
}

func (user User) IsMember() bool {
	return user.HasRole("d_member")
}

func (user User) IsAlumni() bool {
	return user.HasRole("d_alumni")
}
