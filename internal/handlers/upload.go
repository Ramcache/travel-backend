package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/Ramcache/travel-backend/internal/helpers"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Ramcache/travel-backend/internal/models"
)

type UploadHandler struct {
	BaseURL     string
	UploadDir   string
	MaxUploadMB int // лимит на один запрос
}

func NewUploadHandler(baseURL, uploadDir string, maxUploadMB int) *UploadHandler {
	return &UploadHandler{
		BaseURL:     strings.TrimRight(baseURL, "/"),
		UploadDir:   uploadDir,
		MaxUploadMB: maxUploadMB,
	}
}

// Upload
// @Summary Upload files and get URLs
// @Tags upload
// @Accept multipart/form-data
// @Produce json
// @Param files formData []file true "Multiple files"
// @Success 200 {object} models.UploadResponse
// @Failure 400 {object} helpers.ErrorData
// @Failure 413 {object} helpers.ErrorData
// @Failure 500 {object} helpers.ErrorData
// @Router /api/v1/admin/upload [post]
func (h *UploadHandler) Upload(w http.ResponseWriter, r *http.Request) {
	// Ограничим общий размер тела запроса
	maxBytes := int64(h.MaxUploadMB) << 20 // MB -> bytes
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

	// Парсим multipart
	if err := r.ParseMultipartForm(maxBytes); err != nil {
		var msg = "invalid multipart form"
		if errors.Is(err, http.ErrMissingBoundary) || strings.Contains(err.Error(), "http: request body too large") {
			http.Error(w, "payload too large", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	// Поддержим и "files", и "file"
	files := r.MultipartForm.File["files"]
	if len(files) == 0 && r.MultipartForm.File["file"] != nil {
		files = r.MultipartForm.File["file"]
	}
	if len(files) == 0 {
		http.Error(w, "no files uploaded", http.StatusBadRequest)
		return
	}

	// ensure dir
	if err := os.MkdirAll(h.UploadDir, 0o755); err != nil {
		http.Error(w, "cannot create upload dir", http.StatusInternalServerError)
		return
	}

	var urls []string
	for _, fh := range files {
		src, err := fh.Open()
		if err != nil {
			http.Error(w, "cannot open uploaded file", http.StatusInternalServerError)
			return
		}
		defer src.Close()

		// Генерируем безопасное уникальное имя с сохранением расширения
		ext := strings.ToLower(filepath.Ext(fh.Filename))
		safeName := randomName() + ext
		dstPath := filepath.Join(h.UploadDir, safeName)

		dst, err := os.Create(dstPath)
		if err != nil {
			http.Error(w, "cannot create destination file", http.StatusInternalServerError)
			return
		}

		if _, err := io.Copy(dst, src); err != nil {
			dst.Close()
			_ = os.Remove(dstPath)
			http.Error(w, "cannot save file", http.StatusInternalServerError)
			return
		}
		_ = dst.Close()

		urls = append(urls, h.absURL("/uploads/"+safeName, r))
	}

	helpers.JSON(w, http.StatusOK, models.UploadResponse{URLs: urls})
}

func (h *UploadHandler) absURL(path string, r *http.Request) string {
	if h.BaseURL != "" {
		return h.BaseURL + path
	}
	// fallback: собрать по факту запроса (если не задан BaseURL)
	scheme := "http"
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	} else if r.TLS != nil {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s%s", scheme, r.Host, path)
}

func randomName() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%d-%s", time.Now().UnixNano(), hex.EncodeToString(b))
}

// Delete — удаляет один файл по имени
// @Summary Delete uploaded file
// @Tags upload
// @Param filename path string true "File name"
// @Success 200 {string} string "deleted"
// @Failure 404 {string} string "not found"
// @Router /api/v1/admin/upload/{filename} [delete]
func (h *UploadHandler) Delete(w http.ResponseWriter, r *http.Request) {
	filename := chi.URLParam(r, "filename")
	if filename == "" {
		http.Error(w, "filename required", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(h.UploadDir, filepath.Base(filename))
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	if err := os.Remove(filePath); err != nil {
		http.Error(w, "cannot delete file", http.StatusInternalServerError)
		return
	}

	helpers.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// Update — заменяет существующий файл новым
// @Summary Replace uploaded file
// @Tags upload
// @Accept multipart/form-data
// @Produce json
// @Param filename path string true "Old file name"
// @Param file formData file true "New file"
// @Success 200 {object} models.UploadResponse
// @Router /api/v1/admin/upload/{filename} [put]
func (h *UploadHandler) Update(w http.ResponseWriter, r *http.Request) {
	filename := chi.URLParam(r, "filename")
	if filename == "" {
		http.Error(w, "filename required", http.StatusBadRequest)
		return
	}

	oldPath := filepath.Join(h.UploadDir, filepath.Base(filename))
	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	// Парсим новый файл
	if err := r.ParseMultipartForm(int64(h.MaxUploadMB) << 20); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Перезаписываем файл
	dst, err := os.Create(oldPath)
	if err != nil {
		http.Error(w, "cannot overwrite file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "write error", http.StatusInternalServerError)
		return
	}

	url := h.absURL("/uploads/"+filepath.Base(filename), r)
	helpers.JSON(w, http.StatusOK, models.UploadResponse{URLs: []string{url}})
}
