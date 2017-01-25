/*
  log-scaled histogram. concurrency-safe, and performant.

  Based on the ideas in github.com/codahale/hdrhistogram, which itself
  is based on the ideas in some old java code. but not using any of the
  implementation because that one only handles ints, and isn't thread-safe,
  and doesn't have an API which allows calculating of many statistics
  in a single pass.

  Copyright 2017 Nicolas Dade
*/
package loghistogram

import "sync/atomic"

// Histogram is a log-scaled histogram. It holds the accumulated counts
type Histogram struct {
	n                           uint64   // total # of accumulated samples in counts[]. does NOT include outliers
	counts                      []uint64 // buckets of counts
	low_outliers, high_outliers uint64   // counts of <low and >high outliers

	low, high float64 // the low and high bounds of the histogram (set when created)
}

// map from a value to a bucket index in h.counts
func (h *Histogram) valueToBucket(value float64) int {
	return 0

}

// map from a bucket index in h.counts to the center accumulated value mapping to that bucket
func (h *Histogram) bucketToValue(bucket int) float64 {
	return 0.0
}

// New constructs a histogram to hold values between low and high using the given number of buckets
func New(low, high float64, num_buckets int) *Histogram {
	h := &Histogram{
		low:  low,
		high: high,
	}

	h.counts = make([]uint64, num_buckets)

	return h
}

// Accumulate adds a sample with value x to the histogram
func (h *Histogram) Accumulate(x float64) {
	// outliers are accumulated separately just so we know they exist, and count count towards 'N'
	if x < h.low {
		atomic.AddUint64(&h.low_outliers, 1)
		return
	} else if x > h.high {
		atomic.AddUint64(&h.high_outliers, 1)
		return
	}

	i := h.valueToBucket(x)
	atomic.AddUint64(&h.counts[i], 1)
	atomic.AddUint64(&h.n, 1)
}

// Count returns the total number of samples accumulated within low...high. outliers are not included
func (h *Histogram) Count() uint64 { return atomic.LoadUint64(&h.n) }

// Outliers returns the number of outliers (how may samples were outside the low...high bound)
func (h *Histogram) Outliers() uint64 {
	return atomic.LoadUint64(&h.low_outliers) + atomic.LoadUint64(&h.high_outliers)
}
func (h *Histogram) LowOutliers() uint64 {
	return atomic.LoadUint64(&h.low_outliers)
}
func (h *Histogram) HighOutliers() uint64 {
	return atomic.LoadUint64(&h.high_outliers)
}

// Percentiles returns the values at each percentile. Nan is returned if Count is 0 or percentiles are outside the 0...100 range.
// pers argument MUST be sorted low-to-high.
func (h *Histogram) Percentiles(pers ...float64) []float64 {
	if len(pers) == 0 { // check for stupid args
		return nil
	}

	values := make([]float64, len(pers))

	if pers[0] < 50 {
		// find the percentiles from low to high
		a := atomic.LoadUint64(&h.low_outliers)
		n := a + atomic.LoadUint64(&h.n) + atomic.LoadUint64(&h.high_outliers)
		nf := float64(n)
		i := 0
		for j, p := range pers {
			pn := uint64(p*nf/100 + 0.5)
			for a < pn && i < len(h.counts) {
				a += atomic.LoadUint64(&h.counts[i])
				i++
			}
			values[j] = h.bucketToValue(i)
		}
	} else {
		// find the percentiles from high to low. this can be more efficient when asking for things like the 99% percentile
		// because we only need to scan over 1% of the counts.
		// (the log-sized buckets ought to make the outliers efficient, even if there aren't a lot of them)
		a := atomic.LoadUint64(&h.low_outliers) + atomic.LoadUint64(&h.n)
		n := a + +atomic.LoadUint64(&h.low_outliers)
		nf := float64(n)
		i := len(h.counts) - 1
		for j := len(pers) - 1; j >= 0; j-- {
			p := pers[j]
			pn := uint64(p*nf/100 + 0.5)
			for a >= pn && i >= 0 {
				a -= atomic.LoadUint64(&h.counts[i])
				i--
			}
			values[j] = h.bucketToValue(i)
		}
	}

	return values
}

// Percentile calculates one percentile
func (h *Histogram) Percentile(per float64) float64 {
	return h.Percentiles(per)[0]
}

// Dup returns a copy of h
func (h *Histogram) Dup() *Histogram {
	h2 := *h
	// we've copied the struct, but of course not the counts slice
	// so copy that, and while we are at it we need to recompute n, just in case the counts change while we are copying them
	counts := make([]uint64, len(h2.counts))
	n := uint64(0)
	for i := range counts {
		c := atomic.LoadUint64(&h2.counts[i])
		n += c
		counts[i] = c
	}
	h2.counts = counts
	h2.n = n
	return &h2
}

// Sub subtracts h2 from h in-place. h -= h2. h and h2 must be the same size or you're subtracting apples from oranges and you'll get garbage
// Subtracting an earlier copy of the histogram is useful when keeping a running histogram.
func (h *Histogram) Sub(h2 *Histogram) {
	if len(h.counts) != len(h2.counts) {
		panic("subtracting different-sized histograms")
	}
	// I could also check the low and high, but that's sometimes useful, so don't

	for i := range h2.counts {
		c := atomic.LoadUint64(&h2.counts[i])
		atomic.AddUint64(&h.counts[i], -c)
		atomic.AddUint64(&h.n, -c) // keep the 'n' as up-to-date as Accumulate does, rather than adjust n once at the end of the loop
	}
}

// Sub returns h1-h2 without changing h1 nor h2
func Sub(h1, h2 *Histogram) *Histogram {
	if len(h1.counts) != len(h2.counts) {
		panic("subtracting different-sized histograms")
	}
	// I could also check the low and high, but that's sometimes useful, so don't

	h := *h1
	h.counts = make([]uint64, len(h1.counts))
	n := uint64(0)
	for i := range h1.counts {
		c1 := atomic.LoadUint64(&h1.counts[i])
		c2 := atomic.LoadUint64(&h2.counts[i])
		h.counts[i] = c1 - c2
		n += c1 - c2
	}
	h.n = n

	return &h
}
