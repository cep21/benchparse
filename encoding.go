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
var errInvalidKeyNoColon = errors.New("invalid keyvalue: key has no colon")

func (k *KeyValueDecoder) decode(kvLine string) (*KeyValue, error) {
	// https://github.com/golang/proposal/blob/master/design/14313-benchmark-format.md#configuration-lines
	firstSpace := strings.IndexFunc(kvLine, unicode.IsSpace)
	if firstSpace == -1 {
		return nil, errInvalidKeyValue
	}
	// Key *must* start at the first space
	key := strings.TrimRightFunc(kvLine[:firstSpace], unicode.IsSpace)
	value := strings.TrimLeftFunc(kvLine[firstSpace:], unicode.IsSpace)
	if len(key) == 0 {
		return nil, errInvalidKeyValueEmpty
	}
	// Key *must* contain a : at the end
	if !strings.HasSuffix(key, ":") {
		return nil, errInvalidKeyNoColon
	}
	// Trim ":" from the key
	key = key[:len(key)-1]
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

// KeyValueDecoder is used by Decoder to help it configure how to decode key/value pairs of a benchmark result
type KeyValueDecoder struct {
}

type Encoder struct {
}

func (e *Encoder) Encode(w io.Writer, run *Run) error {
	for _, kv := range run.Configuration.keys {
		if _, err := fmt.Fprintf(w, "%s\n", kv.String()); err != nil {
			return err
		}
	}
	for _, r := range run.Results {
		if _, err := fmt.Fprintf(w, "%s\n", r.String()); err != nil {
			return err
		}
	}
	return nil
}