package benchparse

import (
	"bufio"
	"errors"
	"io"
	"strconv"
	"strings"
	"unicode"
)

type Run struct {
	Configuration KeyValueList
	Results       []BenchmarkResult
}

type KeyValueDecoder struct {
}

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

type Decoder struct {
	KeyValueDecoder        KeyValueDecoder
	BenchmarkResultDecoder BenchmarkResultDecoder
}

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

type KeyValueList struct {
	keys []KeyValue
}

func (c KeyValueList) AsMap() map[string]string {
	ret := make(map[string]string)
	for _, k := range c.keys {
		if _, exists := ret[k.Key]; !exists {
			ret[k.Key] = k.Value
		}
	}
	return ret
}

func (c KeyValueList) LookupAll(key string) []string {
	ret := make([]string, 0, 1)
	for _, k := range c.keys {
		if k.Key == key {
			ret = append(ret, k.Value)
		}
	}
	return ret
}

func (c KeyValueList) Lookup(key string) (string, bool) {
	for _, k := range c.keys {
		if k.Key == key {
			return k.Value, true
		}
	}
	return "", false
}

type KeyValue struct {
	Key   string
	Value string
}

func (k KeyValue) String() string {
	if k.Value == "" {
		return k.Key + ":"
	}
	return k.Key + ": " + k.Value
}

type BenchmarkValue struct {
	Value float64
	Unit  string
}

func (b BenchmarkValue) String() string {
	return strconv.FormatFloat(b.Value, 'f', -1, 64) + " " + b.Unit
}

type BenchmarkResult struct {
	Name       string
	Iterations int
	Values     []BenchmarkValue
}

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

func (b BenchmarkResult) BaseName() string {
	return strings.TrimPrefix(b.Name, "Benchmark")
}

func (b BenchmarkResult) String() string {
	ret := make([]string, 0, len(b.Values)+2)
	ret = append(ret, b.Name, strconv.Itoa(b.Iterations))
	for _, v := range b.Values {
		ret = append(ret, v.String())
	}
	return strings.Join(ret, " ")
}
