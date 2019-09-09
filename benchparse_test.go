package benchparse

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecoder_Decode(t *testing.T) {
	t.Run("readme", func(t *testing.T) {
		d := Decoder{}
		run, err := d.Decode(strings.NewReader(readmeExample))
		require.NoError(t, err)
		require.Len(t, run.Results, 27)
		require.Len(t, run.Results[0].Configuration.Order, 9)
	})
	t.Run("noise", func(t *testing.T) {
		d := Decoder{}
		run, err := d.Decode(strings.NewReader(noisyExample))
		require.NoError(t, err)
		require.Len(t, run.Results, 2)
		require.Len(t, run.Results[0].Configuration.Order, 1)
		require.Len(t, run.Results[1].Configuration.Order, 1)
	})
}

func TestDecoder_Stream(t *testing.T) {
	ctx, can := context.WithCancel(context.Background())
	can()
	d := Decoder{}
	i := 0
	err := d.Stream(ctx, strings.NewReader(readmeExample), func(_ BenchmarkResult) {
		i++
	})
	require.Error(t, err)
	require.Equal(t, 0, i)
}

func TestBenchmarkResultDecoder_decodeok(t *testing.T) {
	verifyParses := func(line string, expected string) func(t *testing.T) {
		return func(t *testing.T) {
			d := benchmarkResultDecoder{}
			res, err := d.decode(line)
			require.NoError(t, err)
			require.Equal(t, expected, res.String())
		}
	}
	t.Run("case=simple", verifyParses("Benchmark 1 10 ns/op", "Benchmark 1 10 ns/op"))
	t.Run("case=named", verifyParses("BenchmarkBob 1 10 ns/op", "BenchmarkBob 1 10 ns/op"))
	t.Run("case=manyspaces", verifyParses("BenchmarkBob       1\t10\t  \t  ns/op", "BenchmarkBob 1 10 ns/op"))
	t.Run("case=tworesults", verifyParses("Benchmark 1 10 ns/op 5 MB/s", "Benchmark 1 10 ns/op 5 MB/s"))
}

func TestBenchmarkResultDecoder_decodebad(t *testing.T) {
	verifyFails := func(line string, expected error) func(t *testing.T) {
		return func(t *testing.T) {
			d := benchmarkResultDecoder{}
			_, err := d.decode(line)
			require.Error(t, err)
			require.Equal(t, expected, err)
		}
	}
	t.Run("case=tooshort", verifyFails("Benchmark 1 10", errNotEnoughFields))
	t.Run("case=badprefix", verifyFails("TestBob 1 10 ns/op", errNoPrefixBenchmark))
	t.Run("case=noupper", verifyFails("Benchmarkbob 1 10 ns/op", errUpperAfterBench))
	t.Run("case=oddfields", verifyFails("Benchmarkbob 1 10 ns/op 5 MB/s 10", errEvenFields))

	verifyFailsSomehow := func(line string, msg string) func(t *testing.T) {
		return func(t *testing.T) {
			d := benchmarkResultDecoder{}
			_, err := d.decode(line)
			require.Error(t, err)
			require.Equal(t, msg, err.Error())
		}
	}
	t.Run("case=no_int_iters", verifyFailsSomehow("BenchmarkBob 1.5 10 ns/op", "strconv.Atoi: parsing \"1.5\": invalid syntax"))
	t.Run("case=no_float_value", verifyFailsSomehow("BenchmarkBob 1 10b ns/op", "strconv.ParseFloat: parsing \"10b\": invalid syntax"))
}

func TestKeyValueDecoder_decodeok(t *testing.T) {
	verifyParses := func(kvLine string, key string, value string) func(t *testing.T) {
		return func(t *testing.T) {
			d := keyValueDecoder{}
			kv, err := d.decode(kvLine)
			require.NoError(t, err)
			require.Equal(t, kv.Key, key)
			require.Equal(t, kv.Value, value)
		}
	}
	t.Run("case=simple", verifyParses("akey: bob", "akey", "bob"))
	t.Run("case=emptyvalue", verifyParses("akey: ", "akey", ""))
	t.Run("case=spaces", verifyParses("akey:           bob", "akey", "bob"))
	t.Run("case=strangechars", verifyParses("a--sdfds@#$%$34,>,: bob", "a--sdfds@#$%$34,>,", "bob"))
}

func TestKeyValueDecoder_decodeerror(t *testing.T) {
	verifyFails := func(kvLine string, expectedErr error) func(t *testing.T) {
		return func(t *testing.T) {
			d := keyValueDecoder{}
			_, err := d.decode(kvLine)
			require.Error(t, err)
			require.Equal(t, expectedErr, err)
		}
	}
	t.Run("case=startspace", verifyFails(" akey: bob", errInvalidKeyValueLowercase))
	t.Run("case=empty", verifyFails("", errInvalidKeyNoColon))
	t.Run("case=upperstart", verifyFails("Akey: bob", errInvalidKeyValueLowercase))
	t.Run("case=emptykey", verifyFails(": bob", errInvalidKeyValueEmpty))
	t.Run("case=keywithspaces", verifyFails("a key: bob", errInvalidKeyValueSpaces))
	t.Run("case=keywithtab", verifyFails("a\tkey: bob", errInvalidKeyValueSpaces))
	t.Run("case=keywithnewline", verifyFails("a\nkey: bob", errInvalidKeyValueSpaces))
	t.Run("case=valuewithnewline", verifyFails("akey: bo\nb", errInvalidKeyValueReturn))
}

