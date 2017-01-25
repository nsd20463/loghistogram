package loghistogram

import (
	"math"
	"testing"
)

func TestAccumulate(t *testing.T) {
	h := New(0, 100, 1000)
	for i := float64(0); i <= 100; i++ {
		h.Accumulate(i)
	}

	if h.Outliers() != 0 {
		t.Errorf("h.Outliers() %d != 0", h.Outliers())
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

func TestOutliers(t *testing.T) {
	h := New(-10, 10, 10)
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

func TestEmptyHistogram(t *testing.T) {
	h := New(0, 1, 10)
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

func TestSubtract(t *testing.T) {
	h := New(0, 10, 100)
	h.Accumulate(1)
	h.Accumulate(2)
	h.Accumulate(3)
	h.Accumulate(4)
	h2 := h.Dup()
	h.Accumulate(7)
	h.Accumulate(8)
	h.Accumulate(9)
	if h.Count() != 7 {
		t.Error("Count", h.Count())
	}
	p1 := h.Percentile(50)

	t.Logf("h = %+v\n", h)
	t.Logf("h2 = %+v\n", h2)
	h.Sub(h2)
	t.Logf("h-h2 = %+v\n", h)

	if h.Count() != 3 {
		t.Error("Count", h.Count())
	}
	p2 := h.Percentile(50)
	if p1 == p2 || p1 > p2 {
		t.Error("median before", p1, ", after sub", p2)
	}
}

func BenchmarkAccumulate(b *testing.B) {
	h := New(0, 10000000, 1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Accumulate(float64(i))
	}
}

func BenchmarkRaceyAccumulate(b *testing.B) {
	h := New(0, 10000000, 1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.raceyAccumulate(float64(i))
	}
}

func BenchmarkSingle10thPercentile(b *testing.B) {
	h := New(0, 10000, 1000)
	for i := 0; i < 10000; i += 10 {
		h.Accumulate(float64(i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Percentile(10)
	}
}

func BenchmarkSingle25thPercentile(b *testing.B) {
	h := New(0, 10000, 1000)
	for i := 0; i < 10000; i += 10 {
		h.Accumulate(float64(i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Percentile(25)
	}
}

func BenchmarkSingle50thPercentile(b *testing.B) {
	h := New(0, 10000, 1000)
	for i := 0; i < 10000; i += 10 {
		h.Accumulate(float64(i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Percentile(50)
	}
}

func BenchmarkSingle75thPercentile(b *testing.B) {
	h := New(0, 10000, 1000)
	for i := 0; i < 10000; i += 10 {
		h.Accumulate(float64(i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Percentile(75)
	}
}

func Benchmark2Percentiles(b *testing.B) {
	h := New(0, 10000, 1000)
	for i := 0; i < 10000; i += 10 {
		h.Accumulate(float64(i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Percentiles(25, 50)
	}
}

func Benchmark3Percentiles(b *testing.B) {
	h := New(0, 10000, 1000)
	for i := 0; i < 10000; i += 10 {
		h.Accumulate(float64(i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Percentiles(25, 33, 50)
	}
}

func Benchmark4Percentiles(b *testing.B) {
	h := New(0, 10000, 1000)
	for i := 0; i < 10000; i += 10 {
		h.Accumulate(float64(i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Percentiles(10, 25, 33, 50)
	}
}

func Benchmark75thPercentile(b *testing.B) {
	h := New(0, 10000, 1000)
	for i := 0; i < 10000; i += 10 {
		h.Accumulate(float64(i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Percentiles(75)
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

func BenchmarkLog10(b *testing.B) {
	for i := 0; i < b.N; i++ {
		math.Log10(float64(i + 1))
	}
}

func BenchmarkLog2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		math.Log2(float64(i + 1))
	}
}

func BenchmarkLog(b *testing.B) {
	for i := 0; i < b.N; i++ {
		math.Log(float64(i + 1))
	}
}
