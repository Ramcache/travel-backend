package helpers

func IfEmpty(s *string, def string) string {
	if s == nil || *s == "" {
		return def
	}
	return *s
}
