package protected_endpoints

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func AuthenticationProtectedRoutesV1(router *gin.RouterGroup) {
	router.GET("/authentication", CheckAuthenticationStatus)
}

// Read personal information of user who owns the authentication token
// @Summary Read personal information of  user who owns the authentication token
// @Description Read personal information of  user who owns the authentication token
// @Accept json
// @Produce json
// @Tags Authentication related
// @Security BearerAuth
// @Success 200 "Empty response indicating successful authentication"
// @Success 401 "Access denied: missing or invalid Authorization header"
// @Router /v1/protected/authentication [get]
func CheckAuthenticationStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}
