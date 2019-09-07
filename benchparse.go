package benchparse

import (
	"strconv"
	"strings"
)

// Run is the entire parsed output of a single benchmark run
type Run struct {
	// Configuration is the key/value headers of the benchmark run
	Configuration KeyValueList
	// Results are the result of running each benchmark
	Results []BenchmarkResult
}

// BenchmarkValue is the result of one (of possibly many) benchmark numeric computations
type BenchmarkValue struct {
	Value float64
	Unit  string
}

func (b BenchmarkValue) String() string {
	return strconv.FormatFloat(b.Value, 'f', -1, 64) + " " + b.Unit
}

// BenchmarkResult is a single line of a benchmark result
type BenchmarkResult struct {
	// Name of this benchmark
	Name string
	// Iterations the benchmark run for
	Iterations int
	// Values computed by this benchmark.  Has at least one member
	Values []BenchmarkValue
}

// NameAsKeyValue parses the name of the benchmark as a subtest/subbench split by / assuming you use
// key=value naming for each sub test.
func (b BenchmarkResult) NameAsKeyValue() KeyValueList {
	nameParts := strings.Split(b.Name, "/")
	var keys []KeyValue
	for _, p := range nameParts {
		sections := strings.SplitN(p, "=", 2)
		if len(sections) <= 1 {
			keys = append(keys, KeyValue{
				Key: p,
			})
		} else {
			keys = append(keys, KeyValue{
				Key:   sections[0],
				Value: sections[1],
			})
		}
	}
	return KeyValueList{keys: keys}
}

// BaseName returns the benchmark name with Benchmark trimmed off.  Can possibly be empty string.
func (b BenchmarkResult) BaseName() string {
	return strings.TrimPrefix(b.Name, "Benchmark")
}

// UnitRuntime is the default unit for Go's runtime benchmark.  You're intended to call it with ValueByUnit.
const UnitRuntime = "ns/op"
// UnitRuntime is the default unit for Go's memory allocated benchmark.  You're intended to call it with ValueByUnit.
const UnitBytesAlloc = "B/op"
// UnitRuntime is the default unit for Go's # of allocs benchmark.  You're intended to call it with ValueByUnit.
const UnitObjectAllocs = "allocs/op"

// ValueByUnit returns the first value associated with a unit.  Returns false if the unit did not exist.
func (b BenchmarkResult) ValueByUnit(unit string) (float64, bool) {
	for _, v := range b.Values {
		if unit == v.Unit {
			return v.Value, true
		}
	}
	return 0.0, false
}

func (b BenchmarkResult) String() string {
	ret := make([]string, 0, len(b.Values)+2)
	ret = append(ret, b.Name, strconv.Itoa(b.Iterations))
	for _, v := range b.Values {
		ret = append(ret, v.String())
	}
	return strings.Join(ret, " ")
}
