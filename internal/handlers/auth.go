package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/drTragger/messenger-backend/internal/models"
	"github.com/drTragger/messenger-backend/internal/requests"
	"github.com/drTragger/messenger-backend/internal/responses"
	"github.com/drTragger/messenger-backend/internal/utils"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/drTragger/messenger-backend/internal/repository"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

const (
	TokenExpire            = 24 * time.Hour
	VerificationCodeLength = 6
	VerificationCodeExpire = 5 * time.Minute
)

type AuthHandler struct {
	UserRepo  *repository.UserRepository
	TokenRepo *repository.TokenRepository
	Secret    string
	Trans     *utils.Translator
}

func NewAuthHandler(
	ur *repository.UserRepository,
	tr *repository.TokenRepository,
	secret string,
	trans *utils.Translator,
) *AuthHandler {
	return &AuthHandler{
		UserRepo:  ur,
		TokenRepo: tr,
		Secret:    secret,
		Trans:     trans,
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

	userExists, err := h.UserRepo.GetUserByPhone(payload.Phone)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), err.Error())
		return
	}

	if userExists != nil {
		responses.ValidationResponse(w, h.Trans.Translate(r, "errors.validation", nil), map[string]string{
			"phone": h.Trans.Translate(r, "validation.unique", nil),
		})
		return
	}

	usernameExists, err := h.UserRepo.GetUserByUsername(payload.Username)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), err.Error())
		return
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
		Phone:    payload.Phone,
		Password: string(hashedPassword),
	}

	err = h.UserRepo.CreateUser(&user)
	if err != nil {
		responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), err.Error())
		return
	}

	// Generate verification code
	verificationCode := utils.GenerateRandomCode(VerificationCodeLength)

	// Create a channel to capture errors
	errChan := make(chan error, 2)

	// Use a goroutine to store the verification code in Redis
	go func() {
		errChan <- h.TokenRepo.StoreVerificationCode(r.Context(), user.Phone, verificationCode, VerificationCodeExpire)
	}()

	// Use another goroutine to send the SMS
	go func() {
		smsClient := utils.NewSMSClient()
		errChan <- smsClient.SendSMS(
			payload.Phone,
			h.Trans.Translate(r, "notifications.welcome", map[string]interface{}{
				"Username": user.Username,
				"Code":     verificationCode,
				"Expires":  VerificationCodeExpire.Minutes(),
			}),
		)
	}()

	// Wait for both operations to complete
	for i := 0; i < 2; i++ {
		if err := <-errChan; err != nil {
			responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), err.Error())
			return
		}
	}

	responses.SuccessResponse(w, http.StatusCreated, h.Trans.Translate(r, "success.register", nil), nil)
}

// VerifyCode verifies phone number verification code
func (h *AuthHandler) VerifyCode(w http.ResponseWriter, r *http.Request) {
	var payload requests.VerifyCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		responses.ErrorResponse(w, http.StatusBadRequest, h.Trans.Translate(r, "errors.input", nil), err.Error())
		return
	}

	if err := utils.ValidateStruct(&payload); err != nil {
		responses.ValidationResponse(w, h.Trans.Translate(r, "errors.validation", nil), utils.FormatValidationError(r, err, h.Trans))
		return
	}

	// Retrieve the code from Redis
	ctx := r.Context()
	storedCode, err := h.TokenRepo.GetVerificationCode(ctx, payload.Phone)
	if err != nil || storedCode != payload.Code {
		responses.ErrorResponse(w, http.StatusBadRequest, h.Trans.Translate(r, "errors.code.invalid", nil), "Invalid verification code.")
		return
	}

	// Create a channel to capture errors from goroutines
	errChan := make(chan error, 2)

	// Use a goroutine to mark the phone as verified in the database
	go func() {
		if err := h.UserRepo.VerifyPhone(payload.Phone); err != nil {
			errChan <- err
		} else {
			errChan <- nil
		}
	}()

	// Use another goroutine to delete the verification code from Redis
	go func() {
		if err := h.TokenRepo.DeleteVerificationCode(ctx, payload.Phone); err != nil {
			log.Printf("Failed to delete verification code: %s", err)
		}
		errChan <- nil
	}()

	// Wait for both operations to complete and check for errors
	for i := 0; i < 2; i++ {
		if err := <-errChan; err != nil {
			responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), "An error occurred while processing the request.")
			return
		}
	}

	responses.SuccessResponse(w, http.StatusOK, h.Trans.Translate(r, "success.phone_verification", nil), nil)
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var payload requests.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		responses.ErrorResponse(w, http.StatusBadRequest, h.Trans.Translate(r, "errors.input", nil), err.Error())
		return
	}

	if err := utils.ValidateStruct(&payload); err != nil {
		responses.ValidationResponse(w, h.Trans.Translate(r, "errors.validation", nil), utils.FormatValidationError(r, err, h.Trans))
		return
	}

	user, err := h.UserRepo.GetUserByPhone(payload.Phone)
	if err != nil || user == nil {
		responses.ErrorResponse(w, http.StatusUnauthorized, h.Trans.Translate(r, "errors.credentials", nil), h.Trans.Translate(r, "errors.auth", nil))
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password)); err != nil {
		responses.ErrorResponse(w, http.StatusUnauthorized, h.Trans.Translate(r, "errors.credentials", nil), h.Trans.Translate(r, "errors.auth", nil))
		return
	}

	if user.PhoneVerifiedAt == nil {
		responses.ErrorResponse(w, http.StatusForbidden, h.Trans.Translate(r, "errors.code.unverified", nil), "Phone not verified.")
		return
	}

	tokenExpire := time.Now().Add(TokenExpire).Unix()
	tokenString, err := utils.GenerateJWT(h.Secret, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     tokenExpire,
	})
	if err != nil {
		responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), err.Error())
		return
	}

	// Store token in Redis concurrently
	errChan := make(chan error, 1)
	go func() {
		errChan <- h.TokenRepo.StoreToken(r.Context(), tokenString, user.ID, TokenExpire)
	}()

	if err := <-errChan; err != nil {
		responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), "Failed to store token.")
		return
	}

	responses.SuccessResponse(w, http.StatusOK, h.Trans.Translate(r, "success.login", nil), responses.TokenResponse{
		Token:   tokenString,
		Expires: tokenExpire,
	})
}

