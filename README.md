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

### 5. memprofile, cpuprofile, pprof

> Starting from now, these metrics should be captured while the application is running. Best would be in a production environment where you can experiment the workload. If you're application is not running, you have to write a script that can stress out your app to understand how it behaves under heavy loads.

Your machine must have installed Graphiz. You can install it by running this command:

`sudo apt install graphviz`

To enable `pprof` we have to first discern between Web and non Web applications. For this project will only care about Web applications.  

Again, we've to make another distinction between Web servers using `DefaultServeMux` and Web servers not using it.
Since our application is written with the **Gin Web Framework**, we've to make some extra steps, as discussed in the following section.  

#### Web Servers Using `DefaultServeMux`

It's usually enough to add `_ "net/http/pprof"` in the import section.

#### Web Servers Not Using `DefaultServerMux`

This is the process for a web server which empowers the `Gin Web Framework`.  

First, you have to download a third-party pkg by using this command: `go get -u github.com/gin-contrib/pprof`
Then, you can register the `pprof` endpoints by referencing `pprof.Register(r)` in `main.go` file.

Now, it's time to run the application and let it generates a profile that can be inspected.

#### Collecting Profiles

The first profile you might want to collect is the `allocs` profile.  
This profile takes into consideration all the memory allocations (well, memory allocations at least of _512 bytes_) since the start of the program. Given the fact that it also takes into consideration past allocations, it doesn't have to be done against an active app (it might be idle in this moment and not doing anything).  

The command to download it is:

`curl -v <http://localhost:8080/debug/pprof/allocs> > allocs.out`

We are also concerned to two other profiles: `profile` and `heap`. Let's start with the `heap` which is still relevant to memory usage.  
  
The `heap` profile shows us the current memory allocations. We can use it two ways:

1. Running the command `curl -v <http://localhost:8080/debug/pprof/heap> > heap.out` multiple times. In this way, we take a snapshot of the situation at different periods. We might be able to highlight some trends of the memory by using the `pprof` tool.
2. We can run the command `curl -v <http://localhost:8080/debug/pprof/heap?seconds=10> > heap.out` to get a delta profile after the seconds specified have been elapsed. By using `pprof` we can see increases/decreases the memory consumption has had over the specified time period  
  
The `cpu` profile (which for backward-compatibility is still addressed to as `profile`) highlights the top CPU consumers of our application. As per the `heap` profile we just seen, it should be collected against a live application to get the most out of it.  
To start CPU-profiling for _30 seconds_, run the command `curl -v <http://localhost:8080/debug/pprof/profile?seconds=30> > cpu.out`.  
> Please note this is not a delta profile. It will record all the milliseconds (the unit of measurement for CPU is `ms`) spent in the functions invoked.  
  
#### Visualizing and Understading Profiles Data

For the sake of the demo, we won't explain how to visualize/understand the heap profile. We'll focus only on `allocs` and `cpu` profiles.

##### Interactive Console

###### allocs.out

Let's start from the investigation of the `allocs.out` profile. Those are some steps you can take:

