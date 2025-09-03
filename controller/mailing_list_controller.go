package controller

import (
	"net/http"
	"sentinel/model"
	"sentinel/service"

	"github.com/gin-gonic/gin"
)

func AddEmailToMailingList(c *gin.Context) {
	var email model.MailingList

	if err := c.ShouldBindJSON(&email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: a valid email is required."})
		return
	}

	email, err := service.AddEmailToMailingList(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, email)

}

func GetAllMailingListEmails(c *gin.Context) {
	Require(c, Any(
		RequestTokenHasScope(c, "sentinel:all"),
		All(
			RequestTokenHasScope(c, "user:read"),
			RequestUserHasRole(c, "d_admin"),
		),
	))
	emails := service.GetAllMailingListEmails()
	c.JSON(http.StatusOK, emails)

}
