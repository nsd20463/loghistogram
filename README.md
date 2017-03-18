
Log-scale histogram. Concurrency-safe and efficient.

Based on the ideas in github.com/codahale/hdrhistogram, which itself
is based on the ideas in some old java code. but not using any of the
implementation because I needed a histogram that handled floats,
that was concurrency-safe, and had an API to calculate multiple
percentiles in one pass for efficiency.

Copyright 2017 Nicolas Dade
