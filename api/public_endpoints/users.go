package public_endpoints

import (
	"errors"
	"github.com/fusioncatltd/fusioncat/api"
	"github.com/fusioncatltd/fusioncat/api/input_contracts"
	"github.com/fusioncatltd/fusioncat/common"
	"github.com/fusioncatltd/fusioncat/logic"
	"github.com/gin-gonic/gin"
	"net/http"
)

func UsersPublicRouterV1(router *gin.RouterGroup) {
	router.POST("/users", UsersSignupAction)
}

// Sign up via email and password
// @Summary Sign up via email and password
// @Description Sign up via email and password with optional invitation code
// @Accept json
// @Produce json
// @Tags Public
// @Tags Users management
// @Param project body input_contracts.SignInSignUpApiInputContract true "Sign up request payload"
// @Param code query string false "Optional invitation code"
// @Success 200 {object} api_io_models.SignInSignUpApiInput "Successful sign up"
// @Success 422 {object} api_io_models.DataValidationErrorAPIResponse "JSON payload validation error"
// @Success 409 "User with specified email is already registered in the system"
// @Router /v1/public/users [post]
func UsersSignupAction(c *gin.Context) {

	var input input_contracts.SignInSignUpApiInputContract

	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, api.GetValidationErrors(err))
		return
	}

	usersManager := logic.UserObjectsManager{}
	userObject, err := usersManager.RegisterNewUserWithEmailAndPassword(input.Email, input.Password)

	if err != nil {
		if errors.Is(err, common.FusioncatErrUniqueConstraintViolations) {
			c.AbortWithStatusJSON(http.StatusConflict, nil)
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, nil)
		return
	}

	jwt, _ := common.GenerateGwtToken(userObject.GetID())

	c.Header("Authorization", "Bearer "+jwt)
	c.JSON(http.StatusOK, userObject.Serialize())
}
