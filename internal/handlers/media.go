package handlers

import (
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Ramcache/travel-backend/internal/config"
	"github.com/Ramcache/travel-backend/internal/helpers"
)

type MediaHandler struct {
	cfg *config.Config
	db  *pgxpool.Pool
}

func NewMediaHandler(cfg *config.Config, db *pgxpool.Pool) *MediaHandler {
	return &MediaHandler{cfg: cfg, db: db}
}

// Upload
// @Summary Upload one or multiple photos
// @Tags Media
// @Accept multipart/form-data
// @Produce json
// @Param files formData file true "Фото (можно несколько)"
// @Success 200 {array} map[string]string
// @Router /admin/upload [post]
func (h *MediaHandler) Upload(w http.ResponseWriter, r *http.Request) {
	maxSize := int64(h.cfg.MaxUploadMB) << 20 // MB → bytes
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)

	if err := r.ParseMultipartForm(maxSize); err != nil {
		http.Error(w, "cannot parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		http.Error(w, "no files uploaded", http.StatusBadRequest)
		return
	}

	if err := os.MkdirAll(h.cfg.UploadDir, 0755); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var urls []map[string]string
	for _, fh := range files {
		filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(fh.Filename))
		dst := filepath.Join(h.cfg.UploadDir, filename)

		src, err := fh.Open()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer src.Close()

		out, err := os.Create(dst)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer out.Close()

		if _, err := io.Copy(out, src); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		url := fmt.Sprintf("%s/uploads/%s", h.cfg.AppBaseURL, filename)
		urls = append(urls, map[string]string{"url": url})
	}

	helpers.JSON(w, http.StatusOK, urls)
}

// ListUploads
// @Summary List all uploaded files
// @Tags Media
// @Produce json
// @Success 200 {array} map[string]string
// @Router /admin/uploads [get]
func (h *MediaHandler) ListUploads(w http.ResponseWriter, r *http.Request) {
	files, err := os.ReadDir(h.cfg.UploadDir)
	if err != nil {
		http.Error(w, "failed to read upload directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var result []map[string]string
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		url := fmt.Sprintf("%s/uploads/%s", h.cfg.AppBaseURL, f.Name())
		info, _ := f.Info()
		result = append(result, map[string]string{
			"name":      f.Name(),
			"url":       url,
			"size":      fmt.Sprintf("%.2f KB", float64(info.Size())/1024),
			"createdAt": info.ModTime().Format(time.RFC3339),
		})
	}
	helpers.JSON(w, http.StatusOK, result)
}

// DeleteUpload
// @Summary Delete uploaded file
// @Tags Media
// @Param file query string true "File name"
// @Success 200 {object} map[string]string
// @Router /admin/upload [delete]
func (h *MediaHandler) DeleteUpload(w http.ResponseWriter, r *http.Request) {
	fileName := r.URL.Query().Get("file")
	if fileName == "" {
		http.Error(w, "missing ?file=", http.StatusBadRequest)
		return
	}

	// защита от выхода за директорию
	if strings.Contains(fileName, "..") || strings.Contains(fileName, "/") || strings.Contains(fileName, "\\") {
		http.Error(w, "invalid filename", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(h.cfg.UploadDir, fileName)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	if err := os.Remove(filePath); err != nil {
		http.Error(w, "failed to delete file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	helpers.JSON(w, http.StatusOK, map[string]string{
		"message": "deleted successfully",
		"file":    fileName,
	})
}

// CleanupUnused
// @Summary Cleanup unused uploaded files
// @Tags Media
// @Success 200 {object} map[string]int
// @Router /admin/media/cleanup [post]
func (h *MediaHandler) CleanupUnused(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rows, err := h.db.Query(ctx, `
		SELECT unnest(urls) FROM trips
		UNION
		SELECT unnest(urls) FROM hotels
		UNION
		SELECT unnest(urls) FROM news
	`)
	if err != nil {
		http.Error(w, "failed to collect used URLs: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	used := make(map[string]struct{})
	for rows.Next() {
		var url string
		rows.Scan(&url)
		// Приводим к относительному виду /uploads/file.jpg
		if idx := strings.LastIndex(url, "/uploads/"); idx != -1 {
			used[url[idx+1:]] = struct{}{}
		}
	}

	files, err := os.ReadDir(h.cfg.UploadDir)
	if err != nil {
		http.Error(w, "failed to read upload dir: "+err.Error(), http.StatusInternalServerError)
		return
	}

	deleted := 0
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if _, ok := used[f.Name()]; !ok {
			_ = os.Remove(filepath.Join(h.cfg.UploadDir, f.Name()))
			deleted++
		}
	}

	helpers.JSON(w, http.StatusOK, map[string]int{"deleted": deleted})
}
