/*
  log-scaled histogram. concurrency-safe, and performant.

  Based on the ideas in github.com/codahale/hdrhistogram, which itself
  is based on the ideas in some old java code. but not using any of the
  implementation because that one only handles ints, doesn't deal
  with outliers, isn't thread-safe, and doesn't have an API which
  allows calculating of multiple statistics in a single pass.

  Copyright 2017 Nicolas Dade
*/
package loghistogram

import (
	"fmt"
	"math"
	"sync"
)

const epsilon = 1E-16 // 1E-16 is chosen because it is close to the ~52 bit limit of a float64 mantissa

// Histogram is a log-scaled histogram. It holds the accumulated counts
type Histogram struct {
	shift, scale float64 // precalculated values

	lock sync.Mutex // lock arbitrating access to counts and  n

	n                        uint64   // total # of accumulated samples in counts[], including outliers at counts[0] and counts[N+1]
	counts                   []uint64 // buckets of counts + a low and high outlier bucket at [0] and [N+1]
	middle_bucket_percentile float64  // guess (crude approximation) of the percentile of the values are in the first len(counts)/2 buckets, or -1 if it isn't yet guesses
}

// map from a value to a bucket index in h.counts. returns indexes 0 and = len(h.counts-1) for outliers
func (h *Histogram) valueToBucket(value float64) int {
	v := value - h.shift
	if v < 1.0 {
		return 0
	}

	b := 1 + int(math.Log(v)*h.scale) // benchmarks on amd64 & go1.7 show math.Log is slightly faster than math.Log10 and much faster than math.Log2

	if b >= len(h.counts) {
		b = len(h.counts) - 1
	}

	return b
}

// map from a bucket index into h.counts to the lower bound of values which map to that bucket
// if the bucket is an outlier then the result is not well defined, except that it will be outside the low,high range
func (h *Histogram) bucketToValue(bucket int) float64 {
	v := math.Exp(float64(bucket)/h.scale) + h.shift
	return v
}

// New constructs a histogram to hold values between low and high using the given number of buckets
func New(low, high float64, num_buckets int) *Histogram {
	h := &Histogram{}
	h.init(low, high, num_buckets)
	return h
}

func (h *Histogram) init(low, high float64, num_buckets int) {
	// check for nonsense arguments from broken callers
	if high < low || num_buckets <= 0 {
		panic(fmt.Sprintf("loghistogram.New(%v, %v, %v): invalid arguments", low, high, num_buckets))
	}

	// we want log(low-shift) to be 0, and log(high-shift)*scale = num_buckets+1-epsilon (so it falls inside the last bucket and not right on the edge)
	// so low-shift = 1, or
	shift := low - 1
	// and then
	scale := float64(num_buckets) * (1 - epsilon) / math.Log(high-shift)

	h.counts = make([]uint64, 2+num_buckets)
	h.shift = shift
	h.scale = scale
	h.middle_bucket_percentile = -1
}

func (h *Histogram) swap(new_counts []uint64) (old_counts []uint64) {
	old_counts = h.counts
	h.counts = new_counts
	return old_counts
}

// Accumulate adds a sample with value x to the histogram
func (h *Histogram) Accumulate(x float64) {
	i := h.valueToBucket(x)

	h.lock.Lock()
	h.counts[i]++
	h.n++
	h.lock.Unlock()
}

// test to see how much the lock hurts performance
func (h *Histogram) raceyAccumulate(x float64) {
	i := h.valueToBucket(x)

	h.counts[i]++
	h.n++
}

// Count returns the total number of samples accumulated, including outliers
func (h *Histogram) Count() uint64 {
	h.lock.Lock()
	n := h.n
	h.lock.Unlock()
	return n
}

// Outliers returns the number of outliers on either side (how may samples were outside the low...high bound)
func (h *Histogram) Outliers() (uint64, uint64) {
	h.lock.Lock()
	lo, hi := h.counts[0], h.counts[len(h.counts)-1]
	h.lock.Unlock()
	return lo, hi
}

// Percentiles returns the values at each percentile. NaN is returned if Count is 0 or percentiles are outside the 0...100 range.
// pers argument MUST be sorted low-to-high. NOTE outliers are taken into account as best we can, so the results can be outside
// of low...high if the percentile requested lies within the outliers.
func (h *Histogram) Percentiles(pers ...float64) []float64 {
	// check for stupid args
	if len(pers) == 0 {
		return nil
	}

	values := make([]float64, len(pers))

	// if the data values are evenly spread then scalling for percentiles starting from the highest
	// values to lower ones would be faster (since the high buckets are larger and would have more
	// of the total for fewer buckets scanned). But if you're using this log-scaled histogram rather
	// than a linear histogram it's probably because the distribution of values is skewed. In a common
	// use case of latency measurements, it's often very very skewed, with only a few outliers at the
	// top of the scale. Scanning for the 90% or 99% percentiles (often those of interest) can be
	// more efficient from below as from above, depending on the distribution.
	// A first good guess is to do it from below, but keeping track of the percentile of the middle
	// bucket lets us guess properly next time.

	h.lock.Lock()

	if h.middle_bucket_percentile >= 0 && pers[0] > h.middle_bucket_percentile {
		// find the percentiles from high to low. this can be more efficient when asking for things like the 99% percentile
		// because we only need to scan over 1% of the counts.
		// (the log-sized buckets can make the outliers efficient, even if there aren't a lot of them)
		n := h.n
		a := n
		if n == 0 {
			goto return_nans
		}
		nf := float64(n)
		i := len(h.counts) - 1
		for j := len(pers) - 1; j >= 0; j-- {
			p := pers[j]
			pn := uint64(p * nf / 100)
			for a >= pn && i >= 0 {
				a -= h.counts[i]
				i--
			}
			values[j] = h.bucketToValue(i)
		}
	} else {
		// find the percentiles from low to high
		a := uint64(0)
		n := h.n
		if n == 0 {
			goto return_nans
		}
		nf := float64(n)
		i := 0
		middle_bucket := len(h.counts) / 2
		for j, p := range pers {
			pn := uint64(p * nf / 100)
			for a < pn && i < len(h.counts) {
				a += h.counts[i]
				if i == middle_bucket {
					// update our estimate of the middle bucket's percentile
					h.middle_bucket_percentile = 100 * float64(a) / float64(n)
				}
				i++
			}
			values[j] = h.bucketToValue(i)
		}
	}

	h.lock.Unlock()
	return values

return_nans:
	h.lock.Unlock()
	nan := math.NaN()
	for i := range values {
		values[i] = nan
	}
	return values
}

// Percentile calculates one percentile
func (h *Histogram) Percentile(per float64) float64 {
	return h.Percentiles(per)[0]
}

// Dup returns a copy of h
func (h *Histogram) Dup() *Histogram {
	h.lock.Lock()
	h2 := *h
	// we've copied the struct, but of course not the counts slice
	// so copy that, and while we are at it we need to recompute n, just in case the counts change while we are copying them
	counts := make([]uint64, len(h.counts))
	for i := range counts {
		counts[i] = h.counts[i]
	}
	h2.counts = counts
	h.lock.Unlock()
	return &h2
}
