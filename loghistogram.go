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

// New constructs a histogram to hold values between low and high using the given number of buckets
func New(low, high float64, buckets int) *Histogram {
	return nil
}

type Histogram struct{}

// Accumulate adds a sample with value x to the histogram
func (h *Histogram) Accumulate(x float64) {}

// Count returns the total number of samples accumulated
func (h *Histogram) Count() uint64 { return 0 }

// Percentiles returns the values at each percentile. Nan is returned if Count is 0 or percentiles are outside the 0...100 range.
// pers argument MUST be sorted low-to-high.
func (h *Histogram) Percentiles(pers []float64) []float64 { return nil }

// Dup returns a copy of h
func (h *Histogram) Dup() *Histogram { return nil }

// Sub subtracts h2 from h in-place. h -= h2.
// Subtracting an earlier copy of the histogram is useful when keeping a running histogram.
func (h *Histogram) Sub(h2 *Histogram) {}

// Sub returns h1-h2 without changing h1 nor h2
func Sub(h1, h2 *Histogram) *Histogram {
	return nil
}