func TestBenchmarkResult_AllKeyValuePairs(t *testing.T) {
	verifyPairs := func(in *BenchmarkResult, expected *OrderedStringStringMap) func(t *testing.T) {
		return func(t *testing.T) {
			out := in.AllKeyValuePairs()
			require.Equal(t, expected, out)
		}
	}
	t.Run("case=simple", verifyPairs(&BenchmarkResult{
		Name: "BenchmarkBob",
	}, &OrderedStringStringMap{
		Contents: map[string]string{
			"BenchmarkBob": "",
		},
		Order: []string{
			"BenchmarkBob",
		},
	}))
	t.Run("case=simpleconfig", verifyPairs(&BenchmarkResult{
		Name: "BenchmarkBob/name=bob",
		Configuration: &OrderedStringStringMap{
			Contents: map[string]string{
				"name": "bob",
			},
			Order: []string{
				"name",
			},
		},
	}, &OrderedStringStringMap{
		Contents: map[string]string{
			"BenchmarkBob": "",
			"name":         "bob",
		},
		Order: []string{
			"BenchmarkBob",
			"name",
		},
	}))
	t.Run("case=sometags", verifyPairs(&BenchmarkResult{
		Name: "BenchmarkBob/name=bob",
	}, &OrderedStringStringMap{
		Contents: map[string]string{
			"BenchmarkBob": "",
			"name":         "bob",
		},
		Order: []string{
			"BenchmarkBob",
			"name",
		},
	}))
	t.Run("case=configmatches", verifyPairs(&BenchmarkResult{
		Name: "BenchmarkBob/name=bob",
		Configuration: &OrderedStringStringMap{
			Contents: map[string]string{
				"name": "john",
			},
			Order: []string{
				"name",
			},
		},
	}, &OrderedStringStringMap{
		Contents: map[string]string{
			"BenchmarkBob": "",
			"name":         "bob",
		},
		Order: []string{
			"BenchmarkBob",
			"name",
		},
	}))
	t.Run("case=withdashnum", verifyPairs(&BenchmarkResult{
		Name: "BenchmarkBob/name=bob-8",
	}, &OrderedStringStringMap{
		Contents: map[string]string{
			"BenchmarkBob": "",
			"name":         "bob",
		},
		Order: []string{
			"BenchmarkBob",
			"name",
		},
	}))
	t.Run("case=justdash", verifyPairs(&BenchmarkResult{
		Name: "BenchmarkBob/name=bob-",
	}, &OrderedStringStringMap{
		Contents: map[string]string{
			"BenchmarkBob": "",
			"name":         "bob-",
		},
		Order: []string{
			"BenchmarkBob",
			"name",
		},
	}))
	t.Run("case=dashmixed", verifyPairs(&BenchmarkResult{
		Name: "BenchmarkBob/name=bob-3n",
	}, &OrderedStringStringMap{
		Contents: map[string]string{
			"BenchmarkBob": "",
			"name":         "bob-3n",
		},
		Order: []string{
			"BenchmarkBob",
			"name",
		},
	}))
}

func TestEncoder_Encode_symetric(t *testing.T) {
	symetricEncode := func(s string) func(t *testing.T) {
		return func(t *testing.T) {
			d := Decoder{}
			run, err := d.Decode(strings.NewReader(s))
			require.NoError(t, err)
			e := Encoder{}
			var buf bytes.Buffer
			require.NoError(t, e.Encode(&buf, run))
			require.Equal(t, s, buf.String())
		}
	}
	t.Run("case=empty", symetricEncode(""))
	t.Run("case=readme", symetricEncode(`commit: 7cd9055
BenchmarkDecode/text=digits/level=speed/size=1e4-8 100 154125 ns/op 64.88 MB/s 40418 B/op 7 allocs/op
`))
	t.Run("case=nokeys", symetricEncode(`BenchmarkDecode/text=digits/level=speed/size=1e4-8 100 154125 ns/op 64.88 MB/s 40418 B/op 7 allocs/op
`))
	t.Run("case=changekeys", symetricEncode(`commit: 7cd9055
BenchmarkDecode/text=digits/level=speed/size=1e4-8 100 154125 ns/op 64.88 MB/s 40418 B/op 7 allocs/op
commit: 7cd9056
BenchmarkDecode/text=digits/level=speed/size=1e4-8 100 154125 ns/op 64.88 MB/s 40418 B/op 8 allocs/op
`))
}
