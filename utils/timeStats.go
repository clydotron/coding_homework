package utils

import (
	"encoding/json"
	"sync"
	"time"
)

type TimeStats struct {
	n  int
	t  time.Duration
	mx sync.Mutex
}

type TimeStatsReport struct {
	Total   int   `json:"total"`
	Average int64 `json:"average"`
}

// add a value to the total, increment the counter (n)
func (s *TimeStats) Add(value time.Duration) {
	s.mx.Lock()
	defer s.mx.Unlock()

	s.t += value
	s.n++
}

// record the elapsed time since the supplied start time
func (s *TimeStats) Record(startTime time.Time) {
	s.Add(time.Since(startTime))
}

func (s *TimeStats) GetReport() TimeStatsReport {
	s.mx.Lock()
	defer s.mx.Unlock()

	r := TimeStatsReport{Total: s.n}
	if s.n > 0 {
		r.Average = s.t.Microseconds() / int64(s.n)
	}
	return r
}

func (s *TimeStats) GetReportJSON() ([]byte, error) {
	return json.Marshal(s.GetReport())
}
