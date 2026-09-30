// Package countdown считает количество дней до Нового года.
package countdown

import "time"

const hoursPerDay = 24

// DaysToNewYear возвращает количество дней от даты t до 1 января следующего года.
// Время суток не учитывается: для 31 декабря результат 1, для 1 января — 365 (366).
func DaysToNewYear(t time.Time) int {
	year, month, day := t.Date()

	// в UTC нет перевода часов, поэтому разница всегда кратна 24 часам
	from := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	to := NextNewYear(t)

	return int(to.Sub(from).Hours()) / hoursPerDay
}

// NextNewYear возвращает 1 января года, следующего за t.
func NextNewYear(t time.Time) time.Time {
	return time.Date(t.Year()+1, time.January, 1, 0, 0, 0, 0, time.UTC)
}
