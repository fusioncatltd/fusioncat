package input_contracts

import (
	"github.com/fusioncatltd/fusioncat/common"
	"github.com/go-playground/validator/v10"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

// This validator checks if the field is a valid JSON schema.
var ValidJSONSchemaValidator validator.Func = func(fl validator.FieldLevel) bool {
	f := fl.Field().String()

	_, exists, schemaExtractionError := common.ExtractSchemaField(f)
	if schemaExtractionError != nil {
		return false
	}

	if !exists {
		return false
	}

	_, err := jsonschema.CompileString("", f)
	if err != nil {
		return false
	}
	return true
}

// Most of the names in Fusioncat are alphanumeric with underscores, but some names (like Kafka topics)
// need to include dots too.
var ValidateAlphanumWithUnderscoreAndDots validator.Func = func(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	for _, char := range value {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '_' ||
			char == '.') {
			return false
		}
	}
	return true
}

// And this validator checks if the field is alphanumeric with underscores only.
var ValidateAlphanumWithUnderscore validator.Func = func(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	for _, char := range value {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '_') {
			return false
		}
	}
	return true
}
