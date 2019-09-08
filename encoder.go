package benchparse

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
)

// Decoder helps configure how to decode benchmark results.
type Decoder struct {
	keyValueDecoder        keyValueDecoder
	benchmarkResultDecoder benchmarkResultDecoder
}

// Decode an input stream into a benchmark run.  Returns an error if there are any issues decoding the benchmark,
// for example from reading from in.  The returned run is **NOT** intended to be modified.  It contains public members
// for API convenience, and will share OrderedStringStringMap values to reduce memory allocations.  Do not modify
// the returned Run and expect it to do anything you are wanting it to do.  Instead, create your own Run object and
// assign values to it as you want.
func (d Decoder) Decode(in io.Reader) (*Run, error) {
	ret := &Run{}
	b := bufio.NewScanner(in)
	currentKeys := new(OrderedStringStringMap)
	currentConfigurationIsDirty := false
	for b.Scan() {
		recentLine := strings.TrimSpace(b.Text())
		kv, err := d.keyValueDecoder.decode(recentLine)
		if err == nil {
			if currentConfigurationIsDirty {
				currentKeys = currentKeys.clone()
				currentConfigurationIsDirty = false
			}
			currentKeys.add(kv.Key, kv.Value)
			continue
		}
		brun, err := d.benchmarkResultDecoder.decode(recentLine)
		if err == nil {
			brun.Configuration = currentKeys
			currentConfigurationIsDirty = true
			ret.Results = append(ret.Results, *brun)
		}
	}
	if b.Err() != nil {
		return nil, b.Err()
	}
	return ret, nil
}

// benchmarkResultDecoder is used by Decoder to help it configure how to decode individual benchmark runs
type benchmarkResultDecoder struct {
}

var errNotEnoughFields = errors.New("invalid BenchmarkResult: not enough fields")
var errNoPrefixBenchmark = errors.New("invalid BenchmarkResult: no prefix benchmark")
var errUpperAfterBench = errors.New("invalid BenchmarkResult: no uppercase after benchmark name")
var errEvenFields = errors.New("invalid BenchmarkResult: expect even number of fields")

func (k *benchmarkResultDecoder) decode(kvLine string) (*BenchmarkResult, error) {
	kvLine = strings.TrimSpace(kvLine)
	// https://github.com/golang/proposal/blob/master/design/14313-benchmark-format.md#benchmark-results
	// Note: I thought about using a regex here, but the spec mentions specific functions so I use those directly.
	// "The fields are separated by runs of space characters (as defined by unicode.IsSpace), so the line can be parsed with strings.Fields."
	fields := strings.Fields(kvLine)
	// "The line must have an even number of fields, and at least four."
	if len(fields) < 4 {
		return nil, errNotEnoughFields
	}
	if len(fields)%2 != 0 {
		return nil, errEvenFields
	}
	// "The first field is the benchmark name, which must begin with Benchmark"
	name := fields[0]
	if !strings.HasPrefix(name, "Benchmark") {
		return nil, errNoPrefixBenchmark
	}
	// "followed by an upper case character (as defined by unicode.IsUpper) or the end of the field, as in BenchmarkReverseString or just Benchmark."
	if name != "Benchmark" && !unicode.IsUpper(rune(name[len("Benchmark")])) {
		return nil, errUpperAfterBench
	}
	// "The second field gives the number of iterations run"
	iterations, err := strconv.Atoi(fields[1])
	if err != nil {
		return nil, err
	}
	ret := &BenchmarkResult{
		Name:       name,
		Iterations: iterations,
	}
	// "fields report value/unit pairs"
	for i := 2; i < len(fields); i += 2 {
		unit := fields[i+1]
		// "in which the value is a float64 that can be parsed by strconv.ParseFloat"
		val, err := strconv.ParseFloat(fields[i], 64)
		if err != nil {
			return nil, err
		}
		ret.Values = append(ret.Values, ValueUnitPair{
			Value: val,
			Unit:  unit,
		})
	}
	return ret, nil
}

var errInvalidKeyValueLowercase = errors.New("invalid keyvalue: expect lowercase start")
var errInvalidKeyValueEmpty = errors.New("invalid keyvalue: empty key")
var errInvalidKeyValueSpaces = errors.New("invalid keyvalue: key has spaces or upper case")
var errInvalidKeyNoColon = errors.New("invalid keyvalue: key has no colon")
var errInvalidKeyValueReturn = errors.New("invalid keyvalue: value has newline")

func (k *keyValueDecoder) decode(kvLine string) (*keyValue, error) {
	// https://github.com/golang/proposal/blob/master/design/14313-benchmark-format.md#configuration-lines
	// Note: I thought about using a regex here, but the spec mentions specific functions so I use those directly.
	// "a key-value pair of the form `key: value`
	firstColon := strings.Index(kvLine, ":")
	if firstColon == -1 {
		return nil, errInvalidKeyNoColon
	}
	key := kvLine[:firstColon]
	// Key can have spaces after the colon.  They should be removed.
	// "one or more ASCII space or tab characters separate “key:” from “value.”
	value := strings.TrimLeftFunc(kvLine[firstColon+1:], func(r rune) bool {
		return r == ' ' || r == '\t'
	})
	// "where key begins with a lower case character"
	if len(key) == 0 {
		return nil, errInvalidKeyValueEmpty
	}
	// "where key begins with a lower case character (as defined by unicode.IsLower)"
	if !unicode.IsLower(rune(key[0])) {
		return nil, errInvalidKeyValueLowercase
	}
	//  contains no space characters (as defined by unicode.IsSpace) nor upper case characters (as defined by unicode.IsUpper)
	if strings.IndexFunc(key, func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsUpper(r)
	}) != -1 {
		return nil, errInvalidKeyValueSpaces
	}
	// "There are no restrictions on value, except that it cannot contain a newline character"
	if strings.Contains(value, "\n") {
		return nil, errInvalidKeyValueReturn
	}
	return &keyValue{
		Key:   key,
		Value: strings.TrimLeftFunc(value, unicode.IsSpace),
	}, nil
}

// keyValueDecoder is used by Decoder to help it configure how to decode key/value pairs of a benchmark result
type keyValueDecoder struct {
}

type Encoder struct {
}

func (e *Encoder) Encode(w io.Writer, run *Run) error {
	var previousConfig *OrderedStringStringMap
	for _, r := range run.Results {
		transition := previousConfig.valuesToTransition(r.Configuration)
		for i := range transition.Order {
			if _, err := fmt.Fprintf(w, "%s: %s\n", transition.Order[i], transition.Contents[transition.Order[i]]); err != nil {
				return err
			}
		}
		previousConfig = r.Configuration
		if _, err := fmt.Fprintf(w, "%s\n", r.String()); err != nil {
			return err
		}
	}
	return nil
}

// keyValue is a pair of key + value
type keyValue struct {
	// The key of Key value pair
	Key string
	// The Value of key value pair
	Value string
}

func (k keyValue) String() string {
	if k.Value == "" {
		return k.Key + ":"
	}
	return k.Key + ": " + k.Value
}