// RefreshToken refreshes JWT token
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var payload requests.RefreshTokenRequest

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		responses.ErrorResponse(w, http.StatusBadRequest, h.Trans.Translate(r, "errors.input", nil), err.Error())
		return
	}

	if err := utils.ValidateStruct(&payload); err != nil {
		responses.ValidationResponse(w, h.Trans.Translate(r, "errors.validation", nil), utils.FormatValidationError(r, err, h.Trans))
		return
	}

	token, err := jwt.Parse(payload.Token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New(h.Trans.Translate(r, "errors.token.signing_method", nil))
		}
		return []byte(h.Secret), nil
	})
	if err != nil || !token.Valid {
		responses.ErrorResponse(w, http.StatusUnauthorized, h.Trans.Translate(r, "errors.token.invalid", nil), "Invalid token")
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		responses.ErrorResponse(w, http.StatusUnauthorized, h.Trans.Translate(r, "errors.token.invalid", nil), "Invalid token claims")
		return
	}

	exp, ok := claims["exp"].(float64)
	if !ok || float64(time.Now().Unix()) > exp {
		responses.ErrorResponse(w, http.StatusUnauthorized, h.Trans.Translate(r, "errors.token.expired", nil), "Token expired")
		return
	}

	userID := int(claims["user_id"].(float64))

	// Concurrently delete old token and store new token
	errChan := make(chan error, 2)
	go func() {
		errChan <- h.TokenRepo.DeleteToken(r.Context(), payload.Token, userID)
	}()

	newTokenExpire := time.Now().Add(TokenExpire).Unix()
	newTokenString, err := utils.GenerateJWT(h.Secret, jwt.MapClaims{
		"user_id": userID,
		"exp":     newTokenExpire,
	})
	if err != nil {
		responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), err.Error())
		return
	}
	go func() {
		errChan <- h.TokenRepo.StoreToken(r.Context(), newTokenString, userID, TokenExpire)
	}()

	// Wait for both operations to complete
	for i := 0; i < 2; i++ {
		if err := <-errChan; err != nil {
			responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), "Failed to update token.")
			return
		}
	}

	responses.SuccessResponse(w, http.StatusOK, h.Trans.Translate(r, "success.refresh_token", nil), responses.TokenResponse{
		Token:   newTokenString,
		Expires: newTokenExpire,
	})
}

// Logout handles user logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	tokenString := strings.TrimSpace(strings.Replace(r.Header.Get("Authorization"), "Bearer", "", 1))
	if tokenString == "" {
		responses.ErrorResponse(w, http.StatusUnauthorized, h.Trans.Translate(r, "errors.unauthorized", nil), "Token not provided")
		return
	}

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New(h.Trans.Translate(r, "errors.token.signing_method", nil))
		}
		return []byte(h.Secret), nil
	})
	if err != nil || !token.Valid {
		responses.ErrorResponse(w, http.StatusUnauthorized, h.Trans.Translate(r, "errors.token.invalid", nil), "Invalid token")
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		responses.ErrorResponse(w, http.StatusUnauthorized, h.Trans.Translate(r, "errors.token.invalid", nil), "Invalid token claims")
		return
	}

	userID := int(claims["user_id"].(float64))

	// Delete the token from Redis concurrently
	errChan := make(chan error, 1)
	go func() {
		errChan <- h.TokenRepo.DeleteToken(r.Context(), tokenString, userID)
	}()

	if err := <-errChan; err != nil {
		responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), "Failed to delete token.")
		return
	}

	responses.SuccessResponse(w, http.StatusOK, h.Trans.Translate(r, "success.logout", nil), nil)
}
