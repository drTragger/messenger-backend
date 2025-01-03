package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
)

func RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/lifecheck", func(w http.ResponseWriter, r *http.Request) {
		SuccessResponse(w, http.StatusOK, "Messenger app is running", nil)
	}).Methods("GET")
}
