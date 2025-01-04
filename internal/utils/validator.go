package utils

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// ValidateStruct validates a struct based on its tags
func ValidateStruct(input interface{}) error {
	return validate.Struct(input)
}

// FormatValidationError formats validation errors into a user-friendly translated message
func FormatValidationError(r *http.Request, err error, translator *Translator) map[string]string {
	var validationErrors validator.ValidationErrors
	ok := errors.As(err, &validationErrors)
	fieldErrors := make(map[string]string)

	if !ok {
		// Add a general error if it's not a validation error
		fieldErrors["general"] = translator.Translate(r, "errors.input", nil)
		return fieldErrors
	}

	for _, fieldErr := range validationErrors {
		messageID := fmt.Sprintf("validation.%s", fieldErr.Tag())
		// Use only the tag for translation; field names are kept as-is
		fieldErrors[strings.ToLower(fieldErr.Field())] = translator.Translate(r, messageID, map[string]interface{}{
			"Param": fieldErr.Param(), // Keep only relevant dynamic data
		})
	}
	return fieldErrors
}
