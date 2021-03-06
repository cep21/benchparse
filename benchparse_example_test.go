package benchparse_test

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/cep21/benchparse"
)

func ExampleDecoder_Decode() {
	d := benchparse.Decoder{}
	run, err := d.Decode(strings.NewReader(""))
	if err != nil {
		panic(err)
	}
	fmt.Println(len(run.Results))
	// Output: 0
}

func ExampleEncoder_Encode() {
	run := benchparse.Run{
		Results: []benchparse.BenchmarkResult{
			{
				Name:       "BenchmarkBob",
				Iterations: 1,
				Values: []benchparse.ValueUnitPair{
					{
						Value: 345,
						Unit:  "ns/op",
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

func ExampleDecoder_Decode_complete() {
	d := benchparse.Decoder{}
	run, err := d.Decode(strings.NewReader(`commit: 7cd9055
BenchmarkDecode/text=digits/level=speed/size=1e4-8   	     100	    154125 ns/op	  64.88 MB/s	   40418 B/op	       7 allocs/op
`))
	if err != nil {
		panic(err)
	}
	fmt.Println("The number of results:", len(run.Results))
	fmt.Println("Git commit:", run.Results[0].Configuration.Contents["commit"])
	fmt.Println("Name of first benchmark:", run.Results[0].Name)
	fmt.Println("Level config of first result:", run.Results[0].NameAsKeyValue().Contents["level"])
	testRunTime, _ := run.Results[0].ValueByUnit(benchparse.UnitRuntime)
	fmt.Println("Runtime of first result:", testRunTime)
	_, doesMissOpExists := run.Results[0].ValueByUnit("misses/op")
	fmt.Println("Does unit misses/op exist in the first run:", doesMissOpExists)
	// Output: The number of results: 1
	// Git commit: 7cd9055
	// Name of first benchmark: BenchmarkDecode/text=digits/level=speed/size=1e4-8
	// Level config of first result: speed
	// Runtime of first result: 154125
	// Does unit misses/op exist in the first run: false
}

func ExampleBenchmarkResult_NameAsKeyValue() {
	b := benchparse.BenchmarkResult{
		Name: "BenchmarkDecode/text=digits/level=speed/size=1e4-8",
	}
	fmt.Println(b.NameAsKeyValue().Contents["text"])
	// Output: digits
}

func ExampleBenchmarkResult_AllKeyValuePairs() {
	b := benchparse.BenchmarkResult{
		Configuration: &benchparse.OrderedStringStringMap{
			Contents: map[string]string{"commit": "a3abd32"},
			Order:    []string{"commit"},
		},
		Name: "BenchmarkDecode/text=digits/level=speed/size=1e4-8",
	}
	fmt.Println(b.AllKeyValuePairs().Contents["size"])
	fmt.Println(b.AllKeyValuePairs().Contents["commit"])
	// Output: 1e4
	// a3abd32
}

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

func ExampleDecoder_Decode_changingkeys() {
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

func ExampleOrderedStringStringMap() {
	d := benchparse.Decoder{}
	run, err := d.Decode(strings.NewReader(`
commit: 7cd9055
justthekey:

BenchmarkDecode/text=digits/level=speed/size=1e4-8   	     100	    154125 ns/op	  64.88 MB/s	   40418 B/op	       7 allocs/op
`))
	if err != nil {
		panic(err)
	}
	fmt.Println(run.Results[0].Configuration.Contents["commit"])
	fmt.Println(run.Results[0].Configuration.Contents["justthekey"])
	fmt.Println(run.Results[0].Configuration.Contents["does not exist"])
	// Output: 7cd9055
	//
	//
}

func ExampleBenchmarkResult_ValueByUnit() {
	res := &benchparse.BenchmarkResult{
		Values: []benchparse.ValueUnitPair{
			{
				Value: 125,
				Unit:  "ns/op",
			},
		},
	}
	fmt.Println(res.ValueByUnit(benchparse.UnitRuntime))
	// Output: 125 true
}
