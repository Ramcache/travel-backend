package services

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/Ramcache/travel-backend/internal/helpers"
	"github.com/Ramcache/travel-backend/internal/models"
	"github.com/Ramcache/travel-backend/internal/repository"
)

var (
	ErrInvalidInput  = errors.New("invalid input")
	ErrDuplicateSlug = errors.New("slug already exists")
)

type NewsService struct {
	repo    *repository.NewsRepository
	catRepo *repository.NewsCategoryRepository
	log     *zap.SugaredLogger
}

func NewNewsService(r *repository.NewsRepository, c *repository.NewsCategoryRepository, log *zap.SugaredLogger) *NewsService {
	return &NewsService{repo: r, catRepo: c, log: log}
}

var (
	allowedType   = map[string]struct{}{"photo": {}, "video": {}}
	allowedStatus = map[string]struct{}{"draft": {}, "published": {}, "archived": {}}
)

// mapNotFound переводит ошибку репозитория в сервисную
func mapNotFound(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, repository.ErrNotFound) {
		return ErrNotFound
	}
	return err
}

// ListPublic — список новостей для публичной выдачи
func (s *NewsService) ListPublic(ctx context.Context, p models.ListNewsParams) ([]models.News, int, error) {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.Limit <= 0 || p.Limit > 100 {
		p.Limit = 12
	}
	f := repository.NewsFilter{
		CategoryID: p.CategoryID,
		MediaType:  p.MediaType,
		Search:     p.Search,
		Status:     "published",
		Limit:      p.Limit,
		Offset:     (p.Page - 1) * p.Limit,
	}
	return s.repo.List(ctx, f)
}

// GetPublic — получить новость по slug или ID
func (s *NewsService) GetPublic(ctx context.Context, slugOrID string) (*models.News, error) {
	var n *models.News
	var err error

	if id, ok := tryAtoi(slugOrID); ok {
		n, err = s.repo.GetByID(ctx, id)
	} else {
		n, err = s.repo.GetBySlug(ctx, slugOrID)
	}
	if err != nil {
		return nil, mapNotFound(err)
	}
	if n == nil || n.Status != "published" {
		return nil, ErrNotFound
	}

	// Инкремент просмотров асинхронно
	go func(newsID int) {
		c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := s.repo.IncrementViews(c, newsID); err != nil {
			s.log.Warnw("increment_views_failed", "id", newsID, "err", err)
		}
	}(n.ID)

	return n, nil
}

// AdminList — список новостей для админки
func (s *NewsService) AdminList(ctx context.Context, p models.ListNewsParams) ([]models.News, int, error) {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.Limit <= 0 || p.Limit > 100 {
		p.Limit = 20
	}
	f := repository.NewsFilter{
		CategoryID: p.CategoryID,
		MediaType:  p.MediaType,
		Search:     p.Search,
		Status:     p.Status,
		Limit:      p.Limit,
		Offset:     (p.Page - 1) * p.Limit,
	}
	return s.repo.List(ctx, f)
}

// Create — создать новость
func (s *NewsService) Create(ctx context.Context, authorID *int, req models.CreateNewsRequest) (*models.News, error) {
	if req.Title == "" {
		return nil, helpers.ErrInvalidInput("Заголовок обязателен")
	}
	if ok, _ := s.catRepo.Exists(ctx, req.CategoryID); !ok {
		return nil, helpers.ErrInvalidInput("Некорректная категория")
	}
	if _, ok := allowedType[req.MediaType]; !ok {
		return nil, helpers.ErrInvalidInput("Некорректный тип медиа")
	}
	if req.Status == "" {
		req.Status = "published"
	}
	if _, ok := allowedStatus[req.Status]; !ok {
		return nil, helpers.ErrInvalidInput("Некорректный статус")
	}

	// slug уникальный
	baseSlug := slugify(req.Title)
	slug := baseSlug
	for i := 2; ; i++ {
		exists, err := s.repo.ExistsSlug(ctx, slug)
		if err != nil {
			return nil, err
		}
		if !exists {
			break
		}
		slug = baseSlug + "-" + itoa(i)
		if i > 100 {
			return nil, ErrDuplicateSlug
		}
	}

	var publishedAt time.Time
	if req.PublishedAt != "" {
		t, err := time.Parse(time.RFC3339, req.PublishedAt)
		if err != nil {
			return nil, helpers.ErrInvalidInput("Некорректная дата публикации")
		}
		publishedAt = t
	} else {
		publishedAt = time.Now()
	}

	n := &models.News{
		Slug:        slug,
		Title:       req.Title,
		Excerpt:     req.Excerpt,
		Content:     req.Content,
		CategoryID:  &req.CategoryID,
		MediaType:   req.MediaType,
		URLs:        req.URLs, // 👈 массив ссылок
		VideoURL:    req.VideoURL,
		AuthorID:    authorID,
		Status:      req.Status,
		PublishedAt: publishedAt,
	}
	if err := s.repo.Create(ctx, n); err != nil {
		s.log.Errorw("news_create_failed", "title", req.Title, "err", err)
		return nil, err
	}

	s.log.Infow("news_created", "id", n.ID, "title", n.Title)
	return n, nil
}

