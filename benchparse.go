package benchparse

import (
	"bufio"
	"errors"
	"io"
	"strconv"
	"strings"
	"unicode"
)

// Run is the entire parsed output of a single benchmark run
type Run struct {
	// Configuration is the key/value headers of the benchmark run
	Configuration KeyValueList
	// Results are the result of running each benchmark
	Results []BenchmarkResult
}

// KeyValueDecoder is used by Decoder to help it configure how to decode key/value pairs of a benchmark result
type KeyValueDecoder struct {
}

// BenchmarkResultDecoder is used by Decoder to help it configure how to decode individual benchmark runs
type BenchmarkResultDecoder struct {
}

var errNotEnoughFields = errors.New("invalid BenchmarkResult: not enough fields")
var errNoPrefixBenchmark = errors.New("invalid BenchmarkResult: no prefix benchmark")
var errUpperAfterBench = errors.New("invalid BenchmarkResult: no uppercase after benchmark name")
var errEvenFields = errors.New("invalid BenchmarkResult: expect even number of fields")

func (k *BenchmarkResultDecoder) decode(kvLine string) (*BenchmarkResult, error) {
	kvLine = strings.TrimSpace(kvLine)
	// https://github.com/golang/proposal/blob/master/design/14313-benchmark-format.md#benchmark-results
	fields := strings.Fields(kvLine)
	if len(fields) < 4 {
		return nil, errNotEnoughFields
	}
	if len(fields)%2 != 0 {
		return nil, errEvenFields
	}
	name := fields[0]
	if !strings.HasPrefix(name, "Benchmark") {
		return nil, errNoPrefixBenchmark
	}
	if name != "Benchmark" && !unicode.IsUpper(rune(name[len("Benchmark")])) {
		return nil, errUpperAfterBench
	}
	iterations, err := strconv.Atoi(fields[1])
	if err != nil {
		return nil, err
	}
	ret := &BenchmarkResult{
		Name:       name,
		Iterations: iterations,
	}
	for i := 2; i < len(fields); i += 2 {
		unit := fields[i+1]
		val, err := strconv.ParseFloat(fields[i], 64)
		if err != nil {
			return nil, err
		}
		ret.Values = append(ret.Values, BenchmarkValue{
			Value: val,
			Unit:  unit,
		})
	}
	return ret, nil
}

var errInvalidKeyValue = errors.New("invalid keyvalue: no colon")
var errInvalidKeyValueLowercase = errors.New("invalid keyvalue: expect lowercase start")
var errInvalidKeyValueEmpty = errors.New("invalid keyvalue: empty key")
var errInvalidKeyValueSpaces = errors.New("invalid keyvalue: key has spaces")
var errInvalidKeyValueUppercase = errors.New("invalid keyvalue: key has upper case chars")

func (k *KeyValueDecoder) decode(kvLine string) (*KeyValue, error) {
	// https://github.com/golang/proposal/blob/master/design/14313-benchmark-format.md#configuration-lines
	parts := strings.SplitN(kvLine, ":", 2)
	if len(parts) != 2 {
		return nil, errInvalidKeyValue
	}
	key := parts[0]
	value := parts[1]
	if len(key) == 0 {
		return nil, errInvalidKeyValueEmpty
	}
	if !unicode.IsLower(rune(key[0])) {
		return nil, errInvalidKeyValueLowercase
	}
	if strings.IndexFunc(key, unicode.IsSpace) != -1 {
		return nil, errInvalidKeyValueSpaces
	}
	if strings.IndexFunc(key, unicode.IsUpper) != -1 {
		return nil, errInvalidKeyValueUppercase
	}
	return &KeyValue{
		Key:   key,
		Value: strings.TrimLeftFunc(value, unicode.IsSpace),
	}, nil
}

// Decoder helps configure how to decode benchmark results.
type Decoder struct {
	KeyValueDecoder        KeyValueDecoder
	BenchmarkResultDecoder BenchmarkResultDecoder
}

// Decode an input stream into a benchmark run.  Returns an error if there are any issues decoding the benchmark,
// for example from reading from in.
func (d Decoder) Decode(in io.Reader) (*Run, error) {
	ret := &Run{}
	b := bufio.NewScanner(in)
	for b.Scan() {
		recentLine := strings.TrimSpace(b.Text())
		kv, err := d.KeyValueDecoder.decode(recentLine)
		if err == nil {
			ret.Configuration.keys = append(ret.Configuration.keys, *kv)
			continue
		}
		brun, err := d.BenchmarkResultDecoder.decode(recentLine)
		if err == nil {
			ret.Results = append(ret.Results, *brun)
		}
	}
	if b.Err() != nil {
		return nil, b.Err()
	}
	return ret, nil
}

// KeyValueList is an ordered list of possibly repeating key/value pairs where a key may happen more than once
type KeyValueList struct {
	keys []KeyValue
}

// AsMap returns the key value pairs as a map using only the first key's value if a key is duplicated in the list.
// Runs in O(N).
func (c KeyValueList) AsMap() map[string]string {
	ret := make(map[string]string)
	for _, k := range c.keys {
		if _, exists := ret[k.Key]; !exists {
			ret[k.Key] = k.Value
		}
	}
	return ret
}

// LookupAll returns all values for a single key.  Runs in O(N).
func (c KeyValueList) LookupAll(key string) []string {
	ret := make([]string, 0, 1)
	for _, k := range c.keys {
		if k.Key == key {
			ret = append(ret, k.Value)
		}
	}
	return ret
}

// Lookup a single key's value.  Returns false if the key does not exist (to distinguish from valid keys without values).
// Runs in O(N)
func (c KeyValueList) Lookup(key string) (string, bool) {
	for _, k := range c.keys {
		if k.Key == key {
			return k.Value, true
		}
	}
	return "", false
}

// Get a single key's value.  Returns empty string if the key does not exist.  If you want to know if the key existed,
// use Lookup or LookupAll.
func (c KeyValueList) Get(key string) string {
	ret, _ := c.Lookup(key)
	return ret
}

// KeyValue is a pair of key + value
type KeyValue struct {
	// The key of Key value pair
	Key string
	// The Value of key value pair
	Value string
}

func (k KeyValue) String() string {
	if k.Value == "" {
		return k.Key + ":"
	}
	return k.Key + ": " + k.Value
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
