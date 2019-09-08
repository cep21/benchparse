/*
Package benchparse allows you to easily parse the output format of Go's benchmark results, as well as other outputs
that conform to the benchmark spec.  The entire spec is documented at is defined at https://github.com/golang/proposal/blob/master/design/14313-benchmark-format.md.

Proper use of this library is to pass a io stream into Decode to decode a benchmark run into results.  If you want
to modify those results and later encode them back out, make a deep copy of the BenchmarkResult object since by default,
Decode will share pointers to Configuration objects.
*/
package benchparse
