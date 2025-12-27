package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/KEPTANy/plag-check/user-service/internal/model"
	"github.com/KEPTANy/plag-check/user-service/internal/service"
)

type UserHandler struct {
	UserService service.UserService
}

func NewUserHandler(service service.UserService) *UserHandler {
	return &UserHandler{UserService: service}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "invalid request body",
		})
		return
	}

	if req.Username == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "username must not be empty",
		})
		return
	}

	if req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "password must not be empty",
		})
		return
	}

	if !model.IsValidRole(req.Role) {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "invalid role",
		})
		return
	}

	err := h.UserService.Register(r.Context(), &req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "failed to register a user, try another username",
		})
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "invalid request body",
		})
		return
	}

	if req.Username == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "username must not be empty",
		})
		return
	}

	if req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "password must not be empty",
		})
		return
	}

	if req.DurationMin <= 0 || req.DurationMin > 24*60 {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "duration must be in a range of 1 to 1440 minutes",
		})
		return
	}

	user, err := h.UserService.Login(r.Context(), &req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "failed to log in",
		})
		log.Printf("Failed to log in: %v", err)
		return
	}

	writeJSON(w, http.StatusOK, user)
}
