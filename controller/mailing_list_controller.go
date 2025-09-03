package controller

import (
	"net/http"
	"sentinel/model"
	"sentinel/service"

	"github.com/gin-gonic/gin"
)

func CreateMailingListEntry(c *gin.Context) {
	var entry model.MailingList

	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: a valid email is required."})
		return
	}

	entry, err := service.CreateMailingListEntry(entry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entry)

}

func GetAllMailingListEntries(c *gin.Context) {
	Require(c, Any(
		RequestTokenHasScope(c, "sentinel:all"),
		All(
			RequestTokenHasScope(c, "user:read"),
			RequestUserHasRole(c, "d_admin"),
		),
	))
	entries := service.GetAllMailingListEntries()
	c.JSON(http.StatusOK, entries)

}
