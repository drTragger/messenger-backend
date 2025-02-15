package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/drTragger/messenger-backend/internal/repository"
	"github.com/drTragger/messenger-backend/internal/requests"
	"github.com/drTragger/messenger-backend/internal/responses"
	"github.com/drTragger/messenger-backend/internal/storage"
	"github.com/drTragger/messenger-backend/internal/utils"
	"github.com/drTragger/messenger-backend/internal/websocket"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

const (
	MaxPictureSizeMB = 10
)

type UserHandler struct {
	UserRepo      *repository.UserRepository
	ClientManager *websocket.ClientManager
	Storage       storage.Storage
	Trans         *utils.Translator
}

func NewUserHandler(
	userRepo *repository.UserRepository,
	clientManager *websocket.ClientManager,
	storage storage.Storage,
	trans *utils.Translator,
) *UserHandler {
	return &UserHandler{
		UserRepo:      userRepo,
		ClientManager: clientManager,
		Storage:       storage,
		Trans:         trans,
	}
}

func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")

	if query == "" {
		responses.ValidationResponse(w, h.Trans.Translate(r, "errors.validation", nil), map[string]string{
			"query": h.Trans.Translate(r, "validation.required", nil),
		})
		return
	}

	users, err := h.UserRepo.GetUsersBySearch(query)
	if err != nil {
		responses.ErrorResponse(w, http.StatusInternalServerError, "errors.server", err.Error())
		return
	}

	if len(users) == 0 {
		responses.ErrorResponse(w, http.StatusNotFound, h.Trans.Translate(r, "errors.not_found", nil), "Users not found")
		return
	}

	onlineUsers := h.ClientManager.GetOnlineUsers()

	falsePtr := new(bool)
	*falsePtr = false

	for _, user := range users {
		if status, exists := onlineUsers[user.ID]; exists {
			user.IsOnline = &status.IsOnline
		} else {
			user.IsOnline = falsePtr
		}
	}

	responses.SuccessResponse(w, http.StatusOK, h.Trans.Translate(r, "success.user.search", nil), users)
}

func (h *UserHandler) UpdateProfilePicture(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uint)

	err := r.ParseMultipartForm(MaxPictureSizeMB << 20) // 10MB max file size
	if err != nil {
		responses.ValidationResponse(w, h.Trans.Translate(r, "errors.input", nil), map[string]string{
			"picture": h.Trans.Translate(r, "validation.size", map[string]interface{}{"Param": MaxPictureSizeMB}),
		})
		return
	}

	// Get the file from the form
	file, handler, err := r.FormFile("picture")
	if err != nil {
		responses.ErrorResponse(w, http.StatusBadRequest, h.Trans.Translate(r, "errors.input", nil), err.Error())
		return
	}
	defer file.Close()

	validationRequest := requests.ProfilePictureRequest{
		Picture: handler.Header.Get("Content-Type"),
	}

	// Validate the input
	if err := utils.ValidateStruct(validationRequest); err != nil {
		fieldErrors := utils.FormatValidationError(r, err, h.Trans)
		responses.ValidationResponse(w, h.Trans.Translate(r, "errors.validation", nil), fieldErrors)
		return
	}

	// Retrieve current user profile
	user, err := h.UserRepo.GetUserByID(userID)
	if err != nil {
		responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), err.Error())
		return
	}

	if user == nil {
		responses.ErrorResponse(w, http.StatusNotFound, h.Trans.Translate(r, "errors.not_found", nil), "User not found")
		return
	}

	// Delete old profile picture if it exists
	if user.ProfilePicture != nil {
		err = h.Storage.DeleteFile(storage.ProfilePicturesDir, *user.ProfilePicture)
		if err != nil {
			log.Printf("Failed to delete old profile picture %s: %v", *user.ProfilePicture, err)
		}
	}

	// Use LocalStorage to save the file
	filePath, err := h.Storage.SaveFile(storage.ProfilePicturesDir, handler.Filename, file)
	if err != nil {
		responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), err.Error())
		return
	}

	// Update the user's profile picture in the database
	err = h.UserRepo.UpdateProfilePicture(userID, &filePath)
	if err != nil {
		responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), err.Error())
		return
	}
	user.ProfilePicture = &filePath

	// Respond with success
	responses.SuccessResponse(w, http.StatusOK, h.Trans.Translate(r, "success.user.update_picture", nil), user)
}

func (h *UserHandler) DeleteProfilePicture(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uint)

	// Retrieve current user profile
	user, err := h.UserRepo.GetUserByID(userID)
	if err != nil {
		responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), "Failed to retrieve user")
		return
	}

	if user == nil {
		responses.ErrorResponse(w, http.StatusNotFound, h.Trans.Translate(r, "errors.not_found", nil), "User not found")
		return
	}

	// Delete old profile picture if it exists
	if user.ProfilePicture != nil {
		err = h.Storage.DeleteFile(storage.ProfilePicturesDir, *user.ProfilePicture)
		if err != nil {
			responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), err.Error())
			return
		}
	}

	// Update the user's profile picture in the database
	err = h.UserRepo.UpdateProfilePicture(userID, nil)
	if err != nil {
		responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), err.Error())
		return
	}

	responses.SuccessResponse(w, http.StatusOK, h.Trans.Translate(r, "success.user.delete_picture", nil), nil)
}

func (h *UserHandler) GetProfilePicture(w http.ResponseWriter, r *http.Request) {
	fileName := mux.Vars(r)["filename"]
	if fileName == "" {
		responses.ValidationResponse(w, h.Trans.Translate(r, "errors.validation", nil), map[string]string{
			"file": h.Trans.Translate(r, "validation.required", nil),
		})
		return
	}

	// Get the file from storage
	filePath, err := h.Storage.GetFile(storage.ProfilePicturesDir, fileName)
	if err != nil {
		responses.ErrorResponse(w, http.StatusNotFound, h.Trans.Translate(r, "errors.not_found", nil), fmt.Sprintf("File not found: %v", err))
		return
	}

	responses.ServeFileResponse(w, r, filePath)
}

func (h *UserHandler) ChangePersonalInfo(w http.ResponseWriter, r *http.Request) {
	var payload requests.ChangePersonalInfoRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		responses.ErrorResponse(w, http.StatusBadRequest, h.Trans.Translate(r, "errors.input", nil), err.Error())
		return
	}

	if err := utils.ValidateStruct(&payload); err != nil {
		responses.ValidationResponse(w, h.Trans.Translate(r, "errors.validation", nil), utils.FormatValidationError(r, err, h.Trans))
		return
	}

	userID := r.Context().Value("user_id").(uint)

	user, err := h.UserRepo.GetUserByID(userID)
	if err != nil {
		responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), err.Error())
		return
	}
	if user == nil {
		responses.ErrorResponse(w, http.StatusNotFound, h.Trans.Translate(r, "errors.not_found", nil), fmt.Sprintf("User not found"))
		return
	}

	err = h.UserRepo.ChangePersonalInfo(userID, payload.FirstName, payload.LastName)
	if err != nil {
		responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), err.Error())
		return
	}
	user.FirstName = &payload.FirstName
	user.LastName = payload.LastName

	responses.SuccessResponse(w, http.StatusOK, h.Trans.Translate(r, "success.user.change_personal_info", nil), user)
}
