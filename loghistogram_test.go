package loghistogram

import (
	"math"
	"testing"
	"unsafe"
)

func TestAccumulate(t *testing.T) {
	h := New(0, 100, 1000)
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

func TestOutliers(t *testing.T) {
	h := New(-10, 10, 10)
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

func BenchmarkFloatBits(b *testing.B) {
	for i := 0; i < b.N; i++ {
		f := float64(i)
		f2 := math.Float64frombits(math.Float64bits(f))
		if f != f2 {
			b.Error(f, "!=", f2)
			break
		}
	}
}
func BenchmarkFloatCast(b *testing.B) {
	for i := 0; i < b.N; i++ {
		f := float64(i)
		n := *(*uint64)(unsafe.Pointer(&f))
		f2 := *(*float64)(unsafe.Pointer(&n))
		if f != f2 {
			b.Error(f, "!=", f2)
			break
		}
	}
}
