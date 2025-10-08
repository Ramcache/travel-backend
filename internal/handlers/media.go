package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/Ramcache/travel-backend/internal/config"
	"github.com/Ramcache/travel-backend/internal/helpers"
)

type MediaHandler struct {
	cfg *config.Config
	db  *pgxpool.Pool
	log *zap.SugaredLogger
}

func NewMediaHandler(cfg *config.Config, db *pgxpool.Pool, log *zap.SugaredLogger) *MediaHandler {
	return &MediaHandler{cfg: cfg, db: db, log: log}
}

// Upload
// @Summary Upload photos
// @Description Загрузка одного или нескольких фото (админка)
// @Tags Admin — Media
// @Security Bearer
// @Accept multipart/form-data
// @Produce json
// @Param files formData file true "Фото (можно несколько)"
// @Success 200 {array} map[string]string
// @Failure 400 {object} helpers.ErrorData
// @Failure 500 {object} helpers.ErrorData
// @Router /admin/upload [post]
func (h *MediaHandler) Upload(w http.ResponseWriter, r *http.Request) {
	maxSize := int64(h.cfg.MaxUploadMB) << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)

	if err := r.ParseMultipartForm(maxSize); err != nil {
		h.log.Errorw("Ошибка парсинга multipart формы", "err", err)
		helpers.Error(w, http.StatusBadRequest, "Некорректная форма загрузки файлов")
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		helpers.Error(w, http.StatusBadRequest, "Файлы не прикреплены")
		return
	}

	if err := os.MkdirAll(h.cfg.UploadDir, 0755); err != nil {
		h.log.Errorw("Ошибка создания директории загрузки", "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Ошибка при сохранении файлов")
		return
	}

	var urls []map[string]string
	for _, fh := range files {
		filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(fh.Filename))
		dst := filepath.Join(h.cfg.UploadDir, filename)

		src, err := fh.Open()
		if err != nil {
			h.log.Errorw("Ошибка открытия файла", "err", err, "filename", fh.Filename)
			helpers.Error(w, http.StatusInternalServerError, "Ошибка чтения файла")
			return
		}
		defer src.Close()

		out, err := os.Create(dst)
		if err != nil {
			h.log.Errorw("Ошибка создания файла", "err", err, "path", dst)
			helpers.Error(w, http.StatusInternalServerError, "Ошибка сохранения файла")
			return
		}
		defer out.Close()

		if _, err := io.Copy(out, src); err != nil {
			h.log.Errorw("Ошибка копирования файла", "err", err, "filename", fh.Filename)
			helpers.Error(w, http.StatusInternalServerError, "Ошибка при копировании файла")
			return
		}

		url := fmt.Sprintf("%s/uploads/%s", h.cfg.AppBaseURL, filename)
		urls = append(urls, map[string]string{"url": url})
		h.log.Infow("Файл успешно загружен", "file", filename, "url", url)
	}

	helpers.JSON(w, http.StatusOK, urls)
}

// ListUploads
// @Summary Get all uploaded files
// @Description Получить список всех загруженных файлов (админка)
// @Tags Admin — Media
// @Security Bearer
// @Produce json
// @Success 200 {array} map[string]string
// @Failure 500 {object} helpers.ErrorData
// @Router /admin/uploads [get]
func (h *MediaHandler) ListUploads(w http.ResponseWriter, _ *http.Request) {
	files, err := os.ReadDir(h.cfg.UploadDir)
	if err != nil {
		h.log.Errorw("Ошибка чтения директории загрузок", "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось прочитать список файлов")
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
// @Description Удалить файл по имени (админка)
// @Tags Admin — Media
// @Security Bearer
// @Param file query string true "Имя файла"
// @Success 200 {object} map[string]string
// @Failure 400 {object} helpers.ErrorData
// @Failure 404 {object} helpers.ErrorData
// @Failure 500 {object} helpers.ErrorData
// @Router /admin/upload [delete]
func (h *MediaHandler) DeleteUpload(w http.ResponseWriter, r *http.Request) {
	fileName := r.URL.Query().Get("file")
	if fileName == "" {
		helpers.Error(w, http.StatusBadRequest, "Отсутствует параметр ?file=")
		return
	}

	if strings.Contains(fileName, "..") || strings.ContainsAny(fileName, "/\\") {
		helpers.Error(w, http.StatusBadRequest, "Некорректное имя файла")
		return
	}

	filePath := filepath.Join(h.cfg.UploadDir, fileName)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		helpers.Error(w, http.StatusNotFound, "Файл не найден")
		return
	}

	if err := os.Remove(filePath); err != nil {
		h.log.Errorw("Ошибка удаления файла", "err", err, "file", fileName)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось удалить файл")
		return
	}

	h.log.Infow("Файл удалён", "file", fileName)
	helpers.JSON(w, http.StatusOK, map[string]string{
		"status": "deleted",
		"file":   fileName,
	})
}

// CleanupUnused
// @Summary Cleanup unused media files
// @Description Удалить неиспользуемые фото из /uploads (админка)
// @Tags Admin — Media
// @Security Bearer
// @Success 200 {object} map[string]int
// @Failure 500 {object} helpers.ErrorData
// @Router /admin/media/cleanup [post]
func (h *MediaHandler) CleanupUnused(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	used, err := h.collectUsedFiles(ctx)
	if err != nil {
		h.log.Errorw("Ошибка при сборе использованных ссылок", "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось собрать ссылки из БД")
		return
	}

	files, err := os.ReadDir(h.cfg.UploadDir)
	if err != nil {
		h.log.Errorw("Ошибка чтения директории загрузок", "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось прочитать директорию загрузок")
		return
	}

	deleted := 0
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if _, ok := used[f.Name()]; !ok {
			if err := os.Remove(filepath.Join(h.cfg.UploadDir, f.Name())); err == nil {
				deleted++
				h.log.Infow("Удалён неиспользуемый файл", "file", f.Name())
			}
		}
	}

	helpers.JSON(w, http.StatusOK, map[string]int{"deleted": deleted})
}

// collectUsedFiles — утилита для поиска всех файлов, используемых в БД
func (h *MediaHandler) collectUsedFiles(ctx context.Context) (map[string]struct{}, error) {
	rows, err := h.db.Query(ctx, `
		SELECT unnest(urls) FROM trips
		UNION
		SELECT unnest(urls) FROM hotels
		UNION
		SELECT unnest(urls) FROM news
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	used := make(map[string]struct{})
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			continue
		}
		if idx := strings.LastIndex(url, "/"); idx != -1 {
			used[url[idx+1:]] = struct{}{}
		}
	}
	return used, nil
}
