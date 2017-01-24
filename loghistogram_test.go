package loghistogram

import "testing"

func TestAccumulate(t *testing.T) {
	h := New(0, 10, 10)
	for i := float64(0); i <= 10; i++ {
		h.Accumulate(i)
	}

	if h.Count() != 11 {
		t.Errorf("h.Count %d != 11", h.Count())
	}

	pers := h.Percentiles(0, 50, 90, 99, 100)
	for i, e := range []float64{0, 5, 9, 9, 10} {
		if pers[i] != e {
			t.Errorf("percentile[%d] %v != %v", i, pers[i], e)
		}
	}
}

func TestOutliers(t *testing.T) {
	h.New(-10, 10, 10)
	h.Accumulate(-10.0001)
	h.Accumulate(-99)
	h.Accumulate(10.0001)
	if h.Count() != 0 {
		t.Error("Count", h.Count())
	}
	if h.LowOutliers() != 2 || h.HighOutliers() != 1 || h.Outliers() != 3 {
		t.Error("Outliers", h.LowOutliers(), h.HighOutliers(), h.Outliers())
	}
}

func TestSubtract(t *testing.T) {
	h := New(0, 10, 100)
	h.Accumulate(4)
	h2 := h.Dup()
	h.Accumulate(5)
	if h.Count() != 2 {
		t.Error("Count", h.Count())
	}
	if h.Percentiles(50) != 4.5 {
		t.Error("median", h.Percentiles(50))
	}

	h.Sub(h2)
	if h.Count() != 1 {
		t.Error("Count", h.Count())
	}
	if h.Percentiles(50) != 5 {
		t.Error("median", h.Percentiles(50))
	}
}

func BenchmarkAccumulate(b *testing.B) {
	h := New(0, 10000000, 1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Accumulate(float64(i))
	}
}

func BenchmarkSinglePercentile(b *testing.B) {
	h := New(0, 10000, 1000)
	for i := 0; i < 10000; i += 10 {
		h.Accumulate(float64(i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Percentiles(50)
	}
}

func Benchmark3Percentiles(b *testing.B) {
	h := New(0, 10000, 1000)
	for i := 0; i < 10000; i += 10 {
		h.Accumulate(float64(i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Percentiles(25, 50, 75)
	}
}

func Benchmark99thPercentile(b *testing.B) {
	h := New(0, 10000, 1000)
	for i := 0; i < 10000; i += 10 {
		h.Accumulate(float64(i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Percentiles(99)
	}
}
