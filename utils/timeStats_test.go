package utils

import "testing"

func TestTimeStatsAdd(t *testing.T) {
	s := TimeStats{}
	s.Add(100)

	if s.n != 1 || s.t != 100 {
		t.Errorf("Expected n=1, t=100 got: %v %v", s.n, s.t)
	}

	s.Add(50)
	if s.n != 2 || s.t != 150 {
		t.Errorf("Expected n=2, t=150 got: %v %v", s.n, s.t)
	}

	s.Add(25)
	if s.n != 3 || s.t != 175 {
		t.Errorf("Expected n=3, t=175 got: %v %v", s.n, s.t)
	}
}

func TestTimeStatsEmptyReport(t *testing.T) {

	s := TimeStats{}
	r := s.GetReport()

	if r.Total != 0 || r.Average != 0 {
		t.Errorf("Expected empty report, got: %v", r)
	}
}

func TestTimeStatsGetReport(t *testing.T) {

	s := TimeStats{}

	s.Add(10)
	s.Add(20)
	s.Add(30)

	r := s.GetReport()
	if r.Total != 3 {
		t.Errorf("expected total=3, got %v", r.Total)
	}
	if r.Average != 20 {
		t.Errorf("expected average=20, got %v", r.Average)
	}
}
