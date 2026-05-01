package utils

import "time"

func ToDate(date time.Time) time.Time {
	layout := "2006-01-02 15:04:05"
	t, _ := time.Parse(layout, date.Format(layout))
	return t
}

func DiffSeconds(start time.Time, end time.Time) int64 {
	return int64(end.Sub(start).Seconds())
}

func ToDateTime(date string) (time.Time, error) {
	return time.Parse(time.RFC3339, date)
}
