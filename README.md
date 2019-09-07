# benchparse
[![CircleCI](https://circleci.com/gh/cep21/benchparse.svg)](https://circleci.com/gh/cep21/benchparse)
[![GoDoc](https://godoc.org/github.com/cep21/benchparse?status.svg)](https://godoc.org/github.com/cep21/benchparse)
[![codecov](https://codecov.io/gh/cep21/benchparse/branch/master/graph/badge.svg)](https://codecov.io/gh/cep21/benchparse)

benchparse understands Go's benchmark format and parses it into an easy to read structure.  The entire spec
is defined at https://github.com/golang/proposal/blob/master/design/14313-benchmark-format.md

# Usage

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

# Design Rational

Follows Encode/Encoder/Decode/Decoder pattern of json library.  Tries to follow spec strictly since benchmark results
can also have extra output.

# Contributing

Contributions welcome!  Submit a pull request on github and make sure your code passes `make lint test`.  For
large changes, I strongly recommend [creating an issue](https://github.com/cep21/benchparse/issues) on GitHub first to
confirm your change will be accepted before writing a lot of code.  GitHub issues are also recommended, at your discretion,
for smaller changes or questions.

# License

This library is licensed under the Apache 2.0 License.