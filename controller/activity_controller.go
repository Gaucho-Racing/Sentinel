package controller

import (
    "net/http"
    "sentinel/service"

    "github.com/gin-gonic/gin"
)

// Get all activities for a user (Discord messages/reactions)
func GetActivitiesForUser(c *gin.Context) {
    Require(c, Any(
        RequestTokenHasScope(c, "sentinel:all"),
        All(
            RequestTokenHasScope(c, "user:read"),
            Any(RequestUserHasID(c, c.Param("userID")), RequestUserHasRole(c, "d_admin")),
        ),
    ))

    activities := service.GetActivitiesForUser(c.Param("userID"))
    c.JSON(http.StatusOK, activities)
}

// Get activity counts for a user grouped by day and action
func GetActivityStatsForUser(c *gin.Context) {
    Require(c, Any(
        RequestTokenHasScope(c, "sentinel:all"),
        All(
            RequestTokenHasScope(c, "user:read"),
            Any(RequestUserHasID(c, c.Param("userID")), RequestUserHasRole(c, "d_admin")),
        ),
    ))

    stats := service.GetActivityCountsByDayForUser(c.Param("userID"))
    c.JSON(http.StatusOK, stats)
}

