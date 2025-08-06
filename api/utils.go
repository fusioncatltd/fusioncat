package api

import (
	"github.com/go-playground/validator/v10"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"strings"
)

type APIDataFieldErrorResponseField struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// DataValidationErrorAPIResponse is the response structure for validation errors
type DataValidationErrorAPIResponse struct {
	Errors []APIDataFieldErrorResponseField `json:"error"`
}

// Convert validator.FieldError to a user-friendly error message
func getErrorMsg(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "alphanum":
		return "Only alphanumeric characters are allowed"
	case "alphanum_with_underscore":
		return "Only alphanumeric characters and underscores are allowed"
	case "alphanum_with_underscore_and_dots":
		return "Only alphanumeric characters, dots and underscores are allowed"
	case "lte":
		return "Should be less than " + fe.Param()
	case "gte":
		return "Should be greater than " + fe.Param()
	case "email":
		return "This field is not a valid email"
	case "contains_valid_stringified_json":
		return "This field should contain valid stringified JSONs"
		return "Invalid reference to schema or to schmea version"
	case "valid_json_schema":
		f := fe.Value().(string)

		// For some reason it's not possible to return the error message
		// from validation of JSON schema, so we need to compile it again
		// to get the error message with all the details
		_, err := jsonschema.CompileString("", f)
		if err != nil {

			originalString := err.Error()
			substring := "compilation failed:"
			errorMessage := ""
			index := strings.Index(originalString, substring)
			if index != -1 {
				// Add the length of the substring to skip it too
				result := originalString[index+len(substring):]
				errorMessage = strings.TrimSpace(result)
			} else {
				errorMessage = originalString
			}

			return "Invalid JSON schema: " + errorMessage
		}

		return "Invalid JSON schema"
	}
	return "Unknown error"
}

func GetValidationErrors(err error) DataValidationErrorAPIResponse {
	var errors []APIDataFieldErrorResponseField
	for _, err := range err.(validator.ValidationErrors) {
		errors = append(errors, APIDataFieldErrorResponseField{
			Field:   err.Field(),
			Message: getErrorMsg(err),
		})
	}

	response := DataValidationErrorAPIResponse{Errors: errors}
	return response
}
