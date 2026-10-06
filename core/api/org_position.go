package api

import (
	"errors"
	"net/http"

	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/sentinel/core/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// requireOrgWriteAccess gates every mutation on the org chart. The structure is
// org-wide, so editing it is an admin action; reading it is not. Mirrors
// requireAuditAccess.
func requireOrgWriteAccess(c *gin.Context) {
	Require(c, Any(
		RequestTokenHasInternalAccess(c),
		RequestTokenHasFirstPartyAccess(c) && RequestUserIsAdmin(c),
	))
}

// orgChartGroup is the attached team as the chart needs it. Missing is set when
// the position still references a group that has since been deleted — surfaced
// rather than silently dropped, since the roster beneath it will be empty and
// that would otherwise look like a team with no members.
type orgChartGroup struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	MemberCount int    `json:"member_count"`
	Missing     bool   `json:"missing"`
}

// orgChartNode is one position with everything needed to draw it. Returned as a
// flat list; the client assembles the tree from parent_id.
type orgChartNode struct {
	Position model.OrgPosition       `json:"position"`
	Holder   *model.IdentitySummary  `json:"holder,omitempty"`
	Group    *orgChartGroup          `json:"group,omitempty"`
	Members  []model.IdentitySummary `json:"members"`
}

func GetOrgPositions(c *gin.Context) {
	Require(c, RequestTokenHasFirstPartyAccess(c) || RequestTokenHasInternalAccess(c))

	positions, err := service.GetAllOrgPositions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, positions)
}

// GetOrgChart resolves the whole chart server-side: holders and group rosters
// become identity summaries in one batched lookup. Composing this on the client
// would cost a request per attached group.
func GetOrgChart(c *gin.Context) {
	Require(c, RequestTokenHasFirstPartyAccess(c) || RequestTokenHasInternalAccess(c))

	positions, err := service.GetAllOrgPositions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	memberIDsByGroup := map[string][]string{}
	groupsByID := map[string]model.Group{}
	for _, position := range positions {
		if position.GroupID == "" {
			continue
		}
		if _, done := memberIDsByGroup[position.GroupID]; done {
			continue
		}
		if group, err := service.GetGroupByID(position.GroupID); err == nil {
			groupsByID[position.GroupID] = group
		}
		members, err := service.GetMembersForGroup(position.GroupID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ids := make([]string, 0, len(members))
		for _, member := range members {
			ids = append(ids, member.EntityID)
		}
		memberIDsByGroup[position.GroupID] = ids
	}

	// One batched resolve for every entity the chart mentions — holders and
	// rosters together — rather than a lookup per node.
	wanted := []string{}
	for _, position := range positions {
		if position.EntityID != "" {
			wanted = append(wanted, position.EntityID)
		}
	}
	for _, ids := range memberIDsByGroup {
		wanted = append(wanted, ids...)
	}
	summaries, err := service.GetIdentitySummaries(wanted)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	summaryByID := make(map[string]model.IdentitySummary, len(summaries))
	for _, summary := range summaries {
		summaryByID[summary.ID] = summary
	}

	nodes := make([]orgChartNode, 0, len(positions))
	for _, position := range positions {
		node := orgChartNode{Position: position, Members: []model.IdentitySummary{}}

		if position.EntityID != "" {
			if summary, ok := summaryByID[position.EntityID]; ok {
				node.Holder = &summary
			}
		}

		if position.GroupID != "" {
			group, known := groupsByID[position.GroupID]
			ids := memberIDsByGroup[position.GroupID]
			for _, entityID := range ids {
				// The holder already has their own box; listing them in the
				// roster underneath would draw the same person twice.
				if entityID == position.EntityID {
					continue
				}
				if summary, ok := summaryByID[entityID]; ok {
					node.Members = append(node.Members, summary)
				}
			}
			node.Group = &orgChartGroup{
				ID:          position.GroupID,
				Name:        group.Name,
				MemberCount: len(node.Members),
				Missing:     !known,
			}
		}

		nodes = append(nodes, node)
	}

	c.JSON(http.StatusOK, nodes)
}

type orgPositionRequest struct {
	Title    string `json:"title"`
	ParentID string `json:"parent_id"`
	EntityID string `json:"entity_id"`
	GroupID  string `json:"group_id"`
	Rank     int    `json:"rank"`
}

func CreateOrgPosition(c *gin.Context) {
	requireOrgWriteAccess(c)

	var req orgPositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	position, err := service.CreateOrgPosition(model.OrgPosition{
		Title:    req.Title,
		ParentID: req.ParentID,
		EntityID: req.EntityID,
		GroupID:  req.GroupID,
		Rank:     req.Rank,
	})
	if err != nil {
		respondOrgPositionError(c, err)
		return
	}
	recordAudit(c, model.AuditActionOrgPositionCreated, "org_position", position.ID, model.JSONMap{
		"title": position.Title,
	})
	c.JSON(http.StatusOK, position)
}

func UpdateOrgPosition(c *gin.Context) {
	requireOrgWriteAccess(c)

	id := c.Param("id")
	if _, err := service.GetOrgPositionByID(id); err != nil {
		respondOrgPositionError(c, err)
		return
	}
	var req orgPositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	position, err := service.UpdateOrgPosition(model.OrgPosition{
		ID:       id,
		Title:    req.Title,
		ParentID: req.ParentID,
		EntityID: req.EntityID,
		GroupID:  req.GroupID,
		Rank:     req.Rank,
	})
	if err != nil {
		respondOrgPositionError(c, err)
		return
	}
	recordAudit(c, model.AuditActionOrgPositionUpdated, "org_position", position.ID, model.JSONMap{
		"title": position.Title,
	})
	c.JSON(http.StatusOK, position)
}

func DeleteOrgPosition(c *gin.Context) {
	requireOrgWriteAccess(c)

	id := c.Param("id")
	position, err := service.GetOrgPositionByID(id)
	if err != nil {
		respondOrgPositionError(c, err)
		return
	}
	if err := service.DeleteOrgPosition(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	recordAudit(c, model.AuditActionOrgPositionDeleted, "org_position", id, model.JSONMap{
		"title": position.Title,
	})
	c.JSON(http.StatusOK, gin.H{"message": "position deleted"})
}

// respondOrgPositionError keeps validation failures as 4xx. A cycle or a blank
// title is the caller's mistake, and a 500 would send an admin hunting through
// server logs for a problem they can fix in the form.
func respondOrgPositionError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "position not found"})
	case errors.Is(err, service.ErrOrgPositionCycle),
		errors.Is(err, service.ErrOrgPositionTitleRequired):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
