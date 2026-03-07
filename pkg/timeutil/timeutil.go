// Package timeutil provides time calculation helpers for nullcal.
//
// All week calculations use ISO 8601 convention: Monday is the first day
// of the week.
package timeutil

import "time"

// WeekBounds returns the Monday 00:00:00 and Sunday 23:59:59 for the
// week containing t. Uses ISO 8601 convention (Monday = first day).
func WeekBounds(t time.Time) (monday, sunday time.Time) {
	t = stripTime(t)

	wd := t.Weekday()
	if wd == time.Sunday {
		wd = 7
	}
	offset := int(wd) - int(time.Monday)

	monday = t.AddDate(0, 0, -offset)
	sunday = monday.AddDate(0, 0, 6)
	sunday = time.Date(
		sunday.Year(), sunday.Month(), sunday.Day(),
		23, 59, 59, 0, t.Location(),
	)

	return monday, sunday
}

// WeekNumber returns the ISO 8601 week number for the given time.
func WeekNumber(t time.Time) int {
	_, week := t.ISOWeek()
	return week
}

// DaysOfWeek returns an array of 7 dates starting from Monday of the
// week containing t.
func DaysOfWeek(t time.Time) [7]time.Time {
	monday, _ := WeekBounds(t)
	var days [7]time.Time
	for i := 0; i < 7; i++ {
		days[i] = monday.AddDate(0, 0, i)
	}
	return days
}

// stripTime returns t with the time component zeroed out, preserving the location.
func stripTime(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
