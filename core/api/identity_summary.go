package api

import (
	"net/http"
	"strings"

	"github.com/gaucho-racing/sentinel/core/service"
	"github.com/gin-gonic/gin"
)

const maxIdentitySummaryIDs = 100

type identitySummaryRequest struct {
	IDs []string `json:"ids" binding:"required"`
}

func ResolveIdentitySummaries(c *gin.Context) {
	Require(c, Any(
		RequestTokenHasAudience(c, "sentinel"),
		RequestTokenHasScope(c, "sentinel:all"),
		RequestTokenHasScope(c, "user:read"),
	))

	var req identitySummaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ids is required"})
		return
	}
	if len(req.IDs) > maxIdentitySummaryIDs {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at most 100 entity IDs may be resolved at once"})
		return
	}
	for _, entityID := range req.IDs {
		if strings.TrimSpace(entityID) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "entity IDs must not be empty"})
			return
		}
	}

	summaries, err := service.GetIdentitySummaries(req.IDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, summaries)
}
