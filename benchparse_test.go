package benchparse

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const readmeExample = `commit: 7cd9055
commit-time: 2016-02-11T13:25:45-0500
goos: darwin
goarch: amd64
cpu: Intel(R) Core(TM) i7-4980HQ CPU @ 2.80GHz
cpu-count: 8
cpu-physical-count: 4
os: Mac OS X 10.11.3
mem: 16 GB

BenchmarkDecode/text=digits/level=speed/size=1e4-8   	     100	    154125 ns/op	  64.88 MB/s	   40418 B/op	       7 allocs/op
BenchmarkDecode/text=digits/level=speed/size=1e5-8   	      10	   1367632 ns/op	  73.12 MB/s	   41356 B/op	      14 allocs/op
BenchmarkDecode/text=digits/level=speed/size=1e6-8   	       1	  13879794 ns/op	  72.05 MB/s	   52056 B/op	      94 allocs/op
BenchmarkDecode/text=digits/level=default/size=1e4-8 	     100	    147551 ns/op	  67.77 MB/s	   40418 B/op	       8 allocs/op
BenchmarkDecode/text=digits/level=default/size=1e5-8 	      10	   1197672 ns/op	  83.50 MB/s	   41508 B/op	      13 allocs/op
BenchmarkDecode/text=digits/level=default/size=1e6-8 	       1	  11808775 ns/op	  84.68 MB/s	   53800 B/op	      80 allocs/op
BenchmarkDecode/text=digits/level=best/size=1e4-8    	     100	    143348 ns/op	  69.76 MB/s	   40417 B/op	       8 allocs/op
BenchmarkDecode/text=digits/level=best/size=1e5-8    	      10	   1185527 ns/op	  84.35 MB/s	   41508 B/op	      13 allocs/op
BenchmarkDecode/text=digits/level=best/size=1e6-8    	       1	  11740304 ns/op	  85.18 MB/s	   53800 B/op	      80 allocs/op
BenchmarkDecode/text=twain/level=speed/size=1e4-8    	     100	    143665 ns/op	  69.61 MB/s	   40849 B/op	      15 allocs/op
BenchmarkDecode/text=twain/level=speed/size=1e5-8    	      10	   1390359 ns/op	  71.92 MB/s	   45700 B/op	      31 allocs/op
BenchmarkDecode/text=twain/level=speed/size=1e6-8    	       1	  12128469 ns/op	  82.45 MB/s	   89336 B/op	     221 allocs/op
BenchmarkDecode/text=twain/level=default/size=1e4-8  	     100	    141916 ns/op	  70.46 MB/s	   40849 B/op	      15 allocs/op
BenchmarkDecode/text=twain/level=default/size=1e5-8  	      10	   1076669 ns/op	  92.88 MB/s	   43820 B/op	      28 allocs/op
BenchmarkDecode/text=twain/level=default/size=1e6-8  	       1	  10106485 ns/op	  98.95 MB/s	   71096 B/op	     172 allocs/op
BenchmarkDecode/text=twain/level=best/size=1e4-8     	     100	    138516 ns/op	  72.19 MB/s	   40849 B/op	      15 allocs/op
BenchmarkDecode/text=twain/level=best/size=1e5-8     	      10	   1227964 ns/op	  81.44 MB/s	   43316 B/op	      25 allocs/op
BenchmarkDecode/text=twain/level=best/size=1e6-8     	       1	  10040347 ns/op	  99.60 MB/s	   72120 B/op	     173 allocs/op
BenchmarkEncode/text=digits/level=speed/size=1e4-8   	      30	    482808 ns/op	  20.71 MB/s
BenchmarkEncode/text=digits/level=speed/size=1e5-8   	       5	   2685455 ns/op	  37.24 MB/s
BenchmarkEncode/text=digits/level=speed/size=1e6-8   	       1	  24966055 ns/op	  40.05 MB/s
BenchmarkEncode/text=digits/level=default/size=1e4-8 	      20	    655592 ns/op	  15.25 MB/s
BenchmarkEncode/text=digits/level=default/size=1e5-8 	       1	  13000839 ns/op	   7.69 MB/s
BenchmarkEncode/text=digits/level=default/size=1e6-8 	       1	 136341747 ns/op	   7.33 MB/s
BenchmarkEncode/text=digits/level=best/size=1e4-8    	      20	    668083 ns/op	  14.97 MB/s
BenchmarkEncode/text=digits/level=best/size=1e5-8    	       1	  12301511 ns/op	   8.13 MB/s
BenchmarkEncode/text=digits/level=best/size=1e6-8    	       1	 137962041 ns/op	   7.25 MB/s`

func TestDecoder_Decode(t *testing.T) {
	t.Run("readme", func(t *testing.T) {
		d := Decoder{}
		run, err := d.Decode(strings.NewReader(readmeExample))
		require.NoError(t, err)
		require.Len(t, run.Results, 27)
		require.Len(t, run.Results[0].Configuration.Order, 9)
	})
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
	t.Run("case=empty", verifyFails("", errInvalidKeyNoColon))
	t.Run("case=upperstart", verifyFails("Akey: bob", errInvalidKeyValueLowercase))
	t.Run("case=emptykey", verifyFails(": bob", errInvalidKeyValueEmpty))
	t.Run("case=keywithspaces", verifyFails("a key: bob", errInvalidKeyValueSpaces))
	t.Run("case=keywithtab", verifyFails("a\tkey: bob", errInvalidKeyValueSpaces))
	t.Run("case=keywithnewline", verifyFails("a\nkey: bob", errInvalidKeyValueSpaces))
	t.Run("case=valuewithnewline", verifyFails("akey: bo\nb", errInvalidKeyValueReturn))
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
