package benchparse

import (
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
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
		require.Len(t, run.Configuration.keys, 9)
	})
}
