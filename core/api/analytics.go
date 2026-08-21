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
		RequestTokenHasScope(c, "sentinel:all"),
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

// queryInt reads an integer query param, falling back to def when absent or
// unparseable.
func queryInt(c *gin.Context, key string, def int) int {
	if v := c.Query(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
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
	series, err := service.GetLoginTimeSeries(queryInt(c, "days", 30))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, series)
}

func AnalyticsLoginHeatmap(c *gin.Context) {
	requireAnalyticsAccess(c)
	cells, err := service.GetLoginHeatmap(queryInt(c, "days", 90))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cells)
}

func AnalyticsTopApplications(c *gin.Context) {
	requireAnalyticsAccess(c)
	apps, err := service.GetTopApplications(queryInt(c, "days", 30), queryInt(c, "limit", 10))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, apps)
}

func AnalyticsUserGrowth(c *gin.Context) {
	requireAnalyticsAccess(c)
	growth, err := service.GetUserGrowth(queryInt(c, "months", 12))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, growth)
}

func AnalyticsMemberDemographics(c *gin.Context) {
	requireAnalyticsAccess(c)
	demographics, err := service.GetMemberDemographics(queryInt(c, "major_limit", 10))
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
	funnel, err := service.GetJoinRequestFunnel(queryInt(c, "days", 90))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, funnel)
}

func AnalyticsAuditEvents(c *gin.Context) {
	requireAnalyticsAccess(c)
	events, err := service.GetAuditEvents(service.AuditEventsFilter{
		ActorID:    c.Query("actor_id"),
		Action:     c.Query("action"),
		TargetType: c.Query("target_type"),
		TargetID:   c.Query("target_id"),
		Before:     c.Query("before"),
		After:      c.Query("after"),
		Limit:      c.Query("limit"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, events)
}

func AnalyticsAuditSummary(c *gin.Context) {
	requireAnalyticsAccess(c)
	summary, err := service.GetAuditActionSummary(queryInt(c, "days", 30))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, summary)
}