// Update — обновить новость
func (s *NewsService) Update(ctx context.Context, id int, req models.UpdateNewsRequest) (*models.News, error) {
	n, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, mapNotFound(err)
	}
	if n == nil {
		return nil, ErrNotFound
	}

	if v := req.Slug; v != nil && *v != "" {
		n.Slug = *v
	}
	if v := req.Title; v != nil {
		n.Title = *v
	}
	if v := req.Excerpt; v != nil {
		n.Excerpt = *v
	}
	if v := req.Content; v != nil {
		n.Content = *v
	}
	if v := req.CategoryID; v != nil {
		if ok, _ := s.catRepo.Exists(ctx, *v); !ok {
			return nil, helpers.ErrInvalidInput("Некорректная категория")
		}
		n.CategoryID = v
	}
	if v := req.MediaType; v != nil {
		if _, ok := allowedType[*v]; !ok {
			return nil, helpers.ErrInvalidInput("Некорректный тип медиа")
		}
		n.MediaType = *v
	}
	if v := req.URLs; v != nil { // 👈 массив ссылок
		n.URLs = *v
	}
	if v := req.VideoURL; v != nil {
		n.VideoURL = v
	}
	if v := req.Status; v != nil {
		if _, ok := allowedStatus[*v]; !ok {
			return nil, helpers.ErrInvalidInput("Некорректный статус")
		}
		n.Status = *v
	}
	if v := req.PublishedAt; v != nil && *v != "" {
		t, err := time.Parse(time.RFC3339, *v)
		if err != nil {
			return nil, helpers.ErrInvalidInput("Некорректная дата публикации")
		}
		n.PublishedAt = t
	}

	if err := s.repo.Update(ctx, n); err != nil {
		s.log.Errorw("news_update_failed", "id", id, "err", err)
		return nil, mapNotFound(err)
	}

	s.log.Infow("news_updated", "id", id)
	return n, nil
}

// Delete — удалить новость
func (s *NewsService) Delete(ctx context.Context, id int) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		s.log.Errorw("news_delete_failed", "id", id, "err", err)
		return mapNotFound(err)
	}
	s.log.Infow("news_deleted", "id", id)
	return nil
}

// GetRecent — последние новости
func (s *NewsService) GetRecent(ctx context.Context, limit int) ([]models.News, error) {
	return s.repo.GetRecent(ctx, limit)
}

// GetPopular — популярные новости
func (s *NewsService) GetPopular(ctx context.Context, limit int) ([]models.News, error) {
	return s.repo.GetPopular(ctx, limit)
}

// helpers
func tryAtoi(s string) (int, bool) {
	var n int
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, false
		}
	}
	for i := 0; i < len(s); i++ {
		n = n*10 + int(s[i]-'0')
	}
	return n, true
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

var spacesRe = regexp.MustCompile(`\s+`)

func slugify(s string) string {
	repl := map[rune]string{
		'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d", 'е': "e", 'ё': "e", 'ж': "zh", 'з': "z", 'и': "i", 'й': "y",
		'к': "k", 'л': "l", 'м': "m", 'н': "n", 'о': "o", 'п': "p", 'р': "r", 'с': "s", 'т': "t", 'у': "u", 'ф': "f",
		'х': "h", 'ц': "c", 'ч': "ch", 'ш': "sh", 'щ': "sch", 'ъ': "", 'ы': "y", 'ь': "", 'э': "e", 'ю': "yu", 'я': "ya",
	}
	var b strings.Builder
	for _, r := range s {
		if v, ok := repl[r]; ok {
			b.WriteString(v)
		} else {
			b.WriteRune(r)
		}
	}
	out := strings.ToLower(b.String())
	out = spacesRe.ReplaceAllString(out, "-")

	// оставить только латиницу, цифры и "-"
	cleaned := make([]rune, 0, len(out))
	for _, r := range out {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			cleaned = append(cleaned, r)
		}
	}

	// схлопнуть "--"
	res := make([]rune, 0, len(cleaned))
	var prevDash bool
	for _, r := range cleaned {
		if r == '-' {
			if prevDash {
				continue
			}
			prevDash = true
		} else {
			prevDash = false
		}
		res = append(res, r)
	}
	return strings.Trim(string(res), "-")
}

// PublicList — обёртка для обратной совместимости (используется в TripPageService)
func (s *NewsService) PublicList(ctx context.Context, limit, offset int) ([]models.News, int, error) {
	items, total, err := s.repo.List(ctx, repository.NewsFilter{
		Status: "published",
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		s.log.Errorw("news_public_list_failed", "err", err)
		return nil, 0, err
	}
	return items, total, nil
}
