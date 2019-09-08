package benchparse_test

import (
	"fmt"
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
