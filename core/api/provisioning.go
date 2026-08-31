package api

import (
	"net/http"

	"github.com/gaucho-racing/sentinel/core/service"
	"github.com/gin-gonic/gin"
)

func GetApplicationProvisioningSnapshot(c *gin.Context) {
	Require(c, RequestTokenHasScope(c, "sentinel:all"))
	snapshot, err := service.GetApplicationProvisioningSnapshot(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not build provisioning snapshot"})
		return
	}
	c.JSON(http.StatusOK, snapshot)
}
