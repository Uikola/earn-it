package timeutil

import "time"

// WeekStart возвращает начало недели (понедельник 00:00) для заданного времени в указанной локации,
// конвертированное в UTC.
func WeekStart(t time.Time, loc *time.Location) time.Time {
	localTime := t.In(loc)
	weekday := localTime.Weekday()
	offset := (int(weekday) - int(time.Monday) + 7) % 7
	weekStartLocal := localTime.AddDate(0, 0, -offset).Truncate(24 * time.Hour)
	return weekStartLocal.UTC()
}
