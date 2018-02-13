package goest_worker

import "time"

type Schedule struct {
	Weekday    time.Weekday
	Hour       int
	Minute     int
	Second     int
	NanoSecond int
}

func (s *Schedule) Next() (time.Duration) {
	now := s.now()
	next := now
	if (now.Weekday() != s.Weekday) && s.Weekday != 0 {
		next = next.AddDate(0, 0, (7-int((now.Weekday())))+int(s.Weekday))
	}
	next = time.Date(next.Year(), next.Month(), next.Day(), s.Hour, s.Minute, s.Second, s.NanoSecond, now.Location())
	diff := next.Sub(now)
	if diff < 0 {
		diff += time.Hour * 24
	}
	return diff
}

// mock time for tests
func (s *Schedule) now () (time.Time) {
	return time.Now()
}

