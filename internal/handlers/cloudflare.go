package handlers

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"github.com/Ramcache/travel-backend/internal/services"
)

type CloudflareHandler struct {
	svc *services.CloudflareService
	log *zap.SugaredLogger
}

func NewCloudflareHandler(svc *services.CloudflareService, log *zap.SugaredLogger) *CloudflareHandler {
	return &CloudflareHandler{svc: svc, log: log}
}

type purgeCacheRequest struct {
	PurgeEverything bool     `json:"purge_everything"`
	Files           []string `json:"files"`
	Tags            []string `json:"tags"`
	Hosts           []string `json:"hosts"`
	Prefixes        []string `json:"prefixes"`
}

type purgeCacheResponse struct {
	RequestID string `json:"request_id"`
}

// PurgeCache godoc
// @Summary      Purge Cloudflare cache
// @Description  Purge Cloudflare cache for a zone. Supports full purge or selective purge by URLs, tags, hosts, or prefixes.
// @Tags         Cloudflare
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body PurgeCacheRequest true "Purge cache request"
// @Success      200 {object} PurgeCacheResponse
// @Failure      400 {object} map[string]string "Bad request"
// @Failure      401 {object} map[string]string "Unauthorized"
// @Failure      403 {object} map[string]string "Forbidden"
// @Router       /admin/cloudflare/purge-cache [post]
func (h *CloudflareHandler) PurgeCache(w http.ResponseWriter, r *http.Request) {
	var req purgeCacheRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	out, err := h.svc.PurgeCache(r.Context(), services.PurgeCacheInput{
		PurgeEverything: req.PurgeEverything,
		Files:           req.Files,
		Tags:            req.Tags,
		Hosts:           req.Hosts,
		Prefixes:        req.Prefixes,
	})
	if err != nil {
		// Можно маппить конкретнее, но для старта достаточно 400
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(purgeCacheResponse{RequestID: out.RequestID})
}
