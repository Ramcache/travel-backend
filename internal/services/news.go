package services

import (
	"context"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/Ramcache/travel-backend/internal/models"
	"github.com/Ramcache/travel-backend/internal/repository"
)

type NewsService struct {
	repo *repository.NewsRepository
	log  *zap.SugaredLogger
}

func NewNewsService(r *repository.NewsRepository, log *zap.SugaredLogger) *NewsService {
	return &NewsService{repo: r, log: log}
}

var (
	allowedCat    = map[string]struct{}{"hadj": {}, "company": {}, "other": {}}
	allowedType   = map[string]struct{}{"photo": {}, "video": {}}
	allowedStatus = map[string]struct{}{"draft": {}, "published": {}, "archived": {}}
)

func (s *NewsService) ListPublic(ctx context.Context, p models.ListNewsParams) ([]models.News, int, error) {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.Limit <= 0 || p.Limit > 100 {
		p.Limit = 12
	}
	f := repository.NewsFilter{
		Category:  p.Category,
		MediaType: p.MediaType,
		Search:    p.Search,
		Status:    "published",
		Limit:     p.Limit,
		Offset:    (p.Page - 1) * p.Limit,
	}
	return s.repo.List(ctx, f)
}

func (s *NewsService) GetPublic(ctx context.Context, slugOrID string) (*models.News, error) {
	var n *models.News
	if id, ok := tryAtoi(slugOrID); ok {
		nn, err := s.repo.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		n = nn
	} else {
		nn, err := s.repo.GetBySlug(ctx, slugOrID)
		if err != nil {
			return nil, err
		}
		n = nn
	}

	if n == nil || n.Status != "published" {
		return nil, nil
	}

	go func(newsID int) {
		c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := s.repo.IncrementViews(c, newsID); err != nil {
			s.log.Warn("increment_views_failed", "id", newsID, "err", err)
		}
	}(n.ID)

	return n, nil
}

func (s *NewsService) AdminList(ctx context.Context, p models.ListNewsParams) ([]models.News, int, error) {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.Limit <= 0 || p.Limit > 100 {
		p.Limit = 20
	}
	f := repository.NewsFilter{
		Category:  p.Category,
		MediaType: p.MediaType,
		Search:    p.Search,
		Status:    p.Status,
		Limit:     p.Limit,
		Offset:    (p.Page - 1) * p.Limit,
	}
	return s.repo.List(ctx, f)
}

func (s *NewsService) Create(ctx context.Context, authorID *int, req models.CreateNewsRequest) (*models.News, error) {
	if req.Title == "" {
		return nil, errf("title is required")
	}
	if _, ok := allowedCat[req.Category]; !ok {
		return nil, errf("invalid category")
	}
	if _, ok := allowedType[req.MediaType]; !ok {
		return nil, errf("invalid media_type")
	}
	if req.Status == "" {
		req.Status = "published"
	}
	if _, ok := allowedStatus[req.Status]; !ok {
		return nil, errf("invalid status")
	}

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
	}

	var publishedAt time.Time
	if req.PublishedAt != "" {
		if t, err := time.Parse(time.RFC3339, req.PublishedAt); err == nil {
			publishedAt = t
		} else {
			return nil, errf("invalid published_at")
		}
	} else {
		publishedAt = time.Now()
	}

	n := &models.News{
		Slug:        slug,
		Title:       req.Title,
		Excerpt:     req.Excerpt,
		Content:     req.Content,
		Category:    req.Category,
		MediaType:   req.MediaType,
		PreviewURL:  req.PreviewURL,
		VideoURL:    req.VideoURL,
		AuthorID:    authorID,
		Status:      req.Status,
		PublishedAt: publishedAt,
	}
	if err := s.repo.Create(ctx, n); err != nil {
		return nil, err
	}
	return n, nil
}

func (s *NewsService) Update(ctx context.Context, id int, req models.UpdateNewsRequest) (*models.News, error) {
	n, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if n == nil {
		return nil, nil
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
	if v := req.Category; v != nil {
		if _, ok := allowedCat[*v]; !ok {
			return nil, errf("invalid category")
		}
		n.Category = *v
	}
	if v := req.MediaType; v != nil {
		if _, ok := allowedType[*v]; !ok {
			return nil, errf("invalid media_type")
		}
		n.MediaType = *v
	}
	if v := req.PreviewURL; v != nil {
		n.PreviewURL = *v
	}
	if v := req.VideoURL; v != nil {
		n.VideoURL = v
	}
	if v := req.Status; v != nil {
		if _, ok := allowedStatus[*v]; !ok {
			return nil, errf("invalid status")
		}
		n.Status = *v
	}
	if v := req.PublishedAt; v != nil && *v != "" {
		t, err := time.Parse(time.RFC3339, *v)
		if err != nil {
			return nil, errf("invalid published_at")
		}
		n.PublishedAt = t
	}

	if err := s.repo.Update(ctx, n); err != nil {
		return nil, err
	}
	return n, nil
}

func (s *NewsService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

// helpers
func errf(msg string) error { return &simpleErr{msg: msg} }

type simpleErr struct{ msg string }

func (e *simpleErr) Error() string { return e.msg }

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
	// простая транслитерация ru -> lat
	repl := map[rune]string{
		'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d", 'е': "e", 'ё': "e", 'ж': "zh", 'з': "z", 'и': "i", 'й': "y", 'к': "k", 'л': "l", 'м': "m", 'н': "n", 'о': "o", 'п': "p", 'р': "r", 'с': "s", 'т': "t", 'у': "u", 'ф': "f", 'х': "h", 'ц': "c", 'ч': "ch", 'ш': "sh", 'щ': "sch", 'ъ': "", 'ы': "y", 'ь': "", 'э': "e", 'ю': "yu", 'я': "ya",
		'А': "a", 'Б': "b", 'В': "v", 'Г': "g", 'Д': "d", 'Е': "e", 'Ё': "e", 'Ж': "zh", 'З': "z", 'И': "i", 'Й': "y", 'К': "k", 'Л': "l", 'М': "m", 'Н': "n", 'О': "o", 'П': "p", 'Р': "r", 'С': "s", 'Т': "t", 'У': "u", 'Ф': "f", 'Х': "h", 'Ц': "c", 'Ч': "ch", 'Ш': "sh", 'Щ': "sch", 'Ъ': "", 'Ы': "y", 'Ь': "", 'Э': "e", 'Ю': "yu", 'Я': "ya",
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
	// убрать все, кроме латиницы, цифр и '-'
	cleaned := make([]rune, 0, len(out))
	for _, r := range out {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			cleaned = append(cleaned, r)
		}
	}
	// схлопнуть повторяющиеся '-'
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
	return strings.Trim(resToString(res), "-")
}

func (s *NewsService) GetRecent(ctx context.Context, limit int) ([]models.News, error) {
	return s.repo.GetRecent(ctx, limit)
}

func (s *NewsService) GetPopular(ctx context.Context, limit int) ([]models.News, error) {
	return s.repo.GetPopular(ctx, limit)
}

func (s *NewsService) IncrementViews(ctx context.Context, id int) error {
	return s.repo.IncrementViews(ctx, id)
}

func resToString(rs []rune) string { return string(rs) }
