package helpers

import (
	"regexp"
	"strings"
	"time"
)

var (
	reHM    = regexp.MustCompile(`(?i)(\d+)\s*(?:ч|час|часа|часов)`)
	reMM    = regexp.MustCompile(`(?i)(\d+)\s*(?:м|мин|мин\.?|минута|минуты|минут)`)
	reDD    = regexp.MustCompile(`(?i)(\d+)\s*(?:д|дн|день|дня|дней)`)
	reClock = regexp.MustCompile(`^(?i)(\d{1,2}):(\d{2})$`)
)

// ParseDurationText понимает: "6 часов", "2 часа", "45 минут", "1 день 3 часа", "12:30".
func ParseDurationText(s string) time.Duration {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	// HH:MM
	if m := reClock.FindStringSubmatch(s); len(m) == 3 {
		h := atoi(m[1])
		mi := atoi(m[2])
		return time.Duration(h)*time.Hour + time.Duration(mi)*time.Minute
	}

	var d time.Duration
	if m := reDD.FindStringSubmatch(s); len(m) == 2 {
		d += time.Duration(atoi(m[1])) * 24 * time.Hour
	}
	if m := reHM.FindStringSubmatch(s); len(m) == 2 {
		d += time.Duration(atoi(m[1])) * time.Hour
	}
	if m := reMM.FindStringSubmatch(s); len(m) == 2 {
		d += time.Duration(atoi(m[1])) * time.Minute
	}
	return d
}

func atoi(s string) int {
	n := 0
	for _, ch := range s {
		n = n*10 + int(ch-'0')
	}
	return n
}

// FormatDuration -> "2 дня 3 часа" / "14 часов" / "45 минут"
func FormatDuration(d time.Duration) string {
	if d <= 0 {
		return "0 минут"
	}
	days := int(d / (24 * time.Hour))
	d %= 24 * time.Hour
	hours := int(d / time.Hour)
	d %= time.Hour
	mins := int(d / time.Minute)

	parts := make([]string, 0, 3)
	if days > 0 {
		parts = append(parts, plural(days, "день", "дня", "дней"))
	}
	if hours > 0 {
		parts = append(parts, plural(hours, "час", "часа", "часов"))
	}
	if len(parts) == 0 && mins > 0 {
		parts = append(parts, plural(mins, "минута", "минуты", "минут"))
	}
	return strings.Join(parts, " ")
}

func plural(n int, one, few, many string) string {
	word := many
	if n%10 == 1 && n%100 != 11 {
		word = one
	} else if n%10 >= 2 && n%10 <= 4 && (n%100 < 12 || n%100 > 14) {
		word = few
	}
	return itoa(n) + " " + word
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	buf := make([]byte, 0, 12)
	for n > 0 {
		buf = append([]byte{'0' + byte(n%10)}, buf...)
		n /= 10
	}
	if neg {
		buf = append([]byte{'-'}, buf...)
	}
	return string(buf)
}
