HashLand
========
Hashland is a collection of hash functions and functionality to test them.

Introduction
------------
HashLand contains the following functionality.

1. A (currently barely) curated set of pure Go hash functions including various jenkins hash functions and his latest SpookyV2, Murmur3, Cityhash, sbox, MaHash8v64, CrapWow, Siphash, keccak, skein and more.

2. AES based hash functions extracted Go's runtime and used by the map implementation on Intel X86 architecture machines that support AES-NI.

3. Tests which (mostly) use file based dictionaries to gather statistics about the above hash functions.

4. An extraction with a little generalization of the SMHasher functions from the Go runtime. These functions were already ported from the SMHasher C code by the Go Authors.

5. The ability to benchmark hash functions.

6. A package, "nhash" with a new set of Go interfaces for hash functions that complement the existing core Go streaming hash interface. The core of the proposal is:

	`Hash32(b []byte, seeds ...uint32) uint32`  
	`Hash64(b []byte, seeds ...uint64) uint64`  
	`Hash128(b []byte, seeds ...uint64) (uint64, uint64)`  
	`Hash(in []byte, out []byte, seeds ...uint64) []byte)`

My experiment using a variadic argument as an optional argument for a seed is deemed a failure for two reasons. Performance of variadic arguments causes an allocation on each invocation. I knew that going in and I should have known better. Go style eschews the use of a variable argument as an optional argument, although I am not sure I agree 100%. I am considering changing some or all of the variadic arguments to `…byte` but doing so will make calling these functions less convenient.