1. `go tool pprof allocs.out`
2. Notice the `Type` which is set to `alloc_space` which is _total amount of memory allocated (regardless of released)_. If you want to look at the _total amount of memory allocated and not released yet_, run `go tool pprof -inuse_space allocs.out`. If you care about the _number of objects_, instead of the bytes, simply replace the `space` with `objects` in the command
3. Next, type `top` which is the default for `top10`. Those are the top memory consumers.
4. If you see a line saying `Dropped <N> Nodes`, it means it's trying to reduce some noise and display fewer nodes. You can avoid this by running `-nodefraction=0` (when running the `pprof` tool) or by typing `nodefraction=0` in the interactive console
5. `go tool pprof -nodefraction=0 -list SetupRoutes allocs.out`. When we find out a potential culprit, we can run this to narrow things down.

    ```text
    Total: 1.17GB
    ROUTINE ======================== github.com/ossan-dev/coworkingapp/handlers.SetupRoutes.AuthorizeUser.func3 in /home/ossan/Projects/tech-journey-be/middlewares/auth.go
      512.02kB   918.56MB (flat, cum) 76.96% of Total
             .          .     14:   return func(c *gin.Context) {
             .          .     15:           tokenHeader := c.GetHeader("Authorization")
             .          .     16:           if tokenHeader == "" {
             .          .     17:                   c.JSON(http.StatusUnauthorized, models.CoworkingErr{Code: models.MissingTokenErr, Message: "please provide a jwt token along with the HTTP headers"})
             .          .     18:                   return
             .          .     19:           }
             .          .     20:           secretKey := c.MustGet("ConfigKey").(models.CoworkingConfig).SecretKey
             .          .     21:           secretKeyRunes := [32]rune{}
             .          .     22:           for k, v := range secretKey {
             .          .     23:                   secretKeyRunes[k] = v
             .          .     24:           }
             .   100.52MB     25:           claims, err := utils.ValidateToken(tokenHeader, secretKeyRunes)
             .          .     26:           if err != nil {
             .          .     27:                   c.JSON(http.StatusUnauthorized, models.CoworkingErr{Code: models.TokenNotValidErr, Message: err.Error()})
             .          .     28:                   return
             .          .     29:           }
             .          .     30:           email := (*claims)["sub"].(string)
      512.02kB   512.02kB     31:           db := c.MustGet("DbKey").(gorm.DB)
             .   469.79MB     32:           user, err := models.GetUserByEmail(&db, email)
             .          .     33:           if err != nil {
             .          .     34:                   coworkingErr := err.(models.CoworkingErr)
             .          .     35:                   c.JSON(coworkingErr.StatusCode, coworkingErr)
             .          .     36:                   return
             .          .     37:           }
             .          .     38:           c.Set("UserIdKey", user.ID)
             .   347.74MB     39:           c.Next()
             .          .     40:   }
             .          .     41:}
    ```

> In case you don't find the annotated source code, make sure to use the `trim_path` option.

###### cpu.out

Let's start from the investigation of the `cpu.out` profile. Those are some steps you can take:

1. `go tool pprof cpu.out`
2. The Type is `cpu`
3. Next, type `top` for top CPU-consumers
4. Set `nodefraction=0` since there were dropped nodes
5. When you want to narrow things down and look at only one potential culprit, you can run this command: `go tool pprof -nodefraction=0 -list SetupRoutes cpu.out`

    ```text
    Total: 46.89s
    ROUTINE ======================== github.com/ossan-dev/coworkingapp/handlers.SetupRoutes.AuthorizeUser.func3 in /home/ossan/Projects/tech-journey-be/middlewares/auth.go
          20ms     34.47s (flat, cum) 73.51% of Total
             .          .     14:   return func(c *gin.Context) {
             .       20ms     15:           tokenHeader := c.GetHeader("Authorization")
             .          .     16:           if tokenHeader == "" {
             .          .     17:                   c.JSON(http.StatusUnauthorized, models.CoworkingErr{Code: models.MissingTokenErr, Message: "please provide a jwt token along with the HTTP headers"})
             .          .     18:                   return
             .          .     19:           }
             .          .     20:           secretKey := c.MustGet("ConfigKey").(models.CoworkingConfig).SecretKey
             .          .     21:           secretKeyRunes := [32]rune{}
             .          .     22:           for k, v := range secretKey {
             .          .     23:                   secretKeyRunes[k] = v
             .          .     24:           }
             .      1.12s     25:           claims, err := utils.ValidateToken(tokenHeader, secretKeyRunes)
             .          .     26:           if err != nil {
             .          .     27:                   c.JSON(http.StatusUnauthorized, models.CoworkingErr{Code: models.TokenNotValidErr, Message: err.Error()})
             .          .     28:                   return
             .          .     29:           }
             .          .     30:           email := (*claims)["sub"].(string)
             .       30ms     31:           db := c.MustGet("DbKey").(gorm.DB)
             .     23.02s     32:           user, err := models.GetUserByEmail(&db, email)
             .          .     33:           if err != nil {
             .          .     34:                   coworkingErr := err.(models.CoworkingErr)
             .          .     35:                   c.JSON(coworkingErr.StatusCode, coworkingErr)
             .          .     36:                   return
             .          .     37:           }
          10ms       40ms     38:           c.Set("UserIdKey", user.ID)
             .     10.23s     39:           c.Next()
          10ms       10ms     40:   }
             .          .     41:}
    ```

