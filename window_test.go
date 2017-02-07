package loghistogram

import (
	"math"
	"testing"
)

func TestWindowedAccumulate(t *testing.T) {
	h := NewWindowed(0, 100, 1000)
	for i := float64(0); i <= 100; i++ {
		h.Accumulate(i)
	}

	lo, hi := h.Outliers()
	if (lo | hi) != 0 {
		t.Errorf("h.Outliers() %d,%d != 0", lo, hi)
	}
	if h.Count() != 101 {
		t.Errorf("h.Count %d != 11", h.Count())
	}

	pers := []float64{0, 50, 90, 99, 100}
	vals := h.Percentiles(pers...)
	t.Logf("vals = %f\n", vals)
	for i, e := range []float64{0, 50, 90, 99, 100} {
		if pers[i] != e {
			t.Errorf("percentile[%v (%v%%)] %v != %v", i, pers[i], vals[i], e)
		}
	}
}

func TestWindowedOutliers(t *testing.T) {
	h := NewWindowed(-10, 10, 10)
	h.Accumulate(-10.0001)
	h.Accumulate(-99)
	h.Accumulate(10.0001)
	if h.Count() != 3 {
		t.Error("Count", h.Count())
	}
	lo, hi := h.Outliers()
	if lo != 2 || hi != 1 {
		t.Error("Outliers", lo, hi)
	}
}

func TestWindowedEmptyHistogram(t *testing.T) {
	h := NewWindowed(0, 1, 10)
	p := h.Percentile(50)
	if !math.IsNaN(p) {
		t.Error("Percentile() of empty histogram != NaN")
	}
	ps := h.Percentiles(0, 100)
	for _, p := range ps {
		if !math.IsNaN(p) {
			t.Error("Percentiles() of empty histogram != NaN")
		}
	}
}

func BenchmarkWindowedAccumulate(b *testing.B) {
	h := NewWindowed(0, 10000000, 1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Accumulate(float64(i))
	}
}

func BenchmarkWindowedRaceyAccumulate(b *testing.B) {
	h := NewWindowed(0, 10000000, 1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.raceyAccumulate(float64(i))
	}
}

func BenchmarkWindowedSingle10thPercentile(b *testing.B) {
	h := NewWindowed(0, 10000, 1000)
	for i := 0; i < 10000; i += 10 {
		h.Accumulate(float64(i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Percentile(10)
	}
}

func BenchmarkWindowedSingle25thPercentile(b *testing.B) {
	h := NewWindowed(0, 10000, 1000)
	for i := 0; i < 10000; i += 10 {
		h.Accumulate(float64(i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Percentile(25)
	}
}

func BenchmarkWindowedSingle50thPercentile(b *testing.B) {
	h := NewWindowed(0, 10000, 1000)
	for i := 0; i < 10000; i += 10 {
		h.Accumulate(float64(i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Percentile(50)
	}
}

func BenchmarkWindowedSingle75thPercentile(b *testing.B) {
	h := NewWindowed(0, 10000, 1000)
	for i := 0; i < 10000; i += 10 {
		h.Accumulate(float64(i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Percentile(75)
	}
}

func BenchmarkWindowed2Percentiles(b *testing.B) {
	h := NewWindowed(0, 10000, 1000)
	for i := 0; i < 10000; i += 10 {
		h.Accumulate(float64(i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Percentiles(25, 50)
	}
}

func BenchmarkWindowed3Percentiles(b *testing.B) {
	h := NewWindowed(0, 10000, 1000)
	for i := 0; i < 10000; i += 10 {
		h.Accumulate(float64(i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Percentiles(25, 33, 50)
	}
}

func BenchmarkWindowed4Percentiles(b *testing.B) {
	h := NewWindowed(0, 10000, 1000)
	for i := 0; i < 10000; i += 10 {
		h.Accumulate(float64(i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Percentiles(10, 25, 33, 50)
	}
}

func BenchmarkWindowed75thPercentile(b *testing.B) {
	h := NewWindowed(0, 10000, 1000)
	for i := 0; i < 10000; i += 10 {
		h.Accumulate(float64(i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Percentiles(75)
	}
}

func BenchmarkWindowed99thPercentile(b *testing.B) {
	h := NewWindowed(0, 10000, 1000)
	for i := 0; i < 10000; i += 10 {
		h.Accumulate(float64(i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Percentiles(99)
	}
}
