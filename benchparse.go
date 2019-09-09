package benchparse

import (
	"strconv"
	"strings"
	"unicode"
)

const (
	// UnitRuntime is the default unit for Go's runtime benchmark.  You're intended to call it with ValueByUnit.
	UnitRuntime = "ns/op"
	// UnitBytesAlloc is the default unit for Go's memory allocated benchmark.  You're intended to call it with ValueByUnit.
	UnitBytesAlloc = "B/op"
	// UnitObjectAllocs is the default unit for Go's # of allocs benchmark.  You're intended to call it with ValueByUnit.
	UnitObjectAllocs = "allocs/op"
)

// Run is the entire parsed output of a single benchmark run
type Run struct {
	// Results are the result of running each benchmark
	Results []BenchmarkResult
}

// BenchmarkResult is a single line of a benchmark result
type BenchmarkResult struct {
	// Name of this benchmark.
	Name string
	// Iterations the benchmark run for.
	Iterations int
	// Values computed by this benchmark.  len(Values) >= 1.
	Values []ValueUnitPair
	// Most benchmarks have the same configuration, but the spec allows a single set of benchmarks to have different
	// configurations.  Note that as a memory saving feature, multiple BenchmarkResult may share the same Configuration
	// data by pointing to the same OrderedStringStringMap.  Do not modify the Configuration of any one BenchmarkResult
	// unless you are **sure** they do not share the same OrderedStringStringMap data's backing.
	Configuration *OrderedStringStringMap
}

// ValueUnitPair is the result of one (of possibly many) benchmark numeric computations
type ValueUnitPair struct {
	// Value is the numeric result of a benchmark
	Value float64
	// Unit is the units this value is in
	Unit string
}

func (b ValueUnitPair) String() string {
	return strconv.FormatFloat(b.Value, 'f', -1, 64) + " " + b.Unit
}

// NameAsKeyValue parses the name of the benchmark as a subtest/subbench split by / assuming you use
// key=value naming for each sub test.  One expected format may be "BenchmarkQuery/runs=1000/dist=normal".  For
// pairs that do not contain a =, like "BenchmarkQuery" above, they will be stored inside OrderedStringStringMap with
// the key as their name and an empty value.  If multiple keys are used (which is not recommended), then the last key's
// value will be returned.
//
// Note that there is one special case handling.  Many go benchmarks append a "-N" number to the end of the benchmark
// name.  This can throw off key handling.  If you want to ignore this, you'll have to check the last value in your
// returned map.
func (b BenchmarkResult) NameAsKeyValue() *OrderedStringStringMap {
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
	return &ret
}

// AllKeyValuePairs returns the combination of the configuration key/value pairs followed by the benchmark name's
// key/value pairs.  It handles the special case of -N at the end of the last benchmark key/value pair by removing
// anything matching "-(\d+)" from the last key/value pair of the benchmark name.
func (b BenchmarkResult) AllKeyValuePairs() *OrderedStringStringMap {
	var ret OrderedStringStringMap
	if b.Configuration != nil {
		for _, p := range b.Configuration.Order {
			ret.add(p, b.Configuration.Contents[p])
		}
	}
	namePart := b.NameAsKeyValue()
	for i, p := range namePart.Order {
		if i != len(namePart.Order)-1 {
			ret.add(p, namePart.Contents[p])
			continue
		}
		lastValue := namePart.Contents[p]
		lastDash := strings.LastIndex(lastValue, "-")
		if lastDash == -1 {
			// No "-" means it doesn't match the pattern -N
			ret.add(p, namePart.Contents[p])
			continue
		}
		partAfterDash := lastValue[lastDash+1:]
		if len(partAfterDash) == 0 || strings.IndexFunc(partAfterDash, func(r rune) bool {
			return !unicode.IsNumber(r)
		}) != -1 {
			// Anything after the last - that isn't a number doesn't match the pattern either.
			ret.add(p, namePart.Contents[p])
			continue
		}
		ret.add(p, lastValue[0:lastDash])
	}
	return &ret
}

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