> In case you don't find the annotated source code, make sure to use the `trim_path` option.

##### Graphical Report Generation

> This generation is done via the `graphviz` package. Be sure to have installed it on your machine. First, it generates a file in the `dot` format. Then, all the others format are generated starting from this.

###### allocs.out

To display the data in more UI-friendly style, run:

1. `go tool pprof -svg -nodefraction=0 -output allocs.svg allocs.out`. It will generate an SVG file with the data representation, keep in mind to lower down the `nodefraction` value

###### cpu.out

To display the data in more UI-friendly style, run:

1. `go tool pprof -svg -nodefraction=0 -output cpu.svg cpu.out`. It will generate an SVG file with the data representation, keep in mind to lower down the `nodefraction` value

##### Web UI Interface

###### heaps.out

1. `go tool pprof -http=:8083 -inuse_objects heap.out`
2. `go tool pprof -http=:8083 -alloc_space heap.out`

###### allocs.out

To analyze data in the web server, run `go tool pprof -http=:8081 allocs.out`.  
  
We can do the following considerations:

- `handlers.AddBooking` has a pretty big square (compared to the others). It means has a `cum` value decent
- `handlers.AddBooking` has a decent font size meaning that the `flat` value is consistent
- almost all the nodes connected to the `gin(*Context).Next` has a thick edge meaning that a lot of resources have been used along that path. Plus, the edges are red. These edges are solid (not dashed), meaning that there are no omitted locations in the middle
- The call to the `gin.LoggerWithConfig` has been inlined into the caller which is `gin.Next()`, as you can see from the `inline` on the edge

###### cpu.out

To analyze data in the web server, run `go tool pprof -http=:8082 cpu.out`.  
  
We can do the following considerations:

- there are two calls at the leaf-level of the tree that are impactful. Those **are** `sha256.block` and `Syscall6`. We can try to reduce the amount of total hits on these calls
- in the upper-most part of the graph, there are some nodes with high `cum` and small `flat` values, such as `gin.LoggerWithFunc`, `handlers.AddBooking`, `handlers.SetupRoutes`
- they all have thick, solid, and red edges meaning they're using some resources and there aren't intermediate calls
- The call to the `gin.LoggerWithConfig` has been inlined into the caller which is `gin.Next()`, as you can see from the `inline` on the edge

### 6. trace

#### Generate a Trace with the `debug/pprof/trace` Endpoint

The steps needed to generate a trace to read are:

1. Run the applications with tracing enabled
2. Run the command that starts a 10-second trace. The command is `curl -v <http://localhost:8080/debug/pprof/trace?seconds=10> > trace.out`. It's best to keep it short. Otherwise, it will create multiple trace session to handle large amount of data
3. Run the script to emulate some workload. The amount of operations done should be reduced to finish within the deadline

#### Generate a Trace with the `FlightRecorder` Feature

1. First, make sure to set the `FlightRecorder` up in your production app
2. The setup code in the `main.go` file is:

    ```go
    fr := handlers.NewFlightRecorder()
     if err := fr.FlightRecorderTracer.Start(); err != nil {
      panic(err)
     }
     fr.FlightRecorderTracer.SetSize(4096)
     defer func() {
      if err := fr.FlightRecorderTracer.Stop(); err != nil {
       panic(err)
      }
     }()
    ```

3. Then, you have to create an HTTP handler to serve the request:

    ```go
    func (f *FlightRecorderTracer) Trace(c *gin.Context) {
     file, err := os.Create("cmd/flight_recorder.out")
     if err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
      return
     }
     defer file.Close()
     if _, err := f.FlightRecorderTracer.WriteTo(file); err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
      return
     }
     c.Status(http.StatusOK)
    }
    ```

4. Defer an invocation to the `/trace` endpoint exposed by the HTTP Server
5. You should find a trace file written on your disk

> You can also write the file dynamically in each endpoint based on some custom logic.

#### Read the trace

Those are the considerations:

