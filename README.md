
Log-scale histogram. Concurrency-safe and efficient.

Based on the ideas in github.com/codahale/hdrhistogram, which itself
is based on the ideas in some old java code. but not using any of the
implementation because that one only handles ints, and isn't thread-safe,
and doesn't have an API which allows calculating of many statistics
in a single pass.

Copyright 2017 Nicolas Dade
