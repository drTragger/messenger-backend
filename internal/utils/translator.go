package utils

import (
	"encoding/json"
	"github.com/drTragger/messenger-backend/internal/middleware"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"log"
	"net/http"
)

const FallbackLang = "en"

type Translator struct {
	bundle *i18n.Bundle
}

// NewTranslator initializes the translation bundle and loads the translation files.
func NewTranslator() *Translator {
	bundle := i18n.NewBundle(language.English) // Default language
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	// Load translation files
	_, err := bundle.LoadMessageFile("./locales/en.json")
	if err != nil {
		log.Fatalf("Failed to load en.json: %v", err)
	}
	_, err = bundle.LoadMessageFile("./locales/uk.json")
	if err != nil {
		log.Fatalf("Failed to load uk.json: %v", err)
	}
	_, err = bundle.LoadMessageFile("./locales/pl.json")
	if err != nil {
		log.Fatalf("Failed to load pl.json: %v", err)
	}
	log.Println("Loaded translation files")

	return &Translator{bundle: bundle}
}

// Translate fetches the translation for the given language, message ID, and template data.
func (t *Translator) Translate(r *http.Request, messageID string, data map[string]interface{}) string {
	lang := GetLocale(r)
	localizer := i18n.NewLocalizer(t.bundle, lang)

	// Translate the main message
	translated, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: data, // Use only relevant dynamic data
	})
	if err != nil {
		return messageID // Fallback to message ID if translation is missing
	}
	return translated
}

// GetLocale retrieves the language from the request context.
func GetLocale(r *http.Request) string {
	lang, _ := r.Context().Value(middleware.LanguageKey).(string)
	if lang == "" {
		lang = FallbackLang // Default to English
	}
	return lang
}
