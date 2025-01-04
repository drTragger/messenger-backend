package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/drTragger/messenger-backend/internal/models"
	"github.com/drTragger/messenger-backend/internal/requests"
	"github.com/drTragger/messenger-backend/internal/responses"
	"github.com/drTragger/messenger-backend/internal/utils"
	"net/http"
	"time"

	"github.com/drTragger/messenger-backend/internal/repository"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	UserRepo *repository.UserRepository
	Secret   string
	Trans    *utils.Translator
}

func NewAuthHandler(repo *repository.UserRepository, secret string, trans *utils.Translator) *AuthHandler {
	return &AuthHandler{
		UserRepo: repo,
		Secret:   secret,
		Trans:    trans,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var payload requests.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		responses.ErrorResponse(w, http.StatusBadRequest, h.Trans.Translate(r, "errors.input", nil), err.Error())
		return
	}

	if err := utils.ValidateStruct(&payload); err != nil {
		responses.ValidationResponse(w, h.Trans.Translate(r, "errors.validation", nil), utils.FormatValidationError(r, err, h.Trans))
		return
	}

	userExists, err := h.UserRepo.GetUserByEmail(payload.Email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), err.Error())
		return
	}

	if userExists != nil {
		responses.ValidationResponse(w, h.Trans.Translate(r, "errors.validation", nil), map[string]string{
			"email": h.Trans.Translate(r, "validation.unique", nil),
		})
		return
	}

	usernameExists, err := h.UserRepo.GetUserByUsername(payload.Username)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), err.Error())
	}

	if usernameExists != nil {
		responses.ValidationResponse(w, h.Trans.Translate(r, "errors.validation", nil), map[string]string{
			"username": h.Trans.Translate(r, "validation.unique", nil),
		})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), err.Error())
		return
	}

	user := models.User{
		Username: payload.Username,
		Email:    payload.Email,
		Password: string(hashedPassword),
	}

	err = h.UserRepo.CreateUser(&user)
	if err != nil {
		responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), err.Error())
		return
	}

	responses.SuccessResponse(w, http.StatusCreated, h.Trans.Translate(r, "success.register", nil), nil)
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var credentials requests.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		responses.ErrorResponse(w, http.StatusBadRequest, h.Trans.Translate(r, "errors.input", nil), err.Error())
		return
	}

	if err := utils.ValidateStruct(&credentials); err != nil {
		responses.ValidationResponse(w, h.Trans.Translate(r, "errors.validation", nil), utils.FormatValidationError(r, err, h.Trans))
		return
	}

	user, err := h.UserRepo.GetUserByEmail(credentials.Email)
	if err != nil || user == nil {
		responses.ErrorResponse(w, http.StatusUnauthorized, h.Trans.Translate(r, "errors.credentials", nil), h.Trans.Translate(r, "errors.auth", nil))
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password)); err != nil {
		responses.ErrorResponse(w, http.StatusUnauthorized, h.Trans.Translate(r, "errors.credentials", nil), h.Trans.Translate(r, "errors.auth", nil))
		return
	}

	tokenExpire := time.Now().Add(24 * time.Hour).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     tokenExpire,
	})
	tokenString, err := token.SignedString([]byte(h.Secret))
	if err != nil {
		responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), err.Error())
		return
	}

	responses.SuccessResponse(w, http.StatusOK, h.Trans.Translate(r, "success.login", nil), responses.LoginResponse{
		Token:  tokenString,
		Expire: tokenExpire,
	})
}