Background
----------
In the process of writing a [cuckoo hash table](https://github.com/tildeleb/cuckoo) I wondered which hash functions would be the best ones to use. I was frustrated that the core Go libraries didn't contain a useful set of hash functions for constructing hash tables. Also, I had always wanted to experiment with hash functions. So I spend some time building HashLand to figure it all out. I wrote a bunch of hash functions in pure Go. I spent some time with the gc inliner seeing how much optimization could be done in the context of Go. I forked some hash functions from other repositories, and tested them out. I ended up putting in a bit more effort than I had planned

Quality of Hash Functions
-------------------------
There are no "bad" hash functions here. Most of these do a very good job of hashing keys with good distribution and few duplicate hashes. However, my tests and dictionaries are still very basic. I've been focused on getting the hash functions written and some testing infrastructure up and running. blah blah blah.

Performance
-----------
Some of these are woefully lacking in performance. Many will be difficult to improve with the gc based compilers in pure Go. If you want speed and need crypto quality use ...

Warning
-------
*Don't use the non crypto hash functions if you have uncontrolled inputs (i.e. a web based API or web facing data inputs or an adversary. If you do, use at least SipHash or one of the other crypto hash functions.*

Roadmap
----------
1. A few more hash functions
2. Make sure licensing and author information is accurate
3. Performance optimization
4. A few more tests
5. Better dictionaries
6. Better stats

Non Crypto Hash Functions
-------------------------
	"sbox":			simple hash function         
	"CrapWow":		another simple hash function
	"MaHash8v64":	russian hash function
	"j332c":		Jenkins lookup3 c bits hash
	"j332b":		Jenkins lookup3 b bits hash
	"j232":			Jenkins lookup8 32 bit
	"j264l": 		Jenkins lookup8 64 bit (low bits)
	"j264h": 		Jenkins lookup8 64 bit (high bits)
	"j264xor":		Jenkins lookup8 64 bit (high xor low bits)
	"spooky32":	Jenkins Spooky, 32 bit
	"spooky64":	Jenkins Spooky, 64 bit
	"spooky128h":	Jenkins Spooky, 128 bit, high half
	"spooky128l:	Jenkins Spooky, 128 bit, low half
	"spooky128xor:	Jenkins Spooky, 128 bit, low xor half

Crypto Hash Functions
---------------------
	"aeshash64"
	"siphashal": 
	"siphashah": 
	"siphashbl": 
	"siphashbh": 
	"skein256xor": 
	"skein256low": 
	"skein256hi": 
	"sha1160": 
	"keccak160l"

Usage
-----
	Usage of ./hashland:
	./hashland: [flags] [dictionary-files]
	  -A=false: test A
	  -B=false: test B
	  -C=false: test C
	  -D=false: test D
	  -E=false: test E
	  -F=false: test F
	  -G=false: test G
	  -H=false: test H
	  -I=false: test I
	  -J=false: test J
	  -a=false: run all tests
	  -b=false: run benchmarks
	  -c=false: only test crypto hash functions
	  -cd=false: check for duplicate hashs when running benchmarks
	  -e=1: extra bis in table size
	  -file="": words to read
	  -h32=false: only test 32 bit has functions
	  -h64=false: only test 64 bit has functions
	  -hf="all": hash function
	  -n=100000000: number of hashes for benchmark
	  -ni=200000: number of integer keys
	  -oa=false: open addressing (no buckets)
	  -p=false: table size is primes and use mod
	  -pd=false: print duplicate hashes
	  -sm=false: run SMHasher
	  -v=false: verbose

SMHasher
--------
Still some work to do on this, particularly with `-v`.  

	leb@hula: % hashland -sm -hf=j264 -v
	"TestSmhasherSanity": 118.968289ms
	"TestSmhasherSeed": 30.632449ms
	"TestSmhasherText": 5.923103295s
	"TestSmhasherWindowed": 58.097608577s
	"TestSmhasherAvalanche": 		z=100000, n=16
		z=100000, n=32
		z=100000, n=64
		z=100000, n=128
		z=100000, n=256
		z=100000, n=1600
		z=100000, n=32
		z=100000, n=64
	1m6.748766357s
	"TestSmhasherPermutation": 
		n=8, s=[0 1 2 3 4 5 6 7]
		n=8, s=[0 536870912 1073741824 1610612736 2147483648 2684354560 3221225472 3758096384]
		n=20, s=[0 1]
		n=20, s=[0 2147483648]
		n=6, s=[0 1 2 3 4 5 6 7 536870912 1073741824 1610612736 2147483648 2684354560 3221225472 3758096384]20.717874625s
	"TestSmhasherSparse": 9.937053904s
	"TestSmhasherCyclic": 4.438443435s
	"TestSmhasherSmallKeys": 16.713253851s
	"TestSmhasherZeros": 6.013460606s
	"TestSmhasherAppendedZeros": 126.096µs

Benchmarks (currently broken)
-----------------------------
	leb@hula:~/gotest/src/github.com/tildeleb/hashland % hashland -b -hf=j364 -v
	
	ksiz=4, len(bs)=4
	benchmark32g: gen n=100000000, n=100 M, keySize=4,  size=400 MB
	benchmark32g: 38 Mhashes/sec
	benchmark32g: 151 MB/sec
	
	ksiz=8, len(bs)=8
	benchmark32g: gen n=100000000, n=100 M, keySize=8,  size=800 MB
	benchmark32g: 36 Mhashes/sec
	benchmark32g: 288 MB/sec
	
	ksiz=16, len(bs)=16
	benchmark32g: gen n=100000000, n=100 M, keySize=16,  size=1 GB
	benchmark32g: 22 Mhashes/sec
	benchmark32g: 345 MB/sec
	
	ksiz=32, len(bs)=32
	benchmark32g: gen n=100000000, n=100 M, keySize=32,  size=3 GB
	benchmark32g: 15 Mhashes/sec
	benchmark32g: 469 MB/sec
	
	ksiz=64, len(bs)=64
	benchmark32g: gen n=100000000, n=100 M, keySize=64,  size=6 GB
	benchmark32g: 8 Mhashes/sec
	benchmark32g: 508 MB/sec
	
	ksiz=512, len(bs)=512
	benchmark32g: gen n=10000000, n=10 M, keySize=512,  size=5 GB
	benchmark32g: 1 Mhashes/sec
	benchmark32g: 577 MB/sec
	
	ksiz=1024, len(bs)=1024
	benchmark32g: gen n=10000000, n=10 M, keySize=1024,  size=10 GB
	benchmark32g: 570 khashes/sec
	benchmark32g: 584 MB/sec
	
	leb@hula:~/gotest/src/github.com/tildeleb/hashland %

Tests
-----

	leb@hula:~/gotest/src/github.com/tildeleb/hashland % hashland -oa -A db/pagecounts-20140101-000000
	file="db/pagecounts-20140101-000000", lines=6460902
	Test0 - ReadFile
		          "ReadFile": size=16777216, inserts=0, cols=0, probes=0, cpi=0.00%, ppi=0NaN, dups=0, time=1.81s
	TestA - insert keys
		         "aeshash64": size=16777216, inserts=6460902, cols=1244858, probes=7238283, cpi=7.42%, ppi=1.12, dups=0, time=7.21s
		              "j364": size=16777216, inserts=6460902, cols=1243168, probes=7238324, cpi=7.41%, ppi=1.12, dups=0, time=8.33s
		              "j264": size=16777216, inserts=6460902, cols=1246293, probes=7239983, cpi=7.43%, ppi=1.12, dups=0, time=9.08s
		         "siphash64": size=16777216, inserts=6460902, cols=1244130, probes=7240076, cpi=7.42%, ppi=1.12, dups=0, time=8.61s
		        "MaHash8v64": size=16777216, inserts=6460902, cols=1244055, probes=7237949, cpi=7.42%, ppi=1.12, dups=0, time=10.64s
		          "spooky64": size=16777216, inserts=6460902, cols=1245230, probes=7240621, cpi=7.42%, ppi=1.12, dups=0, time=9.10s
		        "spooky128h": size=16777216, inserts=6460902, cols=1245230, probes=7240621, cpi=7.42%, ppi=1.12, dups=0, time=9.31s
		        "spooky128l": size=16777216, inserts=6460902, cols=1244426, probes=7241584, cpi=7.42%, ppi=1.12, dups=0, time=10.03s
		      "spooky128xor": size=16777216, inserts=6460902, cols=1243033, probes=7240442, cpi=7.41%, ppi=1.12, dups=0, time=9.47s
		             "j332c": size=16777216, inserts=6460902, cols=1243168, probes=7238324, cpi=7.41%, ppi=1.12, dups=4854, time=9.19s
		             "j332b": size=16777216, inserts=6460902, cols=1243329, probes=7242425, cpi=7.41%, ppi=1.12, dups=4846, time=9.17s
		              "j232": size=16777216, inserts=6460902, cols=1245159, probes=7242283, cpi=7.42%, ppi=1.12, dups=4822, time=8.51s
		             "j264l": size=16777216, inserts=6460902, cols=1246293, probes=7239983, cpi=7.43%, ppi=1.12, dups=4897, time=8.58s
		             "j264h": size=16777216, inserts=6460902, cols=1243960, probes=7238234, cpi=7.41%, ppi=1.12, dups=4935, time=8.78s
		           "j264xor": size=16777216, inserts=6460902, cols=1243885, probes=7241105, cpi=7.41%, ppi=1.12, dups=4719, time=9.03s
		          "spooky32": size=16777216, inserts=6460902, cols=1245230, probes=7240621, cpi=7.42%, ppi=1.12, dups=4880, time=9.46s
		              "sbox": size=16777216, inserts=6460902, cols=1244814, probes=7240906, cpi=7.42%, ppi=1.12, dups=8695, time=9.46s
		              "sha1": size=16777216, inserts=6460902, cols=1243120, probes=7237478, cpi=7.41%, ppi=1.12, dups=0, time=13.35s
		         "keccak643": size=16777216, inserts=6460902, cols=1244489, probes=7241060, cpi=7.42%, ppi=1.12, dups=0, time=28.45s
		          "skein256": size=16777216, inserts=6460902, cols=1243784, probes=7239653, cpi=7.41%, ppi=1.12, dups=0, time=28.32s
	leb@hula:~/gotest/src/github.com/tildeleb/hashland % hashland -A db/pagecounts-20140101-000000
	file="db/pagecounts-20140101-000000", lines=6460902
	Test0 - ReadFile
		          "ReadFile": size=16777216, inserts=0, buckets=0, dups=0, q=0.00, time=1.99s
	TestA - insert keys
		         "aeshash64": size=16777216, inserts=6460902, buckets=5362095, dups=0, q=1.00, time=7.67s
		              "j364": size=16777216, inserts=6460902, buckets=5362763, dups=0, q=1.00, time=8.53s
		              "j264": size=16777216, inserts=6460902, buckets=5360151, dups=0, q=1.00, time=8.26s
		         "siphash64": size=16777216, inserts=6460902, buckets=5362116, dups=0, q=1.00, time=7.88s
		        "MaHash8v64": size=16777216, inserts=6460902, buckets=5362919, dups=0, q=1.00, time=9.95s
		          "spooky64": size=16777216, inserts=6460902, buckets=5361474, dups=0, q=1.00, time=8.79s
		        "spooky128h": size=16777216, inserts=6460902, buckets=5361474, dups=0, q=1.00, time=9.05s
		        "spooky128l": size=16777216, inserts=6460902, buckets=5361895, dups=0, q=1.00, time=8.96s
		      "spooky128xor": size=16777216, inserts=6460902, buckets=5362479, dups=0, q=1.00, time=8.63s
		             "j332c": size=16777216, inserts=6460902, buckets=5362763, dups=4854, q=1.00, time=8.29s
		             "j332b": size=16777216, inserts=6460902, buckets=5363075, dups=4846, q=1.00, time=8.48s
		              "j232": size=16777216, inserts=6460902, buckets=5361330, dups=4822, q=1.00, time=8.46s
		             "j264l": size=16777216, inserts=6460902, buckets=5360151, dups=4897, q=1.00, time=8.05s
		             "j264h": size=16777216, inserts=6460902, buckets=5362206, dups=4935, q=1.00, time=8.22s
		           "j264xor": size=16777216, inserts=6460902, buckets=5362622, dups=4719, q=1.00, time=8.04s
		          "spooky32": size=16777216, inserts=6460902, buckets=5361474, dups=4880, q=1.00, time=9.10s
		              "sbox": size=16777216, inserts=6460902, buckets=5361809, dups=8695, q=1.00, time=8.68s
		              "sha1": size=16777216, inserts=6460902, buckets=5362758, dups=0, q=1.00, time=12.48s
		         "keccak643": size=16777216, inserts=6460902, buckets=5362212, dups=0, q=1.00, time=25.58s
		          "skein256": size=16777216, inserts=6460902, buckets=5362896, dups=0, q=1.00, time=28.76s
	leb@hula:~/gotest/src/github.com/tildeleb/hashland % 


Note
----
**Some licensing information may be missing; will be rectified soon**

