package benchparse

import (
	"strconv"
	"strings"
)

// Run is the entire parsed output of a single benchmark run
type Run struct {
	// Results are the result of running each benchmark
	Results []BenchmarkResult
}

// Value is the result of one (of possibly many) benchmark numeric computations
type Value struct {
	Value float64
	Unit  string
}

func (b Value) String() string {
	return strconv.FormatFloat(b.Value, 'f', -1, 64) + " " + b.Unit
}

// BenchmarkResult is a single line of a benchmark result
type BenchmarkResult struct {
	// Name of this benchmark.
	Name string
	// Iterations the benchmark run for.
	Iterations int
	// Values computed by this benchmark.  Has at least one member.
	Values []Value
	// Most benchmarks have the same configuration, but the spec allows a single set of benchmarks to have different
	// configurations.  Note that as a memory saving feature, multiple BenchmarkResult may share the same Configuration
	// data.  Do not modify the Configuration of any one BenchmarkResult unless you are **sure** they do not share the
	// same OrderedStringStringMap data's backing.
	Configuration OrderedStringStringMap
}

// NameAsKeyValue parses the name of the benchmark as a subtest/subbench split by / assuming you use
// key=value naming for each sub test.  One expected format may be "BenchmarkQuery/runs=1000/dist=normal"
func (b BenchmarkResult) NameAsKeyValue() OrderedStringStringMap {
	nameParts := strings.Split(b.Name, "/")
	var ret OrderedStringStringMap
	for _, p := range nameParts {
		sections := strings.SplitN(p, "=", 2)
		if len(sections) <= 1 {
			ret.add(sections[0], "")
		} else {
			ret.add(sections[0], sections[1])
		}
	}
	return ret
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
