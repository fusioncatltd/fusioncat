package protected_endpoints

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func AuthenticationProtectedRoutesV1(router *gin.RouterGroup) {
	router.GET("/authentication", CheckAuthenticationStatus)
}

func CheckAuthenticationStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}
