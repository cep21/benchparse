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
	fmt.Println(run)
	// Output:
}
