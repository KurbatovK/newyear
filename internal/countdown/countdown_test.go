package countdown

import (
	"testing"
	"time"
)

func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func TestDaysToNewYear(t *testing.T) {
	tests := []struct {
		name string
		in   time.Time
		want int
	}{
		{"start of non-leap year", date(2025, time.January, 1), 365},
		{"start of leap year", date(2024, time.January, 1), 366},
		{"end of year", date(2025, time.December, 31), 1},
		{"end of leap year", date(2024, time.December, 31), 1},
		{"before Feb 29 in leap year", date(2024, time.February, 28), 308},
		{"Feb 29 in leap year", date(2024, time.February, 29), 307},
		{"after Feb 29 in leap year", date(2024, time.March, 1), 306},
		{"Feb 28 in non-leap year", date(2025, time.February, 28), 307},
		{"Mar 1 in non-leap year", date(2025, time.March, 1), 306},
		{"century non-leap year", date(1900, time.March, 1), 306},
		{"400-year leap year", date(2000, time.February, 29), 307},
		{"ordinary date", date(2026, time.September, 30), 93},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DaysToNewYear(tt.in); got != tt.want {
				t.Errorf("DaysToNewYear(%s) = %d, want %d", tt.in.Format(time.DateOnly), got, tt.want)
			}
		})
	}
}

func TestDaysToNewYearIgnoresTimeOfDay(t *testing.T) {
	early := time.Date(2025, time.December, 31, 0, 0, 0, 0, time.UTC)
	late := time.Date(2025, time.December, 31, 23, 59, 59, 999999999, time.UTC)

	if a, b := DaysToNewYear(early), DaysToNewYear(late); a != b {
		t.Errorf("DaysToNewYear depends on time of day: %d != %d", a, b)
	}
}

func TestDaysToNewYearUsesLocalDate(t *testing.T) {
	// 20:00 UTC 31 декабря — в Иркутске уже 1 января
	loc := time.FixedZone("UTC+8", 8*60*60)
	in := time.Date(2025, time.December, 31, 20, 0, 0, 0, time.UTC).In(loc)

	if got, want := DaysToNewYear(in), 365; got != want {
		t.Errorf("DaysToNewYear(%s) = %d, want %d", in, got, want)
	}
}

func TestDaysToNewYearAcrossDSTChange(t *testing.T) {
	loc, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Skipf("time zone database is not available: %v", err)
	}
	// в конце октября в Европе переводят часы
	in := time.Date(2025, time.October, 1, 0, 30, 0, 0, loc)

	if got, want := DaysToNewYear(in), 92; got != want {
		t.Errorf("DaysToNewYear(%s) = %d, want %d", in, got, want)
	}
}

func BenchmarkDaysToNewYear(b *testing.B) {
	in := date(2024, time.February, 29)
	for b.Loop() {
		DaysToNewYear(in)
	}
}
