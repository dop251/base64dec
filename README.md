# Universal Lenient Base64 Decoder

This is a modified version of the standard base64 decoder from `encoding/base64`. Key differences:

- Accepts both Standard and URL-safe encoding, no need to specify it explicitly.
- Works for both padded and raw inputs.
- The input can be a `string` or a `[]byte`, no allocation or copy is performed for either.

The goal was to create a decoder compatible with the one used in nodejs (and possibly elsewhere too). Unfortunately
the standard package does not expose enough API to do this efficiently, and trying all 4 possible variants
(standard/url, padded/raw) seemed wasteful.

Because there is no need to convert the input from `string` to `[]byte`, decoding strings (specially long ones)
is more efficient than with the standard library. Below is the comparison between `BenchmarkDecodeBase64Std`
and `BenchmarkDecodeBase64` on go1.21.3:

```
goos: linux
goarch: amd64
pkg: base64dec
cpu: Intel(R) Core(TM) i7-2600S CPU @ 2.80GHz
                    │    old.txt    │               new.txt               │
                    │    sec/op     │   sec/op     vs base                │
DecodeBase64/2-8       28.85n ±  3%   22.95n ± 2%  -20.45% (p=0.000 n=10)
DecodeBase64/4-8       48.05n ±  0%   41.41n ± 1%  -13.83% (p=0.000 n=10)
DecodeBase64/8-8       38.31n ±  2%   29.55n ± 1%  -22.89% (p=0.000 n=10)
DecodeBase64/64-8      406.7n ±  5%   113.0n ± 2%  -72.20% (p=0.000 n=10)
DecodeBase64/8192-8   31.670µ ± 10%   9.995µ ± 2%  -68.44% (p=0.000 n=10)
geomean                232.8n         126.0n       -45.89%

                    │    old.txt    │                new.txt                 │
                    │      B/s      │      B/s       vs base                 │
DecodeBase64/2-8      132.2Mi ±  3%    166.2Mi ± 2%   +25.70% (p=0.000 n=10)
DecodeBase64/4-8      158.8Mi ±  0%    184.2Mi ± 1%   +16.04% (p=0.000 n=10)
DecodeBase64/8-8      298.7Mi ±  2%    387.3Mi ± 1%   +29.67% (p=0.000 n=10)
DecodeBase64/64-8     206.4Mi ±  5%    742.4Mi ± 2%  +259.71% (p=0.000 n=10)
DecodeBase64/8192-8   329.0Mi ± 11%   1042.4Mi ± 2%  +216.87% (p=0.000 n=10)
geomean               211.8Mi          391.4Mi        +84.81%

                    │    old.txt     │                 new.txt                  │
                    │      B/op      │    B/op      vs base                     │
DecodeBase64/2-8        0.000 ± 0%      0.000 ± 0%         ~ (p=1.000 n=10) ¹
DecodeBase64/4-8        0.000 ± 0%      0.000 ± 0%         ~ (p=1.000 n=10) ¹
DecodeBase64/8-8        0.000 ± 0%      0.000 ± 0%         ~ (p=1.000 n=10) ¹
DecodeBase64/64-8       96.00 ± 0%       0.00 ± 0%  -100.00% (p=0.000 n=10)
DecodeBase64/8192-8   12.00Ki ± 0%     0.00Ki ± 0%  -100.00% (p=0.000 n=10)
geomean                            ²                ?                       ² ³
¹ all samples are equal
² summaries must be >0 to compute geomean
³ ratios must be >0 to compute geomean

                    │   old.txt    │                 new.txt                 │
                    │  allocs/op   │ allocs/op   vs base                     │
DecodeBase64/2-8      0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
DecodeBase64/4-8      0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
DecodeBase64/8-8      0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
DecodeBase64/64-8     1.000 ± 0%     0.000 ± 0%  -100.00% (p=0.000 n=10)
DecodeBase64/8192-8   1.000 ± 0%     0.000 ± 0%  -100.00% (p=0.000 n=10)
geomean                          ²               ?                       ² ³
¹ all samples are equal
² summaries must be >0 to compute geomean
³ ratios must be >0 to compute geomean
```