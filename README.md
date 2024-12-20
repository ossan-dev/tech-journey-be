# Coworking App

This project will sustain my upcoming speech about Profiling & Tracing. It emulates whatever needs to be taken into consideration when aiming to improve the performance of a project regarding memory and CPU point of views. It's like a TODO list.

## Steps

### 1. Discovery Overview

Inspect the source code to get familiar with it. Understand its dependencies, look at config files, potential bottlenecks, and so on.

### 2. Compile-Time checks

First and foremost, you should empower the Go compiler about memory/CPU optimizations it takes based on your source code. By running it, you can start addressing things that are relevant to the area of code that needs to further being investigated.

First command: `go build -gcflags=-m &>> compiler.txt`. Further ref: <https://askubuntu.com/a/420983/1546072>
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
