package api

import (
	"net/http"

	"github.com/gaucho-racing/sentinel/discord/pkg/logger"
	"github.com/gaucho-racing/sentinel/discord/service"
	"github.com/gin-gonic/gin"
)

type discordRole struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Color       int    `json:"color"`
	Position    int    `json:"position"`
	Hoist       bool   `json:"hoist"`
	Mentionable bool   `json:"mentionable"`
	Managed     bool   `json:"managed"`
}

type discordChannel struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     int    `json:"type"`
	Position int    `json:"position"`
	ParentID string `json:"parent_id"`
	Topic    string `json:"topic"`
	NSFW     bool   `json:"nsfw"`
}

func GetRoles(c *gin.Context) {
	roles, err := service.GetGuildRoles()
	if err != nil {
		logger.SugarLogger.Errorf("Failed to fetch guild roles: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to fetch roles from Discord"})
		return
	}

	out := make([]discordRole, 0, len(roles))
	for _, r := range roles {
		out = append(out, discordRole{
			ID:          r.ID,
			Name:        r.Name,
			Color:       r.Color,
			Position:    r.Position,
			Hoist:       r.Hoist,
			Mentionable: r.Mentionable,
			Managed:     r.Managed,
		})
	}
	c.JSON(http.StatusOK, out)
}

func GetChannels(c *gin.Context) {
	channels, err := service.GetGuildChannels()
	if err != nil {
		logger.SugarLogger.Errorf("Failed to fetch guild channels: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to fetch channels from Discord"})
		return
	}

	out := make([]discordChannel, 0, len(channels))
	for _, ch := range channels {
		out = append(out, discordChannel{
			ID:       ch.ID,
			Name:     ch.Name,
			Type:     int(ch.Type),
			Position: ch.Position,
			ParentID: ch.ParentID,
			Topic:    ch.Topic,
			NSFW:     ch.NSFW,
		})
	}
	c.JSON(http.StatusOK, out)
}
