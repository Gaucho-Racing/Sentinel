package api

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gaucho-racing/sentinel/discord/pkg/logger"
	"github.com/gaucho-racing/sentinel/discord/service"
	"github.com/gin-gonic/gin"
)

type onboardingTokenInfo struct {
	DiscordID         string `json:"discord_id"`
	DiscordUsername   string `json:"discord_username"`
	DiscordGlobalName string `json:"discord_global_name"`
	DiscordAvatarURL  string `json:"discord_avatar_url"`
}

func GetOnboardingToken(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	id := c.Param("id")
	token, err := service.GetOnboardingTokenByID(id)
	switch {
	case errors.Is(err, service.ErrOnboardingTokenNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "onboarding token not found"})
		return
	case errors.Is(err, service.ErrOnboardingTokenInvalid):
		c.JSON(http.StatusGone, gin.H{"error": "onboarding token expired or already used"})
		return
	case err != nil:
		logger.SugarLogger.Errorf("Failed to fetch onboarding token %s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, onboardingTokenInfo{
		DiscordID:         token.DiscordID,
		DiscordUsername:   token.DiscordUsername,
		DiscordGlobalName: token.DiscordGlobalName,
		DiscordAvatarURL:  token.DiscordAvatarURL,
	})
}

type consumeRequest struct {
	Email                 string `json:"email" binding:"required"`
	Password              string `json:"password" binding:"required"`
	Username              string `json:"username" binding:"required"`
	FirstName             string `json:"first_name" binding:"required"`
	LastName              string `json:"last_name" binding:"required"`
	Gender                string `json:"gender" binding:"required"`
	Birthday              string `json:"birthday" binding:"required"`
	PhoneNumber           string `json:"phone_number" binding:"required"`
	GraduateLevel         string `json:"graduate_level" binding:"required"`
	GraduationYear        int    `json:"graduation_year"`
	Major                 string `json:"major"`
	ShirtSize             string `json:"shirt_size" binding:"required"`
	JacketSize            string `json:"jacket_size" binding:"required"`
	SAERegistrationNumber string `json:"sae_registration_number"`
	OccupationTitle       string `json:"occupation_title"`
	OccupationCompany     string `json:"occupation_company"`
	InitialRole           string `json:"initial_role" binding:"required"`
}

var validInitialRoles = map[string]bool{
	"member": true,
	"alumni": true,
	"other":  true,
}

const ucsbEmailDomain = "ucsb.edu"

func ConsumeOnboardingToken(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	id := c.Param("id")

	var req consumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !validInitialRoles[req.InitialRole] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid initial_role"})
		return
	}

	emailDomain := ""
	if parts := strings.SplitN(req.Email, "@", 2); len(parts) == 2 {
		emailDomain = strings.ToLower(parts[1])
	}

	switch req.InitialRole {
	case "member":
		if emailDomain != ucsbEmailDomain {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "current members must sign up with their @ucsb.edu email",
			})
			return
		}
		if req.GraduationYear > 0 && req.GraduationYear < time.Now().Year() {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "current members can't have a graduation year in the past",
			})
			return
		}
	case "alumni":
		if emailDomain == ucsbEmailDomain {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "alumni must sign up with a personal email since @ucsb.edu addresses expire after graduation",
			})
			return
		}
	}

	entityID, err := service.ConsumeOnboardingToken(id, service.OnboardingConsumePayload{
		Email:                 req.Email,
		Password:              req.Password,
		Username:              req.Username,
		FirstName:             req.FirstName,
		LastName:              req.LastName,
		Gender:                req.Gender,
		Birthday:              req.Birthday,
		PhoneNumber:           req.PhoneNumber,
		GraduateLevel:         req.GraduateLevel,
		GraduationYear:        req.GraduationYear,
		Major:                 req.Major,
		ShirtSize:             req.ShirtSize,
		JacketSize:            req.JacketSize,
		SAERegistrationNumber: req.SAERegistrationNumber,
		OccupationTitle:       req.OccupationTitle,
		OccupationCompany:     req.OccupationCompany,
		InitialRole:           req.InitialRole,
	})

	switch {
	case errors.Is(err, service.ErrOnboardingTokenNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	case errors.Is(err, service.ErrOnboardingTokenInvalid):
		c.JSON(http.StatusGone, gin.H{"error": err.Error()})
		return
	case errors.Is(err, service.ErrUsernameTaken):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	case err != nil:
		logger.SugarLogger.Errorf("Failed to consume onboarding token %s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"entity_id": entityID})
}
