# Coworking App

This project will sustain my upcoming speech about Profiling & Tracing. It emulates whatever needs to be taken into consideration when aiming to improve the performance of a project regarding memory and CPU point of views. It's like a TODO list.

## Steps

### 1. Discovery Overview

Inspect the source code to get familiar with it. Understand its dependencies, look at config files, potential bottlenecks, and so on.

### 2. Compile-Time Checks

First and foremost, you should empower the Go compiler about memory/CPU optimizations it takes based on your source code. By running it, you can start addressing things that are relevant to the area of code that needs to further being investigated.

First command: `go build -gcflags=-m &>> compiler.txt`. Further ref: <https://askubuntu.com/a/420983/1546072>. The `&>>` operator creates the specified file if not exist or it appends to it.
Second command: `go build -gcflags="-m -m" &>> compiler.txt`
Third command: `go build -gcflags="-m -m -l" &>> compiler.txt`

The first optimization is:

1. **escapes to heap** happening in several parts. It was happening due to some root causes:
    1. Usage of interfaces. If we call a function that accepts `any` which is the `interface{}` there isn't much we can do (e.g. `c.Set()`, `c.JSON()`)
    2. Creating a fire & forget pointer variable inside a function
    3. Use a slice whenever you know in advance the how many elements do you need (could be switched into an array)
    4. `func literal escapes to heap`: can be fixed by prepending the comment `go:noinline`

### 3. Measure Performance

Now, you've optmized something and something not. The un-optimized things are left as they were. We need to make sure that they doesn't impact performance. Let's use th e `runtime` package.
Basically, the only thing we're left with is to wrap the **unoptimized** call within the function invocation `PrintMemStats` which prints some memory information.  
By doing that, we can make sure our calls doesn't affect too much the performance. For reference, you can have a look at the file `utils/mem_usage.go`. The fields of the `MemStats` struct we care about are:

1. `Alloc`
2. `TotalAlloc`
3. `Sys`
4. `NumGC`

In our code we saw this impact:

```text
Before:

Alloc = 1 MB
TotalAlloc = 1 MB
Sys = 7 MB
NumGC = 0

After:

Alloc = 2 MB
TotalAlloc = 2 MB
Sys = 7 MB
NumGC = 0
```

We can overlook this un-optimization since it's not worthwhile. Plus, this code will be run once at the program startup.

### 4. Using Benchmarks

Useful when you've two implementations of the same feature (e.g. two ways to parse JSON files and have the slices with the relevant data).
The source code (which has two solvers function) is within the file `models/models.go`. Instead the benchmark is contained in the `models/models_test.go`.
We used the sub-benchmark technique to keep it simpler and more readable.
To execute it, `cd models/` and then `go test -bench=.`.

```text
goos: linux
goarch: amd64
pkg: github.com/ossan-dev/coworkingapp/models
cpu: 13th Gen Intel(R) Core(TM) i7-1355U
BenchmarkParsingJsonFile/ParseModelWithUnmarshal-1-rooms-12               154105              6793 ns/op            1184 B/op         20 allocs/op
BenchmarkParsingJsonFile/ParseModelWithDecoder-1-rooms-12                 189138              6677 ns/op            1408 B/op         23 allocs/op
BenchmarkParsingJsonFile/ParseModelWithUnmarshal-999-rooms-12                152           7368017 ns/op         2196225 B/op       8030 allocs/op
BenchmarkParsingJsonFile/ParseModelWithDecoder-999-rooms-12                  217           5111452 ns/op         1261293 B/op       8018 allocs/op
BenchmarkParsingJsonFile/ParseModelWithUnmarshal-9999-rooms-12                33          35786404 ns/op        28643201 B/op      80043 allocs/op
BenchmarkParsingJsonFile/ParseModelWithDecoder-9999-rooms-12                  33          36373821 ns/op        18910856 B/op      80023 allocs/op
PASS
ok      github.com/ossan-dev/coworkingapp/models        9.407s
```
