package utils

import "time"

func ToDate(date string) time.Time {
	layout := "2006-01-02 15:04:05"
	t, err := time.Parse(layout, date)
	if err != nil {
		return time.Time{}
	}
	return t
}

func DiffSeconds(startStr, endStr string) (int64, error) {
	layout := time.RFC3339

	start, err := time.Parse(layout, startStr)
	if err != nil {
		return 0, err
	}

	end, err := time.Parse(layout, endStr)
	if err != nil {
		return 0, err
	}

	return int64(end.Sub(start).Seconds()), nil
}
