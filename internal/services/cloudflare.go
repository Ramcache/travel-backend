package services

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"github.com/Ramcache/travel-backend/internal/repository"
)

type CloudflareService struct {
	cf     *repository.CloudflareRepository
	zoneID string
	log    *zap.SugaredLogger
}

func NewCloudflareService(cf *repository.CloudflareRepository, zoneID string, log *zap.SugaredLogger) *CloudflareService {
	return &CloudflareService{
		cf:     cf,
		zoneID: zoneID,
		log:    log,
	}
}

type PurgeCacheInput struct {
	PurgeEverything bool
	Files           []string
	Tags            []string
	Hosts           []string
	Prefixes        []string
}

type PurgeCacheOutput struct {
	RequestID string
}

func (s *CloudflareService) PurgeCache(ctx context.Context, in PurgeCacheInput) (PurgeCacheOutput, error) {
	if strings.TrimSpace(s.zoneID) == "" {
		return PurgeCacheOutput{}, fmt.Errorf("cloudflare zone id is not configured")
	}

	// Нормализация: чистим пустые элементы
	clean := func(xs []string) []string {
		out := make([]string, 0, len(xs))
		for _, x := range xs {
			x = strings.TrimSpace(x)
			if x != "" {
				out = append(out, x)
			}
		}
		return out
	}

	req := repository.CloudflarePurgeCacheRequest{
		PurgeEverything: in.PurgeEverything,
		Files:           clean(in.Files),
		Tags:            clean(in.Tags),
		Hosts:           clean(in.Hosts),
		Prefixes:        clean(in.Prefixes),
	}

	res, err := s.cf.PurgeCache(ctx, s.zoneID, req)
	if err != nil {
		s.log.Warnw("cloudflare purge failed", "err", err)
		return PurgeCacheOutput{}, err
	}

	s.log.Infow("cloudflare purge ok", "request_id", res.ID, "purge_everything", in.PurgeEverything, "files_count", len(req.Files))
	return PurgeCacheOutput{RequestID: res.ID}, nil
}
