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

	_ = e.POST("/v1/public/users").
		WithJSON(signInPayloadWithNewEmail).
		Expect().
		Status(http.StatusOK).
		Header("Authorization").NotEmpty()

}
