package handler

import (
	"log"
	"net/http"
	"strconv"

	"github.com/KEPTANy/plag-check/analysis-service/internal/middleware"
	"github.com/KEPTANy/plag-check/analysis-service/internal/service"
)

type AnalysisHandler struct {
	AnalysisService service.AnalysisService
}

func NewAnalysisHandler(service service.AnalysisService) *AnalysisHandler {
	return &AnalysisHandler{AnalysisService: service}
}

func (h *AnalysisHandler) CheckPlagiarism(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	role, ok := middleware.GetRoleFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	if role != "teacher" {
		http.Error(w, `{"error": "only teachers can check plagiarism"}`, http.StatusForbidden)
		return
	}

	results, err := h.AnalysisService.CheckPlagiarism(r.Context())
	if err != nil {
		log.Printf("Error checking plagiarism: %v", err)
		http.Error(w, `{"error": "internal error"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"plagiarism_results": results,
	})
}

func (h *AnalysisHandler) GetWordCloud(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	role, ok := middleware.GetRoleFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	if role != "teacher" {
		http.Error(w, `{"error": "only teachers can generate word clouds"}`, http.StatusForbidden)
		return
	}

	fileIDStr := r.PathValue("id")
	fileID, err := strconv.Atoi(fileIDStr)
	if err != nil {
		http.Error(w, `{"error": "invalid file ID"}`, http.StatusBadRequest)
		return
	}

	imageData, err := h.AnalysisService.GetWordCloud(r.Context(), fileID)
	if err != nil {
		log.Printf("Error generating word cloud: %v", err)
		http.Error(w, `{"error": "failed to generate word cloud"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.WriteHeader(http.StatusOK)
	w.Write(imageData)
}
