package public_endpoints

import (
	"errors"
	"github.com/fusioncatltd/fusioncat/api"
	"github.com/fusioncatltd/fusioncat/api/input_contracts"
	"github.com/fusioncatltd/fusioncat/common"
	"github.com/fusioncatltd/fusioncat/logic"

	//"github.com/fusioncatalyst/mono/server/api_io_models"
	//"github.com/fusioncatalyst/mono/server/common/authentication"
	//errors2 "github.com/fusioncatalyst/mono/server/common/errors"
	//"github.com/fusioncatalyst/mono/server/objects"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
)

func AuthenticationPublicRoutesV1(router *gin.RouterGroup) {
	router.POST("/authentication", AuthenticateViaCredentialsAction)
}

func AuthenticateViaCredentialsAction(c *gin.Context) {

	var input input_contracts.SignInSignUpApiInputContract

	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, api.GetValidationErrors(err))
		return
	}

	usersManager := logic.UserObjectsManager{}
	userObject, err := usersManager.FindByEmail(input.Email)
	if errors.Is(err, common.FusioncatErrRecordNotFound) {
		c.AbortWithStatusJSON(http.StatusUnauthorized, nil)
		c.Abort()
		return
	}

	passwordValidationResult := userObject.VerifyPassword(input.Password)
	if !passwordValidationResult {
		c.AbortWithStatusJSON(http.StatusUnauthorized, nil)
		return
	}

	jwt, _ := common.GenerateGwtToken(userObject.GetID())
	c.Header("Authorization", "Bearer "+jwt)

	// set cookie for frontend
	c.SetCookie(
		os.Getenv("COOKIE_NAME"),
		jwt,
		86400,
		"/",
		"",
		true,
		true,
	)

	c.JSON(http.StatusOK, userObject.Serialize())
}
