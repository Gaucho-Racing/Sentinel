package controller

import (
	"sentinel/utils"

	"github.com/gin-gonic/gin"
)

func NginxValidate(c *gin.Context) {
	// Dump all request info
	requestInfo := map[string]interface{}{
		"Method":     c.Request.Method,
		"URL":        c.Request.URL.String(),
		"Proto":      c.Request.Proto,
		"Header":     c.Request.Header,
		"Host":       c.Request.Host,
		"RemoteAddr": c.Request.RemoteAddr,
		"RequestURI": c.Request.RequestURI,
	}

	// Get query parameters
	queryParams := c.Request.URL.Query()
	if len(queryParams) > 0 {
		requestInfo["QueryParams"] = queryParams
	}

	// Get form data
	if err := c.Request.ParseForm(); err == nil {
		if len(c.Request.PostForm) > 0 {
			requestInfo["FormData"] = c.Request.PostForm
		}
	}

	utils.SugarLogger.Infof("Request info: %v", requestInfo)

	// Log or send the request info
	c.JSON(200, gin.H{
		"message":      "Request info dumped",
		"request_info": requestInfo,
	})
}

func NginxFail(c *gin.Context) {
	utils.SugarLogger.Errorf("Responding unauthenticated")
	c.JSON(401, gin.H{
		"message": "Unauthenticated",
	})
}

func NginxSuccess(c *gin.Context) {
	utils.SugarLogger.Infof("Responding authenticated")
	c.JSON(200, gin.H{
		"message": "Authenticated",
	})
}
