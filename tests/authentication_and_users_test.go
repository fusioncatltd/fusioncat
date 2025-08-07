package tests

import (
	"fmt"
	"github.com/fusioncatltd/fusioncat/api/input_contracts"
	"github.com/gavv/httpexpect/v2"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestSignUpOfNewUser(t *testing.T) {
	h := os.Getenv("TESTSERVER_URL")
	e := httpexpect.Default(t, h)

	uniqueEmail := fmt.Sprintf("test-email-%s@mail.com", strconv.FormatInt(time.Now().UnixNano(), 10))
	signInPayload := input_contracts.SignInSignUpApiInputContract{
		Email:    uniqueEmail,
		Password: "123456789",
	}

	// Checking authentication status before creating a user
	_ = e.GET("/v1/protected/authentication").
		WithJSON(signInPayload).
		Expect().
		Status(http.StatusUnauthorized)

	// Attempt to create a new user
	_ = e.POST("/v1/public/users").
		WithJSON(signInPayload).
		Expect().
		Status(http.StatusOK)

	// Attempt to create the same user again
	_ = e.POST("/v1/public/users").
		WithJSON(signInPayload).
		Expect().
		Status(http.StatusConflict)

	newUniqueEmail := fmt.Sprintf("test-email-%s@mail.com", strconv.FormatInt(time.Now().UnixNano(), 10))
	signInPayloadWithNewEmail := input_contracts.SignInSignUpApiInputContract{
		Email:    newUniqueEmail,
		Password: "123456789",
	}

	// Attempt to create a new user with a different email
	secondSignUpResponse := e.POST("/v1/public/users").
		WithJSON(signInPayloadWithNewEmail).
		Expect().
		Status(http.StatusOK)

	secondSignUpResponse.Header("Authorization").NotEmpty()
	bearer := secondSignUpResponse.Raw().Header.Get("Authorization")

	// When calling with the bearer token, the authentication status should be OK
	_ = e.GET("/v1/protected/authentication").
		WithHeader("Authorization", bearer).
		Expect().
		Status(http.StatusOK)

	// When calling without it, the authentication status should be Unauthorized
	_ = e.GET("/v1/protected/authentication").
		WithJSON(signInPayload).
		Expect().
		Status(http.StatusUnauthorized)
}
