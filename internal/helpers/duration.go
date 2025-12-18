package helpers

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	reHM    = regexp.MustCompile(`(?i)(\d+)\s*(?:ч|час(?:а|ов)?)\s*(?:$|\s)`)
	reMM    = regexp.MustCompile(`(?i)(\d+)\s*(?:м|мин(?:\.|ута|уты|ут)?|минут)\s*(?:$|\s)`)
	reDD    = regexp.MustCompile(`(?i)(\d+)\s*(?:д|дн(?:ей|я)?|день|дня|дней)\s*(?:$|\s)`)
	reClock = regexp.MustCompile(`(?i)^\s*(\d{1,2})\s*:\s*(\d{2})\s*$`)
)

// ParseDurationText понимает: "6 часов", "2 часа", "45 минут", "1 день 3 часа", "12:30", "2ч30м".
func ParseDurationText(s string) time.Duration {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}

	if m := reClock.FindStringSubmatch(s); len(m) == 3 {
		h, err1 := strconv.Atoi(m[1])
		mi, err2 := strconv.Atoi(m[2])
		if err1 != nil || err2 != nil || mi < 0 || mi > 59 {
			return 0
		}
		// Если вам нужно разрешить "36:00" как 36 часов — НЕ ограничивайте h до 23.
		if h < 0 {
			return 0
		}
		return time.Duration(h)*time.Hour + time.Duration(mi)*time.Minute
	}

	var d time.Duration
	for _, m := range reDD.FindAllStringSubmatch(s+" ", -1) { // + " " чтобы сработало (?:$|\s)
		if n, err := strconv.Atoi(m[1]); err == nil {
			d += time.Duration(n) * 24 * time.Hour
		}
	}
	for _, m := range reHM.FindAllStringSubmatch(s+" ", -1) {
		if n, err := strconv.Atoi(m[1]); err == nil {
			d += time.Duration(n) * time.Hour
		}
	}
	for _, m := range reMM.FindAllStringSubmatch(s+" ", -1) {
		if n, err := strconv.Atoi(m[1]); err == nil {
			d += time.Duration(n) * time.Minute
		}
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
// FormatDuration -> "2 дня 3 часа 10 минут" / "14 часов" / "45 минут"
func FormatDuration(d time.Duration) string {
	if d <= 0 {
		return "0 минут"
	}

	days := int(d / (24 * time.Hour))
	d -= time.Duration(days) * 24 * time.Hour

	hours := int(d / time.Hour)
	d -= time.Duration(hours) * time.Hour

	mins := int(d / time.Minute)

	parts := make([]string, 0, 3)
	if days > 0 {
		parts = append(parts, plural(days, "день", "дня", "дней"))
	}
	if hours > 0 {
		parts = append(parts, plural(hours, "час", "часа", "часов"))
	}
	if mins > 0 || len(parts) == 0 {
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
