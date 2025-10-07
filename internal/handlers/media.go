package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/Ramcache/travel-backend/internal/config"
	"github.com/Ramcache/travel-backend/internal/helpers"
)

type MediaHandler struct {
	cfg *config.Config
}

func NewMediaHandler(cfg *config.Config) *MediaHandler {
	return &MediaHandler{cfg: cfg}
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
