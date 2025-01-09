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