1. Start reading the trace with the command: `go tool trace trace.out`
2. Amount of time: 2538ms
3. Heap Max Used: 9.2MiB
4. Wall Duration GC: 3.9ms
5. GC time relative to the elapsed time: 0.15%

Two of the options you'll most likely to use are:

- `View trace by proc`: the usual trace execution image you can find around
- `Goroutine analysis`: a table listing all the goroutines' details

### 7. `gops` tool

`gops` is a tool developed at Google (it could be considered like the old version of the `pprof` & `trace` tools). It's needed to monitor and manage Go processes that run either locally or remotely.

#### Installation

First, you have to install it with the command `go install github.com/google/gops@latest`

To confirm its installation, you can run either `which gops` or `gops --help`.

#### Agent

The diagnostic agent reports additionally information about processes we might want to further analyze. The information collected are (non exhaustive list):

1. current stack trace
2. go version used to build the process
3. memory stats

To enable the `gops` agent, we've to make a small code change in the program's startup logic.
We changed the code in the `main.go` file to make it listen to the `gops` agent:

```go
 // gops agent
 if err := agent.Listen(agent.Options{}); err != nil {
  panic(err)
 }
```

#### Usage

First, you have to build the source code app.
`gops` could be used locally by stating the `PID` or remotely by sticking to the `host:port` combination.
Here, we'll show an example of the local usage. After you have run the binary compiled above, you'll be ready to issue some commands.

##### `gops`

By running the `gops` tools without any params you can see the `go` processes running on your machine (the columns names have been added by me):

```shell
PID    PPID  Name          Go version      Location of the program
75672  75315 gopls         go1.23.4        /home/ossan/go/bin/gopls
75686  75672 gopls         go1.23.4        /home/ossan/go/bin/gopls
103370 94075 coworkingapp* go1.23.4        /home/ossan/Projects/tech-journey-be/coworkingapp
103583 75348 gops          go1.23.4        /home/ossan/go/bin/gops
```

The `Go` version is the one we used to build the program.
The **\*** (**star**) means the processes running a `gops` agent (in our case we enabled it above).

##### `gops <PID> [duration]`

We run `gops 103370` to see the details of our app.

```shell
parent PID:     94075
threads:        12
memory usage:   0.136%
cpu usage:      0.406%
username:       ossan
cmd+args:       ./coworkingapp
elapsed time:   07:25
local/remote:   127.0.0.1:46377 <-> 0.0.0.0:0 (LISTEN)
local/remote:   127.0.0.1:33206 <-> 127.0.0.1:54322 (ESTABLISHED)
local/remote:   :::8080 <-> :::0 (LISTEN)
```

By running the command `gops 103370 2s`:

```shell
parent PID:     94075
threads:        13
memory usage:   0.139%
cpu usage:      0.406%
cpu usage (2s): 0.500%
username:       ossan
cmd+args:       ./coworkingapp
elapsed time:   09:11
local/remote:   127.0.0.1:46377 <-> 0.0.0.0:0 (LISTEN)
local/remote:   127.0.0.1:33206 <-> 127.0.0.1:54322 (ESTABLISHED)
local/remote:   :::8080 <-> :::0 (LISTEN)
```

This command will also report the amount of CPU used in the specified period which should match the format `time.ParseDuration` (e.g. `cpu usage (2s): 0.500%`).

##### `gops tree`

It displays all the trees of the current running process:

```shell
...
├── 75315
│   └── 75672 (gopls) {go1.23.4}
│       └── 75686 (gopls) {go1.23.4}
├── 75348
│   └── 111870 (gops) {go1.23.4}
└── 94075
    └── [*]  103370 (coworkingapp) {go1.23.4}
```

#### `gops stack`

`gops stack 103370` will return all the active stack traces per goroutines belonging to the requested `PID`.

