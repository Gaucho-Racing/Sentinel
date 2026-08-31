package api

import (
	"net/http"
	"strconv"

	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/sentinel/core/service"
	"github.com/gin-gonic/gin"
)

// requireAnalyticsAccess gates the analytics + audit endpoints. They expose
// team-wide aggregates and administrative history, so access is limited to
// first-party admin sessions (the dashboard UI) and internal automation
// (sentinel:all). Mirrors the GetApplicationSecret gate.
func requireAnalyticsAccess(c *gin.Context) {
	Require(c, Any(
		RequestTokenHasInternalAccess(c),
		RequestTokenHasAudience(c, "sentinel") && RequestUserIsAdmin(c),
	))
}

// recordAudit writes an audit row for the current request, resolving the actor
// and client IP from context. Best-effort: never blocks or fails the handler.
func recordAudit(c *gin.Context, action model.AuditAction, targetType string, targetID string, metadata model.JSONMap) {
	service.RecordAuditEvent(model.AuditEvent{
		ActorID:    GetRequestTokenEntityID(c),
		Action:     string(action),
		TargetType: targetType,
		TargetID:   targetID,
		IPAddress:  c.ClientIP(),
		Metadata:   metadata,
	})
}

func boundedQueryInt(c *gin.Context, key string, fallback int, maximum int) (int, bool) {
	raw := c.Query(key)
	if raw == "" {
		return fallback, true
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 || value > maximum {
		c.JSON(http.StatusBadRequest, gin.H{"error": key + " must be between 1 and " + strconv.Itoa(maximum)})
		return 0, false
	}
	return value, true
}

func AnalyticsOverview(c *gin.Context) {
	requireAnalyticsAccess(c)
	overview, err := service.GetAnalyticsOverview()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, overview)
}

func AnalyticsLoginTimeSeries(c *gin.Context) {
	requireAnalyticsAccess(c)
	days, ok := boundedQueryInt(c, "days", 30, service.MaxAnalyticsDays)
	if !ok {
		return
	}
	series, err := service.GetLoginTimeSeries(days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, series)
}

func AnalyticsLoginHeatmap(c *gin.Context) {
	requireAnalyticsAccess(c)
	days, ok := boundedQueryInt(c, "days", 90, service.MaxAnalyticsDays)
	if !ok {
		return
	}
	cells, err := service.GetLoginHeatmap(days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cells)
}

func AnalyticsTopApplications(c *gin.Context) {
	requireAnalyticsAccess(c)
	days, ok := boundedQueryInt(c, "days", 30, service.MaxAnalyticsDays)
	if !ok {
		return
	}
	limit, ok := boundedQueryInt(c, "limit", 10, service.MaxAnalyticsLimit)
	if !ok {
		return
	}
	apps, err := service.GetTopApplications(days, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, apps)
}

func AnalyticsUserGrowth(c *gin.Context) {
	requireAnalyticsAccess(c)
	months, ok := boundedQueryInt(c, "months", 12, service.MaxAnalyticsMonths)
	if !ok {
		return
	}
	growth, err := service.GetUserGrowth(months)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, growth)
}

func AnalyticsMemberDemographics(c *gin.Context) {
	requireAnalyticsAccess(c)
	limit, ok := boundedQueryInt(c, "major_limit", 10, service.MaxAnalyticsLimit)
	if !ok {
		return
	}
	demographics, err := service.GetMemberDemographics(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, demographics)
}

func AnalyticsAuthMethods(c *gin.Context) {
	requireAnalyticsAccess(c)
	breakdown, err := service.GetAuthMethodBreakdown()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, breakdown)
}

func AnalyticsGroupMembership(c *gin.Context) {
	requireAnalyticsAccess(c)
	breakdown, err := service.GetGroupMembershipBreakdown()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, breakdown)
}

func AnalyticsJoinRequests(c *gin.Context) {
	requireAnalyticsAccess(c)
	days, ok := boundedQueryInt(c, "days", 90, service.MaxAnalyticsDays)
	if !ok {
		return
	}
	funnel, err := service.GetJoinRequestFunnel(days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, funnel)
}

func AnalyticsAuditEvents(c *gin.Context) {
	requireAnalyticsAccess(c)
	limit, ok := boundedQueryInt(c, "limit", 100, service.MaxAuditEventLimit)
	if !ok {
		return
	}
	events, err := service.GetAuditEvents(service.AuditEventsFilter{
		ActorID:    c.Query("actor_id"),
		Action:     c.Query("action"),
		TargetType: c.Query("target_type"),
		TargetID:   c.Query("target_id"),
		Before:     c.Query("before"),
		After:      c.Query("after"),
		Limit:      strconv.Itoa(limit),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, events)
}

func AnalyticsAuditSummary(c *gin.Context) {
	requireAnalyticsAccess(c)
	days, ok := boundedQueryInt(c, "days", 30, service.MaxAnalyticsDays)
	if !ok {
		return
	}
	summary, err := service.GetAuditActionSummary(days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, summary)
}
