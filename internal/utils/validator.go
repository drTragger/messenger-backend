package utils

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/drTragger/messenger-backend/internal/models"
	"github.com/go-playground/validator/v10"
)

var (
	validate   *validator.Validate
	phoneRegex = regexp.MustCompile(`^\+?[1-9][0-9]{9,14}$`) // Regex for E.164 format or similar
)

func init() {
	validate = validator.New()
	err := validate.RegisterValidation("phone", validatePhoneNumber)
	if err != nil {
		log.Fatal(err)
	}
	err = validate.RegisterValidation("messageType", validateMessageType)
	if err != nil {
		log.Fatal(err)
	}
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

// validatePhoneNumber validates phone numbers for E.164 format of similar
func validatePhoneNumber(fl validator.FieldLevel) bool {
	return phoneRegex.MatchString(fl.Field().String())
}

// validateMessageType validates message type to fit models.MessageType
func validateMessageType(fl validator.FieldLevel) bool {
	messageType := fl.Field().String()
	switch models.MessageType(messageType) {
	case models.TextMessage, models.ImageMessage, models.VideoMessage, models.FileMessage, models.SystemMessage:
		return true
	default:
		return false
	}
}