```shell
...
goroutine 1 [IO wait, 2 minutes]:
internal/poll.runtime_pollWait(0x7d3bf35c6450, 0x72)
        /usr/local/go/src/runtime/netpoll.go:351 +0x85
internal/poll.(*pollDesc).wait(0xc000230f00?, 0x25?, 0x0)
        /usr/local/go/src/internal/poll/fd_poll_runtime.go:84 +0x27
internal/poll.(*pollDesc).waitRead(...)
        /usr/local/go/src/internal/poll/fd_poll_runtime.go:89
internal/poll.(*FD).Accept(0xc000230f00)
        /usr/local/go/src/internal/poll/fd_unix.go:620 +0x295
net.(*netFD).accept(0xc000230f00)
        /usr/local/go/src/net/fd_unix.go:172 +0x29
net.(*TCPListener).accept(0xc0000abac0)
        /usr/local/go/src/net/tcpsock_posix.go:159 +0x1e
net.(*TCPListener).Accept(0xc0000abac0)
        /usr/local/go/src/net/tcpsock.go:372 +0x30
net/http.(*Server).Serve(0xc0000ca2d0, {0xd4a418, 0xc0000abac0})
        /usr/local/go/src/net/http/server.go:3330 +0x30c
net/http.(*Server).ListenAndServe(0xc0000ca2d0)
        /usr/local/go/src/net/http/server.go:3259 +0x71
net/http.ListenAndServe(...)
        /usr/local/go/src/net/http/server.go:3514
github.com/gin-gonic/gin.(*Engine).Run(0xc000594340, {0xc000051ef8, 0x1, 0x1})
        /home/ossan/go/pkg/mod/github.com/gin-gonic/gin@v1.10.0/gin.go:399 +0x211
main.main()
        /home/ossan/Projects/tech-journey-be/main.go:78 +0x4c5
...
```

#### `gops memstats`

By running `gops memstats 103370`, you could see the memory stats of the requested process:

```shell
alloc: 4.38MB (4590944 bytes)
total-alloc: 52.02MB (54542968 bytes)
sys: 19.40MB (20337928 bytes)
lookups: 0
mallocs: 132297
frees: 124170
heap-alloc: 4.38MB (4590944 bytes)
heap-sys: 11.34MB (11894784 bytes)
heap-idle: 5.15MB (5398528 bytes)
heap-in-use: 6.20MB (6496256 bytes)
heap-released: 3.01MB (3153920 bytes)
heap-objects: 8127
stack-in-use: 672.00KB (688128 bytes)
stack-sys: 672.00KB (688128 bytes)
stack-mspan-inuse: 112.66KB (115360 bytes)
stack-mspan-sys: 159.38KB (163200 bytes)
stack-mcache-inuse: 14.06KB (14400 bytes)
stack-mcache-sys: 15.23KB (15600 bytes)
other-sys: 2.97MB (3112617 bytes)
gc-sys: 2.87MB (3011264 bytes)
next-gc: when heap-alloc >= 7.39MB (7747016 bytes)
last-gc: 2025-02-06 15:56:23.871439777 +0100 CET
gc-pause-total: 2.325873ms
gc-pause: 175077
gc-pause-end: 1738853783871439777
num-gc: 17
num-forced-gc: 0
gc-cpu-fraction: 4.687698217513668e-06
enable-gc: true
debug-gc: false
```

#### `gops gc <PID>`

You can force a GC cycle on the target process. It will block until the GC cycle has been completed.

#### `gops setgc`

It will set the GC on the target process. Examples are:

1. `gops setgc <PID> 10`: to set it to 10%
2. `gops setgc <PID> off`: to turn the GC off

#### `gops version <PID>`

It's used to see with which Go version a program has been built. The command `gops version 103370` yield `go1.23.4`.

#### `gops stats <PID>`

It prints runtime statistics. `gops stats 103370` prints:

```shell
goroutines: 4
OS threads: 15
GOMAXPROCS: 12
num CPU: 12
```

#### `gops pprof` commands

By using `gops` you can also access the capabilities of the `pprof` tool.
Some commands are:

1. `gops pprof-cpu <PID>`: it starts a 30-second CPU profile. Then, you land to the `pprof` interactive console
2. `gops pprof-heap <PID>`: it runs a heap-based profile. Then, you land to the `pprof` interactive console
3. `gops trace <PID>`: it runs a 5-second trace.

> **Please be sure to not have started a trace in your web server, otherwise you'll get an error similar to this `gops: runtime error: tracing is already enabled`.**
