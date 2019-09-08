# benchparse
[![CircleCI](https://circleci.com/gh/cep21/benchparse.svg)](https://circleci.com/gh/cep21/benchparse)
[![GoDoc](https://godoc.org/github.com/cep21/benchparse?status.svg)](https://godoc.org/github.com/cep21/benchparse)
[![codecov](https://codecov.io/gh/cep21/benchparse/branch/master/graph/badge.svg)](https://codecov.io/gh/cep21/benchparse)

benchparse understands Go's benchmark format and parses it into an easy to read structure.  The entire spec
is defined at https://github.com/golang/proposal/blob/master/design/14313-benchmark-format.md.  There are a few subtle
parts of the spec that make it less trivial than I thought to parse and conform to correctly.

# Usage

## Decoding benchmarks
```go
    func ExampleDecoder_Decode() {
        d := benchparse.Decoder{}
        run, err := d.Decode(strings.NewReader(""))
        if err != nil {
            panic(err)
        }
        fmt.Println(run)
        // Output:
    }
```

## Encoding benchmarks

```go
func ExampleEncoder_Encode() {
	run := benchparse.Run{
		Results:[] benchparse.BenchmarkResult{
			{
				Name:          "BenchmarkBob",
				Iterations:    1,
				Values:        []benchparse.ValueUnitPair{
					{
						Value: 345,
						Unit: "ns/op",
					},
				},
			},
	}}
	e := benchparse.Encoder{}
	if err := e.Encode(os.Stdout, &run); err != nil {
		panic(err)
	}
	// Output: BenchmarkBob 1 345 ns/op
}
```

## Example with changing keys

```go
func ExampleChangingKeys() {
	d := benchparse.Decoder{}
	run, err := d.Decode(strings.NewReader(`
commit: 7cd9055
BenchmarkDecode/text=digits/level=speed/size=1e4-8   	     100	    154125 ns/op	  64.88 MB/s	   40418 B/op	       7 allocs/op
commit: ab322f4
BenchmarkDecode/text=digits/level=speed/size=1e4-8   	     100	    154125 ns/op	  64.88 MB/s	   40418 B/op	       8 allocs/op
`))
	if err != nil {
		panic(err)
	}
	fmt.Println("commit of first run", run.Results[0].Configuration.Contents["commit"])
	fmt.Println("commit of second run", run.Results[1].Configuration.Contents["commit"])
	// Output: commit of first run 7cd9055
	// commit of second run ab322f4
}
```

## Example with streaming data

```go
func ExampleDecoder_Stream() {
	d := benchparse.Decoder{}
	err := d.Stream(context.Background(), strings.NewReader(`
BenchmarkDecode   	     100	    154125 ns/op	  64.88 MB/s	   40418 B/op	       7 allocs/op
BenchmarkEncode   	     100	    154125 ns/op	  64.88 MB/s	   40418 B/op	       8 allocs/op
`), func(result benchparse.BenchmarkResult) {
		fmt.Println("I got a result named", result.Name)
	})
	if err != nil {
		panic(err)
	}
	// Output: I got a result named BenchmarkDecode
	// I got a result named BenchmarkEncode
}
```

## More complete example
```go
func ExampleRun() {
	d := benchparse.Decoder{}
	run, err := d.Decode(strings.NewReader(`commit: 7cd9055
BenchmarkDecode/text=digits/level=speed/size=1e4-8   	     100	    154125 ns/op	  64.88 MB/s	   40418 B/op	       7 allocs/op
`))
	if err != nil {
		panic(err)
	}
	fmt.Println("The number of results:", len(run.Results))
	fmt.Println("Git commit:", run.Results[0].Configuration.Contents["commit"])
	fmt.Println("Base name of first result:", run.Results[0].BaseName())
	fmt.Println("Level config of first result:", run.Results[0].NameAsKeyValue().Contents["level"])
	testRunTime, _ := run.Results[0].ValueByUnit(benchparse.UnitRuntime)
	fmt.Println("Runtime of first result:", testRunTime)
	_, doesMissOpExists := run.Results[0].ValueByUnit("misses/op")
	fmt.Println("Does unit misses/op exist in the first run:", doesMissOpExists)
	// Output: The number of results: 1
	// Git commit: 7cd9055
	// Base name of first result: Decode/text=digits/level=speed/size=1e4-8
	// Level config of first result: speed
	// Runtime of first result: 154125
	// Does unit misses/op exist in the first run: false
}
```

# Design Rational

Follows Encode/Encoder/Decode/Decoder pattern of json library.  Tries to follow spec strictly since benchmark results
can also have extra output.  Naming is derived from the proposal's format document.  The benchmark will be decoded
into a structure like below

```go
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
	Unit  string
}
// OrderedStringStringMap is a map of strings to strings that maintains ordering.
// This statement implies uniqueness of keys per benchmark.
// "The interpretation of a key/value pair is up to tooling, but the key/value pair is considered to describe all benchmark results that follow, until overwritten by a configuration line with the same key."
type OrderedStringStringMap struct {
	// Contents are the values inside this map
	Contents map[string]string
	// Order is the string order of the contents of this map.  It is intended that len(Order) == len(Contents) and the
	// keys of Contents are all inside Order.
	Order []string
}
```

A few things about the structure of this spec stick out.

* Benchmarks often have the same configuration but do not have to.  It is possible for the benchmark output to change
the key/value configuration of a benchmark while running.
* Benchmark keys are stored in an ordered map to make encoding and decoding between benchmark outputs as symetric as
possible.
* The implementation had the option to use regex parsing, but since the spec is very clear about exact go functions
that should imply deliminators, I use those functions directly.
* There is no strict requirement that the benchmark output contain values for allocations or runtime.  There are unit
helpers to look these up.  For example:

```go

func ExampleBenchmarkResult_ValueByUnit() {
	res := &benchparse.BenchmarkResult{
		Values: []benchparse.ValueUnitPair {
			{
				Value: 125,
				Unit: "ns/op",
			},
		},
	}
	fmt.Println(res.ValueByUnit(benchparse.UnitRuntime))
	// Output: 125 true
}
```

# Similar tools

There is a similar tool at https://godoc.org/golang.org/x/tools/benchmark/parse which also parses benchmark output, but
does so in a very limited way and not to the flexibility defined by the README spec.

# Contributing

Contributions welcome!  Submit a pull request on github and make sure your code passes `make lint test`.  For
large changes, I strongly recommend [creating an issue](https://github.com/cep21/benchparse/issues) on GitHub first to
confirm your change will be accepted before writing a lot of code.  GitHub issues are also recommended, at your discretion,
for smaller changes or questions.

# License

This library is licensed under the Apache 2.0 License.