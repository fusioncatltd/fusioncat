package protected_endpoints

import (
	"github.com/fusioncatltd/fusioncat/logic"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

func MeProtectedRoutesV1(router *gin.RouterGroup) {
	router.GET("/me", GetMyProfileAction)
}

// Read personal information of  user who owns the authentication token
// @Summary Read personal information of  user who owns the authentication token
// @Description Read personal information of  user who owns the authentication token
// @Accept json
// @Produce json
// @Tags Personal information
// @Success 200 {object} objects.UserDBSerializerStruct "User information
// @Success 401 "Access denied: missing or invalid Authorization header"
// @Router /v1/protected/me [get]
func GetMyProfileAction(c *gin.Context) {
	userID, _ := c.Get("UserID")
	usersManager := logic.UserObjectsManager{}
	userObject, err := usersManager.FindByID(userID.(uuid.UUID))

	if err != nil {
		panic("User is not found: " + err.Error())
	}

	c.JSON(http.StatusOK, userObject.Serialize())
}
