/*
  windowed log-scaled histogram.

  The windowing is simple. We rotate the histogram every
  1/2 window, keeping the previous histogram around.
  Statistics are calculated by summing across both
  the current and the previous windows. For histograms
  this is fine as long as a statistically sufficient number
  of events happen in 1/2 a window period.

  Copyright 2017 Nicolas Dade
*/

package loghistogram

import "math"

type WindowedHistogram struct {
	Histogram

	prev struct { // previous window's data
		n      uint64
		counts []uint64
	}
}

func NewWindowed(low, high float64, num_buckets int) *WindowedHistogram {
	h := &WindowedHistogram{}
	h.Histogram.init(low, high, num_buckets)
	h.prev.counts = make([]uint64, len(h.counts)) // make a dummy, all-zeroed prev array so we don't have to think when rotating

	return h
}

func (h *WindowedHistogram) Window() {
	h.lock.Lock()

	// zero the previous array. we reuse it to avoid creating garbage
	for i := range h.prev.counts {
		h.prev.counts[i] = 0
	}
	h.prev.counts, h.counts = h.counts, h.prev.counts
	h.prev.n, h.n = h.n, 0

	h.lock.Unlock()
}

func (h *WindowedHistogram) Count() uint64 {
	h.lock.Lock()
	n := h.n + h.prev.n
	h.lock.Unlock()
	return n
}

func (h *WindowedHistogram) Outliers() (uint64, uint64) {
	h.lock.Lock()
	lo := h.counts[0] + h.prev.counts[0]
	hi := h.counts[len(h.counts)-1] + h.prev.counts[len(h.counts)-1]
	h.lock.Unlock()
	return lo, hi
}

func (h *WindowedHistogram) Percentiles(pers ...float64) []float64 {
	if len(pers) == 0 {
		return nil
	}

	values := make([]float64, len(pers))

	h.lock.Lock()
	middle_bucket := len(h.counts) / 2
	if h.middle_bucket_percentile >= 0 && pers[0] > h.middle_bucket_percentile {
		n := h.n + h.prev.n
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
				if i == middle_bucket {
					h.middle_bucket_percentile = 100 * float64(a) / float64(n)
				}
				a -= h.counts[i] + h.prev.counts[i]
				i--
			}
			values[j] = h.bucketToValue(i)
		}
	} else {
		a := uint64(0)
		n := h.n + h.prev.n
		if n == 0 {
			goto return_nans
		}
		nf := float64(n)
		i := 0
		for j, p := range pers {
			pn := uint64(p * nf / 100)
			for a < pn && i < len(h.counts) {
				a += h.counts[i] + h.prev.counts[i]
				if i == middle_bucket {
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

func (h *WindowedHistogram) Percentile(per float64) float64 {
	return h.Percentiles(per)[0]
}
