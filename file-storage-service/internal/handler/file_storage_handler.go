package handler

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/KEPTANy/plag-check/file-storage-service/internal/middleware"
	"github.com/KEPTANy/plag-check/file-storage-service/internal/service"
	"github.com/gofrs/uuid/v5"
)

type FileStorageHandler struct {
	FileStorageService service.FileStorageService
	MaxFileSize        int64
}

func NewFileStorageHandler(service service.FileStorageService) *FileStorageHandler {
	return &FileStorageHandler{FileStorageService: service}
}

func (h *FileStorageHandler) Upload(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	role, ok := middleware.GetRoleFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	if role != "student" {
		http.Error(w, `{"error": "only students can upload solutions"}`, http.StatusForbidden)
		return
	}

	if err := r.ParseMultipartForm(h.MaxFileSize); err != nil {
		http.Error(w, `{"error": "failed to parse form"}`, http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, `{"error": "file is required"}`, http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileData, err := h.FileStorageService.UploadFile(r.Context(), userID, file, header)
	if err != nil {
		http.Error(w, `{"error": "failed to upload file"}`, http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"file_info": fileData,
	})
}

func (h *FileStorageHandler) Download(w http.ResponseWriter, r *http.Request) {
	fileIDStr := r.PathValue("id")
	fileID, err := strconv.Atoi(fileIDStr)
	if err != nil {
		http.Error(w, `{"error": "invalid solution ID"}`, http.StatusBadRequest)
		return
	}

	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	role, ok := middleware.GetRoleFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	fileInfo, file, err := h.FileStorageService.DownloadFile(r.Context(), fileID)
	if err != nil {
		http.Error(w, `{"error": "forbiden access"}`, http.StatusForbidden)
		return
	}
	defer file.Close()

	if role != "teacher" && userID != fileInfo.StudentID {
		http.Error(w, `{"error": "forbiden access"}`, http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileInfo.Filename))
	w.Header().Set("Content-Length", strconv.FormatInt(fileInfo.FileSize, 10))

	io.Copy(w, file)
}

func (h *FileStorageHandler) ListUserFiles(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	role, ok := middleware.GetRoleFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	reqUserID, err := uuid.FromString((r.PathValue("userid")))
	if err != nil {
		http.Error(w, `{"error": "bad user id"}`, http.StatusBadRequest)
		return
	}

	if role != "teacher" && userID != reqUserID {
		http.Error(w, `{"error": "forbiden access"}`, http.StatusForbidden)
		return
	}

	files, err := h.FileStorageService.ListFilesByUser(r.Context(), reqUserID)
	if err != nil {
		http.Error(w, `{"error": "internal error"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"files": files,
	})
}

func (h *FileStorageHandler) ListFilesByHash(w http.ResponseWriter, r *http.Request) {
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

	hash := r.PathValue("hash")

	if role != "teacher" {
		http.Error(w, `{"error": "forbiden access"}`, http.StatusForbidden)
		return
	}

	files, err := h.FileStorageService.ListFilesByHash(r.Context(), hash)
	if err != nil {
		http.Error(w, `{"error": "internal error"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"files": files,
	})
}
