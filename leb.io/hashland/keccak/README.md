# TL;DR
## 10x faster than https://github.com/ebfe/keccak

### Usage

```go
import "github.com/steakknife/keccak"

...

h := keccak.New256()
h.Write([]byte("Awesome"))
keccakhash := h.Sum(nil)
```

### keccak cgo 

    PASS
    BenchmarkKeccak224Write1MiB      100      11056848 ns/op
    BenchmarkKeccak224Sum         500000          3870 ns/op
    BenchmarkKeccak256Write1MiB      100      10845596 ns/op
    BenchmarkKeccak256Sum         500000          4010 ns/op
    BenchmarkKeccak384Write1MiB      100      15315352 ns/op
    BenchmarkKeccak384Sum         500000          4004 ns/op
    BenchmarkKeccak512Write1MiB      100      21881999 ns/op
    BenchmarkKeccak512Sum         500000          3807 ns/op
    ok

### https://github.com/ebfe/keccak 

    PASS
    BenchmarkKeccak224Write1MiB       20     116710461 ns/op
    BenchmarkKeccak224Sum         100000         17185 ns/op
    BenchmarkKeccak256Write1MiB       20     117566765 ns/op
    BenchmarkKeccak256Sum         100000         17011 ns/op
    BenchmarkKeccak384Write1MiB       10     160507004 ns/op
    BenchmarkKeccak384Sum         100000         17076 ns/op
    BenchmarkKeccak512Write1MiB       10     229374464 ns/op
    BenchmarkKeccak512Sum         100000         17011 ns/op
    ok

### optimized 64-bit implementation borrowed from http://csrc.nist.gov/groups/ST/hash/sha-3/Round3/documents/Keccak_FinalRnd.zip
### tests borrowed from https://github.com/ebfe/keccak
