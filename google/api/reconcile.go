package api

import (
	"net/http"

	"github.com/gaucho-racing/sentinel/google/service"
	"github.com/gin-gonic/gin"
)

// TriggerReconcile kicks a full reconcile sweep in the background. Useful for
// ops and for applying a binding change without waiting for the cron.
func TriggerReconcile(c *gin.Context) {
	Require(c, RequestTokenHasScope(c, "sentinel:all"))
	service.TriggerReconcile()
	c.JSON(http.StatusAccepted, gin.H{"message": "reconcile triggered"})
}
