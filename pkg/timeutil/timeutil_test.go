package timeutil

import (
	"testing"
	"time"
)

func TestWeekBounds(t *testing.T) {
	tests := []struct {
		name       string
		input      time.Time
		wantMonday time.Time
		wantSunday time.Time
	}{
		{
			name:       "wednesday mid-week",
			input:      time.Date(2026, 3, 11, 15, 30, 0, 0, time.UTC),
			wantMonday: time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC),
			wantSunday: time.Date(2026, 3, 15, 23, 59, 59, 0, time.UTC),
		},
		{
			name:       "monday itself",
			input:      time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC),
			wantMonday: time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC),
			wantSunday: time.Date(2026, 3, 15, 23, 59, 59, 0, time.UTC),
		},
		{
			name:       "sunday itself",
			input:      time.Date(2026, 3, 15, 20, 0, 0, 0, time.UTC),
			wantMonday: time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC),
			wantSunday: time.Date(2026, 3, 15, 23, 59, 59, 0, time.UTC),
		},
		{
			name:       "year boundary crossing",
			input:      time.Date(2025, 12, 31, 12, 0, 0, 0, time.UTC),
			wantMonday: time.Date(2025, 12, 29, 0, 0, 0, 0, time.UTC),
			wantSunday: time.Date(2026, 1, 4, 23, 59, 59, 0, time.UTC),
		},
		{
			name:       "january first on thursday",
			input:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			wantMonday: time.Date(2025, 12, 29, 0, 0, 0, 0, time.UTC),
			wantSunday: time.Date(2026, 1, 4, 23, 59, 59, 0, time.UTC),
		},
		{
			name:       "saturday",
			input:      time.Date(2026, 3, 14, 8, 0, 0, 0, time.UTC),
			wantMonday: time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC),
			wantSunday: time.Date(2026, 3, 15, 23, 59, 59, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monday, sunday := WeekBounds(tt.input)
			if !monday.Equal(tt.wantMonday) {
				t.Errorf("monday = %v, want %v", monday, tt.wantMonday)
			}
			if !sunday.Equal(tt.wantSunday) {
				t.Errorf("sunday = %v, want %v", sunday, tt.wantSunday)
			}
		})
	}
}

func TestWeekNumber(t *testing.T) {
	tests := []struct {
		name string
		date time.Time
		want int
	}{
		{
			name: "week 11 of 2026",
			date: time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC),
			want: 11,
		},
		{
			name: "week 1 crossing year",
			date: time.Date(2025, 12, 29, 0, 0, 0, 0, time.UTC),
			want: 1, // ISO: this Monday belongs to week 1 of 2026
		},
		{
			name: "first day of year",
			date: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WeekNumber(tt.date)
			if got != tt.want {
				t.Errorf("WeekNumber = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestDaysOfWeek(t *testing.T) {
	wednesday := time.Date(2026, 3, 11, 12, 0, 0, 0, time.UTC)
	days := DaysOfWeek(wednesday)

	// First day should be Monday.
	if days[0].Weekday() != time.Monday {
		t.Errorf("first day = %v, want Monday", days[0].Weekday())
	}

	// Last day should be Sunday.
	if days[6].Weekday() != time.Sunday {
		t.Errorf("last day = %v, want Sunday", days[6].Weekday())
	}

	// All 7 days should be consecutive.
	for i := 1; i < 7; i++ {
		diff := days[i].Sub(days[i-1])
		if diff != 24*time.Hour {
			t.Errorf("gap between day %d and %d = %v, want 24h", i-1, i, diff)
		}
	}
}
