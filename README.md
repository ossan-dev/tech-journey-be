# 1. Coworking App

This project will sustain my upcoming speech about Profiling & Tracing. It emulates whatever needs to be taken into consideration when aiming to improve the performance of a project regarding memory and CPU point of views. It's like a TODO list.

## 1.1. Steps

### 1.1.1. Discovery Overview

Inspect the source code to get familiar with it. Understand its dependencies, look at config files, potential bottlenecks, and so on.

### 1.1.2. Compile-Time Checks

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

### 1.1.3. `sync.Pool`

If you find out there are several allocations (especialy for large objects), you can do some economy with the `sync.Pool` provided by the `sync` package.  

This is an ideal use case if you make several allocations. It can be quite often when working with multiple goroutines. The `sync.Pool` has also the benefit to be thread-safe.

> Please note if you're allocating small objects, the cost might outperform the benefit.

In the `cmd/booking.go` file, we were running hundreds of thousands HTTP requests against our endpoint. We used the worker pool pattern to share the load between the OS cores visible to our Go program (one goroutine for each logical core).  
  
However, instead of allocating/deallocating the `*http.Request` several times, we could have used the `sync.Pool` objects. This prevent disposal of each instance after the usage and make everything less aggressive in regards to memory. Instead of disposing it, we put it back in the pool.

### 1.1.4. Measure Performance

Now, you've optmized something and something not. The un-optimized things are left as they were. We need to make sure that they doesn't impact performance. Let's use the `runtime` package.
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

### 1.1.5. Using Benchmarks

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

### 1.1.6. `perflock`

The `perflock` tool is used to run benchmarks with a stable and predictable CPU behavior.
This is the process to follow to get rid of noise while CPU benchmarking:

1. Make sure to have installed this `sudo apt install power-profiles-daemon` on your machine
2. Close unnecessary applications
3. `sudo systemctl stop snapd.service`
4. Disable background services that might cause CPU spikes
5. `sudo systemctl stop unattended-upgrades.service`
6. Disable power management features
7. `sudo systemctl stop thermald`
8. `sudo systemctl stop power-profiles-daemon`
9. Install the `perflock` tool provided by aclements
10. `GOBIN=$PWD go install github.com/aclements/perflock/cmd/perflock@latest`
11. `sudo install ./perflock /usr/local/bin/`
12. Install the `cpupower` tools
13. `sudo apt install linux-tools-common linux-tools-generic`
14. Disable CPU frequency scaling (set to performance mode)
15. `sudo cpupower frequency-set -g performance`
16. Disable Turbo Boost (to prevent frequency variations). The command is different on AMD-powered system
17. `echo 0 | sudo tee /sys/devices/system/cpu/intel_pstate/no_turbo`
18. Disable CPU core sleep states
19. `sudo cpupower idle-set -D 0`
20. If you want to isolate CPUs for benchmark:
    1. `sudo nano /etc/default/grub`
    2. Add line `isolcpus=1-3 nohz_full=1-3 rcu_nocbs=1-3`. Add CPU isolation parameters (isolate cores 1-3 for benchmarking). Modify the GRUB_CMDLINE_LINUX to include the line above
    3. `sudo update-grub`
    4. `sudo reboot`
21. In a terminal shell, run `sudo perflock -governor none -daemon` to freeze the CPU settings set above
22. Issue the benchmark command `go test bench=. > base.txt` (you can generate multiple benchmarks here to compare with `benchstat` tool)
23. Ctrl + C to quit the `perflock` process
24. Make sure we no longer have the `perflock` process running with the command `pgrep perflock`
25. Restoring CPU frequency scaling
26. `sudo cpupower frequency-set -g powersave`, you can list the available options by running `sudo cpupower frequency-info`
27. Re-enable Turbo Boost
28. `echo 1 | sudo tee /sys/devices/system/cpu/intel_pstate/no_turbo`
29. Re-enable CPU idle states
30. `sudo cpupower idle-set -E`
31. Restart previously stopped services
32. `sudo systemctl start thermald`
33. `sudo systemctl start power-profiles-daemon`
34. `sudo systemctl start snapd.service`
35. `sudo systemctl start unattended-upgrades.service`
36. Confirm the power profile used by running `powerprofilesctl`

### 1.1.7. `benchcmp`

To compare two benchmarks you can use the `benchcmp` tool. To check the improvements/degrades made by a souce code change (base vs new).  

First, make sure to have it installed on your machine:

```shell
go install golang.org/x/tools/cmd/benchcmp@latest
```

To confirm its installation, run `benchcmp`.  
  
To give it a run, follow these instructions:

1. `cd models`
2. `go test -run='^$' -bench=. -count=2 | tee  benchcmp_base.txt`. This collects the baseline values for our code
3. `go test -run='^$' -bench=. -count=2 | tee  benchcmp_new.txt`. This collects the new values after one or more changes
4. `benchcmp benchcmp_base.txt benchcmp_new.txt`. To compare the two benchmark files collected.

    ```text
    benchcmp is deprecated in favor of benchstat: https://pkg.go.dev/golang.org/x/perf/cmd/benchstat
    benchmark                                                          old ns/op     new ns/op     delta
    BenchmarkParsingJsonFile/ParseModelWithUnmarshal-1-rooms-12        5794          7349          +26.84%
    BenchmarkParsingJsonFile/ParseModelWithUnmarshal-1-rooms-12        5849          6477          +10.74%
    BenchmarkParsingJsonFile/ParseModelWithDecoder-1-rooms-12          5981          6323          +5.72%
    BenchmarkParsingJsonFile/ParseModelWithDecoder-1-rooms-12          6031          6304          +4.53%
    BenchmarkParsingJsonFile/ParseModelWithUnmarshal-999-rooms-12      5802131       6407896       +10.44%
    BenchmarkParsingJsonFile/ParseModelWithUnmarshal-999-rooms-12      6409106       6312111       -1.51%
    BenchmarkParsingJsonFile/ParseModelWithDecoder-999-rooms-12        5159356       5902194       +14.40%
    BenchmarkParsingJsonFile/ParseModelWithDecoder-999-rooms-12        4804592       5462472       +13.69%
    BenchmarkParsingJsonFile/ParseModelWithUnmarshal-9999-rooms-12     34110651      38876070      +13.97%
    BenchmarkParsingJsonFile/ParseModelWithUnmarshal-9999-rooms-12     37132162      38867294      +4.67%
    BenchmarkParsingJsonFile/ParseModelWithDecoder-9999-rooms-12       33686471      35937839      +6.68%
    BenchmarkParsingJsonFile/ParseModelWithDecoder-9999-rooms-12       33761637      34765636      +2.97%

    benchmark                                                          old allocs     new allocs     delta
    BenchmarkParsingJsonFile/ParseModelWithUnmarshal-1-rooms-12        20             20             +0.00%
    BenchmarkParsingJsonFile/ParseModelWithUnmarshal-1-rooms-12        20             20             +0.00%
    BenchmarkParsingJsonFile/ParseModelWithDecoder-1-rooms-12          23             23             +0.00%
    BenchmarkParsingJsonFile/ParseModelWithDecoder-1-rooms-12          23             23             +0.00%
    BenchmarkParsingJsonFile/ParseModelWithUnmarshal-999-rooms-12      8030           8030           +0.00%
    BenchmarkParsingJsonFile/ParseModelWithUnmarshal-999-rooms-12      8030           8030           +0.00%
    BenchmarkParsingJsonFile/ParseModelWithDecoder-999-rooms-12        8018           8018           +0.00%
    BenchmarkParsingJsonFile/ParseModelWithDecoder-999-rooms-12        8018           8018           +0.00%
    BenchmarkParsingJsonFile/ParseModelWithUnmarshal-9999-rooms-12     80043          80043          +0.00%
    BenchmarkParsingJsonFile/ParseModelWithUnmarshal-9999-rooms-12     80043          80043          +0.00%
    BenchmarkParsingJsonFile/ParseModelWithDecoder-9999-rooms-12       80023          80024          +0.00%
    BenchmarkParsingJsonFile/ParseModelWithDecoder-9999-rooms-12       80023          80024          +0.00%

    benchmark                                                          old bytes     new bytes     delta
    BenchmarkParsingJsonFile/ParseModelWithUnmarshal-1-rooms-12        1184          1184          +0.00%
    BenchmarkParsingJsonFile/ParseModelWithUnmarshal-1-rooms-12        1184          1184          +0.00%
    BenchmarkParsingJsonFile/ParseModelWithDecoder-1-rooms-12          1408          1408          +0.00%
    BenchmarkParsingJsonFile/ParseModelWithDecoder-1-rooms-12          1408          1408          +0.00%
    BenchmarkParsingJsonFile/ParseModelWithUnmarshal-999-rooms-12      2196205       2196299       +0.00%
    BenchmarkParsingJsonFile/ParseModelWithUnmarshal-999-rooms-12      2196213       2196189       -0.00%
    BenchmarkParsingJsonFile/ParseModelWithDecoder-999-rooms-12        1261316       1261337       +0.00%
    BenchmarkParsingJsonFile/ParseModelWithDecoder-999-rooms-12        1261293       1261312       +0.00%
    BenchmarkParsingJsonFile/ParseModelWithUnmarshal-9999-rooms-12     28643201      28643208      +0.00%
    BenchmarkParsingJsonFile/ParseModelWithUnmarshal-9999-rooms-12     28643216      28643214      -0.00%
    BenchmarkParsingJsonFile/ParseModelWithDecoder-9999-rooms-12       18910880      18910883      +0.00%
    BenchmarkParsingJsonFile/ParseModelWithDecoder-9999-rooms-12       18910843      18910884      +0.00%
    ```

### 1.1.8. `benchstat`

To understand how a change impacts performance we should get a performance delta. `benchstat` helps with the **A/B** testing.  
First, be sure to have installed it on your machine. If not, please run:

```shell
go install golang.org/x/perf/cmd/benchstat@latest
```

To confirm its installation, run `benchstat`.  
To give it a run, follow these instructions:

1. `cd models`
2. `go test -run='^$' -bench=. -count=2 > base.txt`. This collects the baseline values for our code
3. `go test -run='^$' -bench=. -count=2 > new.txt`. This collects the new values after one or more changes
4. `benchstat base.txt new.txt`. This will compare the two files collected.

    ```text
    goos: linux
    goarch: amd64
    pkg: github.com/ossan-dev/coworkingapp/models
    cpu: 13th Gen Intel(R) Core(TM) i7-1355U
                                                          │   base.txt   │               new.txt                │
                                                          │    sec/op    │    sec/op     vs base                │
    ParsingJsonFile/ParseModelWithUnmarshal-1-rooms-12      6.674µ ± ∞ ¹   6.298µ ± ∞ ¹       ~ (p=0.667 n=2) ²
    ParsingJsonFile/ParseModelWithDecoder-1-rooms-12        6.776µ ± ∞ ¹   7.054µ ± ∞ ¹       ~ (p=1.000 n=2) ²
    ParsingJsonFile/ParseModelWithUnmarshal-999-rooms-12    7.189m ± ∞ ¹   6.905m ± ∞ ¹       ~ (p=0.333 n=2) ²
    ParsingJsonFile/ParseModelWithDecoder-999-rooms-12      5.258m ± ∞ ¹   5.305m ± ∞ ¹       ~ (p=1.000 n=2) ²
    ParsingJsonFile/ParseModelWithUnmarshal-9999-rooms-12   36.26m ± ∞ ¹   37.73m ± ∞ ¹       ~ (p=0.333 n=2) ²
    ParsingJsonFile/ParseModelWithDecoder-9999-rooms-12     34.64m ± ∞ ¹   35.80m ± ∞ ¹       ~ (p=0.333 n=2) ²
    geomean                                                 1.136m         1.140m        +0.39%
    ¹ need >= 6 samples for confidence interval at level 0.95
    ² need >= 4 samples to detect a difference at alpha level 0.05

                                                          │   base.txt    │                new.txt                │
                                                          │     B/op      │     B/op       vs base                │
    ParsingJsonFile/ParseModelWithUnmarshal-1-rooms-12      1.156Ki ± ∞ ¹   1.156Ki ± ∞ ¹       ~ (p=1.000 n=2) ²
    ParsingJsonFile/ParseModelWithDecoder-1-rooms-12        1.375Ki ± ∞ ¹   1.375Ki ± ∞ ¹       ~ (p=1.000 n=2) ²
    ParsingJsonFile/ParseModelWithUnmarshal-999-rooms-12    2.095Mi ± ∞ ¹   2.094Mi ± ∞ ¹       ~ (p=0.667 n=2) ³
    ParsingJsonFile/ParseModelWithDecoder-999-rooms-12      1.203Mi ± ∞ ¹   1.203Mi ± ∞ ¹       ~ (p=0.667 n=2) ³
    ParsingJsonFile/ParseModelWithUnmarshal-9999-rooms-12   27.32Mi ± ∞ ¹   27.32Mi ± ∞ ¹       ~ (p=1.000 n=2) ³
    ParsingJsonFile/ParseModelWithDecoder-9999-rooms-12     18.03Mi ± ∞ ¹   18.03Mi ± ∞ ¹       ~ (p=1.000 n=2) ³
    geomean                                                 359.8Ki         359.8Ki        -0.00%
    ¹ need >= 6 samples for confidence interval at level 0.95
    ² all samples are equal
    ³ need >= 4 samples to detect a difference at alpha level 0.05

                                                          │   base.txt   │               new.txt                │
                                                          │  allocs/op   │  allocs/op    vs base                │
    ParsingJsonFile/ParseModelWithUnmarshal-1-rooms-12       20.00 ± ∞ ¹    20.00 ± ∞ ¹       ~ (p=1.000 n=2) ²
    ParsingJsonFile/ParseModelWithDecoder-1-rooms-12         23.00 ± ∞ ¹    23.00 ± ∞ ¹       ~ (p=1.000 n=2) ²
    ParsingJsonFile/ParseModelWithUnmarshal-999-rooms-12    8.030k ± ∞ ¹   8.030k ± ∞ ¹       ~ (p=1.000 n=2) ²
    ParsingJsonFile/ParseModelWithDecoder-999-rooms-12      8.018k ± ∞ ¹   8.018k ± ∞ ¹       ~ (p=1.000 n=2) ²
    ParsingJsonFile/ParseModelWithUnmarshal-9999-rooms-12   80.04k ± ∞ ¹   80.04k ± ∞ ¹       ~ (p=1.000 n=2) ²
    ParsingJsonFile/ParseModelWithDecoder-9999-rooms-12     80.02k ± ∞ ¹   80.02k ± ∞ ¹       ~ (p=1.000 n=2) ³
    geomean                                                 2.397k         2.397k        +0.00%
    ¹ need >= 6 samples for confidence interval at level 0.95
    ² all samples are equal
    ³ need >= 4 samples to detect a difference at alpha level 0.05
    ```

To have significant values, you should stick to, at least, 10 runs of the benchmark. Ideally, this number should be close to 20. The higher the runs are more reduced will be the noise.  
Benchmarks should also be run on an idle machine to not mess up with the outcome.  
  
To speed things up, you can pre-compile the binary by running the command `go test -c`.

### 1.1.9. memprofile, cpuprofile, pprof

> Starting from now, these metrics should be captured while the application is running. Best would be in a production environment where you can experiment the workload. If you're application is not running, you have to write a script that can stress out your app to understand how it behaves under heavy loads.

Your machine must have installed Graphiz. You can install it by running this command:

`sudo apt install graphviz`

To enable `pprof` we have to first discern between Web and non Web applications. For this project will only care about Web applications.  

Again, we've to make another distinction between Web servers using `DefaultServeMux` and Web servers not using it.
Since our application is written with the **Gin Web Framework**, we've to make some extra steps, as discussed in the following section.  

#### 1.1.9.1. Web Servers Using `DefaultServeMux`

It's usually enough to add `_ "net/http/pprof"` in the import section.

#### 1.1.9.2. Web Servers Not Using `DefaultServerMux`

This is the process for a web server which empowers the `Gin Web Framework`.  

First, you have to download a third-party pkg by using this command: `go get -u github.com/gin-contrib/pprof`
Then, you can register the `pprof` endpoints by referencing `pprof.Register(r)` in `main.go` file.

Now, it's time to run the application and let it generates a profile that can be inspected.

#### 1.1.9.3. Collecting Profiles

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
  
#### 1.1.9.4. Visualizing and Understading Profiles Data

For the sake of the demo, we won't explain how to visualize/understand the heap profile. We'll focus only on `allocs` and `cpu` profiles.

##### 1.1.9.4.1. Interactive Console

###### 1.1.9.4.1.1. allocs.out

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

###### 1.1.9.4.1.2. cpu.out

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

##### 1.1.9.4.2. Graphical Report Generation

> This generation is done via the `graphviz` package. Be sure to have installed it on your machine. First, it generates a file in the `dot` format. Then, all the others format are generated starting from this.

###### 1.1.9.4.2.1. allocs.out

To display the data in more UI-friendly style, run:

1. `go tool pprof -svg -nodefraction=0 -output allocs.svg allocs.out`. It will generate an SVG file with the data representation, keep in mind to lower down the `nodefraction` value

###### 1.1.9.4.2.2. cpu.out

To display the data in more UI-friendly style, run:

1. `go tool pprof -svg -nodefraction=0 -output cpu.svg cpu.out`. It will generate an SVG file with the data representation, keep in mind to lower down the `nodefraction` value

##### 1.1.9.4.3. Web UI Interface

###### 1.1.9.4.3.1. heaps.out

1. `go tool pprof -http=:8083 -inuse_objects heap.out`
2. `go tool pprof -http=:8083 -alloc_space heap.out`

###### 1.1.9.4.3.2. allocs.out

To analyze data in the web server, run `go tool pprof -http=:8081 allocs.out`.  
  
We can do the following considerations:

- `handlers.AddBooking` has a pretty big square (compared to the others). It means has a `cum` value decent
- `handlers.AddBooking` has a decent font size meaning that the `flat` value is consistent
- almost all the nodes connected to the `gin(*Context).Next` has a thick edge meaning that a lot of resources have been used along that path. Plus, the edges are red. These edges are solid (not dashed), meaning that there are no omitted locations in the middle
- The call to the `gin.LoggerWithConfig` has been inlined into the caller which is `gin.Next()`, as you can see from the `inline` on the edge

###### 1.1.9.4.3.3. cpu.out

To analyze data in the web server, run `go tool pprof -http=:8082 cpu.out`.  
  
We can do the following considerations:

- there are two calls at the leaf-level of the tree that are impactful. Those **are** `sha256.block` and `Syscall6`. We can try to reduce the amount of total hits on these calls
- in the upper-most part of the graph, there are some nodes with high `cum` and small `flat` values, such as `gin.LoggerWithFunc`, `handlers.AddBooking`, `handlers.SetupRoutes`
- they all have thick, solid, and red edges meaning they're using some resources and there aren't intermediate calls
- The call to the `gin.LoggerWithConfig` has been inlined into the caller which is `gin.Next()`, as you can see from the `inline` on the edge

### 1.1.10. `perf` Linux tool

To profile CPU-bound functions of the tools used to invoke the tests, you should follow this guide:

1. located in `/cmd` folder
2. `go build -gcflags "-N -l" -o booking_cmd` => build by disabling optimizations and inlining
3. `./booking_cmd` => to run the program to profile
4. in another shell, run `sudo perf record -p $(pidof booking_cmd) -g --call-graph dwarf sleep 45` => it profiles the program specified by the PID
    1. `record` tells to profile data
    2. `-p $(pidof booking_cmd)` specifies the PID of the Go program to profile
    3. `-g`enables call-graph recording, which shows the function call stack leading to the events
    4. `--call-graph dwarf` specifies the method for capturing the call graph (DWARF Debugging Information)
5. it produces something like (and a `perf.data` file):

    ```text
    [ perf record: Woken up 1586 times to write data ]
    [ perf record: Captured and wrote 407,624 MB perf.data (50367 samples) ]
    ```

6. `sudo perf report -i perf.data --stdio` to visualize data in the STD console
7. `sudo perf report -i perf.data` to visualize data in the interactive console

> The recommend way to visualize stuff is via the flamegraph by [Brendan Gregg](https://github.com/brendangregg/FlameGraph)

### 1.1.11. trace

#### 1.1.11.1. Generate a Trace with the `debug/pprof/trace` Endpoint

The steps needed to generate a trace to read are:

1. Run the applications with tracing enabled
2. Run the command that starts a 10-second trace. The command is `curl -v <http://localhost:8080/debug/pprof/trace?seconds=10> > trace.out`. It's best to keep it short. Otherwise, it will create multiple trace session to handle large amount of data
3. Run the script to emulate some workload. The amount of operations done should be reduced to finish within the deadline

#### 1.1.11.2. Generate a Trace with the `FlightRecorder` Feature

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

#### 1.1.11.3. Read the trace

Those are the considerations:

1. Start reading the trace with the command: `go tool trace trace.out`
2. Amount of time: 2538ms
3. Heap Max Used: 9.2MiB
4. Wall Duration GC: 3.9ms
5. GC time relative to the elapsed time: 0.15%

Two of the options you'll most likely to use are:

- `View trace by proc`: the usual trace execution image you can find around
- `Goroutine analysis`: a table listing all the goroutines' details

### 1.1.12. `gops` tool

`gops` is a tool developed at Google (it could be considered like the old version of the `pprof` & `trace` tools). It's needed to monitor and manage Go processes that run either locally or remotely.

#### 1.1.12.1. Installation

First, you have to install it with the command `go install github.com/google/gops@latest`

To confirm its installation, you can run either `which gops` or `gops --help`.

#### 1.1.12.2. Agent

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

#### 1.1.12.3. Usage

First, you have to build the source code app.
`gops` could be used locally by stating the `PID` or remotely by sticking to the `host:port` combination.
Here, we'll show an example of the local usage. After you have run the binary compiled above, you'll be ready to issue some commands.

##### 1.1.12.3.1. `gops`

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

##### 1.1.12.3.2. `gops <PID> [duration]`

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

##### 1.1.12.3.3. `gops tree`

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

#### 1.1.12.4. `gops stack`

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

#### 1.1.12.5. `gops memstats`

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

#### 1.1.12.6. `gops gc <PID>`

You can force a GC cycle on the target process. It will block until the GC cycle has been completed.

#### 1.1.12.7. `gops setgc`

It will set the GC on the target process. Examples are:

1. `gops setgc <PID> 10`: to set it to 10%
2. `gops setgc <PID> off`: to turn the GC off

#### 1.1.12.8. `gops version <PID>`

It's used to see with which Go version a program has been built. The command `gops version 103370` yield `go1.23.4`.

#### 1.1.12.9. `gops stats <PID>`

It prints runtime statistics. `gops stats 103370` prints:

```shell
goroutines: 4
OS threads: 15
GOMAXPROCS: 12
num CPU: 12
```

#### 1.1.12.10. `gops pprof` commands

By using `gops` you can also access the capabilities of the `pprof` tool.
Some commands are:

1. `gops pprof-cpu <PID>`: it starts a 30-second CPU profile. Then, you land to the `pprof` interactive console
2. `gops pprof-heap <PID>`: it runs a heap-based profile. Then, you land to the `pprof` interactive console
3. `gops trace <PID>`: it runs a 5-second trace.

> **Please be sure to not have started a trace in your web server, otherwise you'll get an error similar to this `gops: runtime error: tracing is already enabled`.**

### 1.1.13. Stress-Tests Tools

Here, we'll see a couple of tools that will help you in smoothening your stress-testing experience. The first one is `hey`.

#### 1.1.13.1. The `hey` tool

First, be sure to have `hey` tool installed on your machine. You can download it from [their GitHub repo](https://github.com/rakyll/hey). After you downloaded the binary, please move it to the `$PATH` folder to run it anywhere.

> You can rename to **hey** the binary.

Then, run the command `which hey` or `hey` to confirm its installation.

Before going ahead, let's install another tool (via Go) that will help us with the visualization. Run the command `go get -u github.com/asoorm/hey-hdr` for a system-wide installation of the tool.

First, make sure to run the application. Then, you're ready to run some commands and inspect the results.  
  
Here's a list commands run:

1. `hey http://localhost:8080/rooms`

    ```text
    Summary:
      Total:        0.1949 secs
      Slowest:      0.1462 secs
      Fastest:      0.0004 secs
      Average:      0.0413 secs
      Requests/sec: 1026.3224
      
      Total data:   109400 bytes
      Size/request: 547 bytes

    Response time histogram:
      0.000 [1]     |
      0.015 [81]    |■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
      0.030 [14]    |■■■■■■■
      0.044 [28]    |■■■■■■■■■■■■■■
      0.059 [16]    |■■■■■■■■
      0.073 [17]    |■■■■■■■■
      0.088 [11]    |■■■■■
      0.102 [8]     |■■■■
      0.117 [11]    |■■■■■
      0.132 [9]     |■■■■
      0.146 [4]     |■■


    Latency distribution:
      10% in 0.0028 secs
      25% in 0.0071 secs
      50% in 0.0314 secs
      75% in 0.0671 secs
      90% in 0.1086 secs
      95% in 0.1205 secs
      99% in 0.1459 secs

    Details (average, fastest, slowest):
      DNS+dialup:   0.0002 secs, 0.0004 secs, 0.1462 secs
      DNS-lookup:   0.0002 secs, 0.0000 secs, 0.0011 secs
      req write:    0.0000 secs, 0.0000 secs, 0.0005 secs
      resp wait:    0.0409 secs, 0.0004 secs, 0.1447 secs
      resp read:    0.0001 secs, 0.0000 secs, 0.0020 secs

    Status code distribution:
      [200] 200 responses
    ```

2. The pattern seems okay. We have 43% of the total calls with a response time `<= 15ms`. If we take into consideration `50ms` as an acceptable time, the percentage goes up to 65.60%. We won't see any increasing pattern or spike in the calls with most latency
3. `hey -n 10000 -c 8 http://localhost:8080/rooms` will send 10.000 request by using 8 connections

    ```text
    Summary:
      Total:        3.2958 secs
      Slowest:      0.0339 secs
      Fastest:      0.0002 secs
      Average:      0.0026 secs
      Requests/sec: 3034.1612
      
      Total data:   5470000 bytes
      Size/request: 547 bytes

    Response time histogram:
      0.000 [1]     |
      0.004 [7384]  |■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
      0.007 [1599]  |■■■■■■■■■
      0.010 [727]   |■■■■
      0.014 [210]   |■
      0.017 [46]    |
      0.020 [22]    |
      0.024 [6]     |
      0.027 [2]     |
      0.031 [2]     |
      0.034 [1]     |


    Latency distribution:
      10% in 0.0004 secs
      25% in 0.0006 secs
      50% in 0.0010 secs
      75% in 0.0048 secs
      90% in 0.0069 secs
      95% in 0.0084 secs
      99% in 0.0131 secs

    Details (average, fastest, slowest):
      DNS+dialup:   0.0000 secs, 0.0002 secs, 0.0339 secs
      DNS-lookup:   0.0000 secs, 0.0000 secs, 0.0003 secs
      req write:    0.0000 secs, 0.0000 secs, 0.0023 secs
      resp wait:    0.0025 secs, 0.0001 secs, 0.0337 secs
      resp read:    0.0000 secs, 0.0000 secs, 0.0010 secs

    Status code distribution:
      [200] 10000 responses
    ```

4. A full list of params can be found by running `hey`:

    ```text
    Usage: hey [options...] <url>

    Options:
      -n  Number of requests to run. Default is 200.
      -c  Number of workers to run concurrently. Total number of requests cannot
          be smaller than the concurrency level. Default is 50.
      -q  Rate limit, in queries per second (QPS) per worker. Default is no rate limit.
      -z  Duration of application to send requests. When duration is reached,
          application stops and exits. If duration is specified, n is ignored.
          Examples: -z 10s -z 3m.
      -o  Output type. If none provided, a summary is printed.
          "csv" is the only supported alternative. Dumps the response
          metrics in comma-separated values format.

      -m  HTTP method, one of GET, POST, PUT, DELETE, HEAD, OPTIONS.
      -H  Custom HTTP header. You can specify as many as needed by repeating the flag.
          For example, -H "Accept: text/html" -H "Content-Type: application/xml" .
      -t  Timeout for each request in seconds. Default is 20, use 0 for infinite.
      -A  HTTP Accept header.
      -d  HTTP request body.
      -D  HTTP request body from file. For example, /home/user/file.txt or ./file.txt.
      -T  Content-type, defaults to "text/html".
      -a  Basic authentication, username:password.
      -x  HTTP Proxy address as host:port.
      -h2 Enable HTTP/2.

      -host HTTP Host header.

      -disable-compression  Disable compression.
      -disable-keepalive    Disable keep-alive, prevents re-use of TCP
                            connections between different HTTP requests.
      -disable-redirects    Disable following of HTTP redirects
      -cpus                 Number of used cpu cores.
                            (default for current machine is 12 cores)
    ```

#### 1.1.13.2. The Apache Bench or `ab` tool

Another tool for performance testing is Apache Bench. The installation process is:

1. Run `sudo apt-get update`
2. Run `sudo apt-get install apache2-utils`. Please note that this tool can be installed on whatever machine you want. It doesn't have to be installed on the same machine that hosts the target web server

To verify the installation, run the command `ab -V` to print the version and exit.  
  
We can run some commands:

1. `ab -c 100 -n 20000 -r <http://localhost:8080/rooms>` where:
    1. `-c` means the number of requests to perform per time. Default is one request per time
    2. `-n` means the number of requests to perform in the benchmarking session. Default is to perform a single request
    3. `-r` means to not exit on socket receive errors
    4. As a positional argument we've to specify the target URL to benchmark

    ```text
    Benchmarking localhost (be patient)
    Completed 2000 requests
    Completed 4000 requests
    Completed 6000 requests
    Completed 8000 requests
    Completed 10000 requests
    Completed 12000 requests
    Completed 14000 requests
    Completed 16000 requests
    Completed 18000 requests
    Completed 20000 requests
    Finished 20000 requests


    Server Software:        
    Server Hostname:        localhost
    Server Port:            8080

    Document Path:          /rooms
    Document Length:        547 bytes

    Concurrency Level:      100
    Time taken for tests:   15.881 seconds
    Complete requests:      20000
    Failed requests:        0
    Total transferred:      16380000 bytes
    HTML transferred:       10940000 bytes
    Requests per second:    1259.38 [#/sec] (mean)
    Time per request:       79.404 [ms] (mean)
    Time per request:       0.794 [ms] (mean, across all concurrent requests)
    Transfer rate:          1007.26 [Kbytes/sec] received

    Connection Times (ms)
                  min  mean[+/-sd] median   max
    Connect:        0    0   0.5      0       9
    Processing:     1   79  65.2     96     318
    Waiting:        0   78  65.2     95     318
    Total:          1   79  65.3     96     318

    Percentage of the requests served within a certain time (ms)
      50%     96
      66%    117
      75%    131
      80%    137
      90%    154
      95%    170
      98%    204
      99%    252
     100%    318 (longest request)
    ```

2. Some relevant values are:
    1. _Document Length_: it's the size in bytes of the first successful document returned. If the response length changes during the test, it's considered an error
    2. _Concurrency Level_: number of concurrent clients (equivalents to web browsers) used during the test
    3. _Time taken for tests_: elapsed time between the creation of the first socket to the moment the last response is received
    4. _Complete requests_: number of successful responses received
    5. _Requests per seconds_: self-explanatory
    6. _Time per request_: self-explanatory

### 1.1.14. coredumps, crashdumps

Before you start looking into coredumps, make sure your `ulimit` is set to something reasonable. It defaults to `0`, which means the max core file size can be `zero`. On development machine, this can be set to `unlimited` by using this command:

`ulimit -c unlimited`

To check the value, you can run `ulimit -c` and you should get back `unlimited`.

> _Please note this command must be run in the folder where you want to generate core dump files._

**Crashdumps** are **coredumps** written to the disk when a program is crashing. By default Go doesn't enable crashdumps but you can do the following:

1. Build your application.  
    **Please note that the values of variables won't be clear due to compiler optimizations.**
2. Make sure to build the app with the command: `go build -gcflags=all="-N -l" -o coworkingapp`
3. Run it with the command `GOTRACEBACK=crash ./coworkingapp`
4. Send the signal `SIGQUIT` to the web server within the same active shell. The keyboard shortcut should be `Ctrl + \` but this may vary based on your machine/shell/IDE
5. You will have the program stak traces printed in the console (and a core dump file written)
6. use the `coredumpctl` tool to load the crashdump file
7. to install it: `sudo apt-get install systemd-coredump`
8. run `coredumpctl` to list the crashdumps you can load
9. `coredumpctl debug --debugger=dlv --debugger-arguments=core` to load the crashdump

To get a core dump without having to kill the process, you can follow this:

1. Run the web server with `./coworkingapp` (the compilation is the same as above)
2. Find the PID of the server (you can use `gops`)
3. Download a core dump file with the `gcore` command `gcore <PID>`
4. You can start inspecting the file by using the command: `dlv core ./coworkingapp core.44457`, where `./coworkingapp` is the name of the binary and `core.44457` is the core dump file

From this point you have all the commands of the `dlv` interactive console such as:

- `bt` to list frames and see the last visible frame
- `ls` to show the source code related to the last visible frame
- `frame <frameNumber>` to pick a specific frame. You can use the number on the left
- `locals` to print the values of local variables in scope at the moment of the crash
- `p <variableName>` to check value of a specific variable
- `vars <packageName>` to dump package variables

Some features will be disabled since the core dump is not an active process but it's a snapshot.

### 1.1.15. Go Env Vars

We can set Go env vars to change the runtime behavior.
> Make sure to rebuild the binary since the compiler optimizations have been disabled by the last build command.

#### 1.1.15.1. GOGC

The `GOGC` var sets the aggressviness of the Garbage Collector. Default Value `100` (when the heap doubles in size).

1. To run it **less** often, let's say when the heap gets 4x: `GOGC=200 ./coworkingapp`
2. To run it **more** often, run `GOGC=20 ./coworkingapp`
3. To disable it, `GOGC=off ./coworkingapp`

### 1.1.16. GOTRACEBACK

The `GOTRACEBACK` controls the level of details when a panic hits the top of our program. Default Value `single` (prints only the goroutine which seems to have caused the issue).

1. To run it and suppress all tracebacks, `GOTRACEBACK=none ./coworkingapp`

### 1.1.17. GOMAXPROCS

> Since we're going to use the trace tool here, please be sure to be able to download a trace from the trace endpoint. If not, disable the trace added in the `gops` section and rebuild the binary.  
  
The `GOMAXPROCS` controls the number of OS threads allocated to goroutines in our program. Default Value is the number of cores (or whatever your machine considers a CPU) visible at program startup.

1. `GOMAXPROCS=4 ./coworkingapp` will set the number of OS threads to use to `4`
2. To confirm this, you can issue the command `curl -v http://localhost:8080/debug/pprof/trace?seconds=10 > four_procs_trace.out`
3. Open the trace with `go tool trace four_procs_trace.out`
4. Select `View by procs`. You should see a decreased number of procs used

### 1.1.18. GOMEMLIMIT

Used to set the maximum amount of memory the Go program can use. This is usually set when you experience `OOM` crashes. A reasonable value could be the 80% of the total machine memory.
To limit the program to use only 5MiB, you should use the command:

```shell
GOMEMLIMIT=5000000 ./coworkingapp
```

### 1.1.19. GODEBUG

To better understand the GC performance, we can enable the `gctrace` facility. To enable the program to report the GC data, you should run it with this command:

```shell
GODEBUG=gctrace=1 ./coworkingapp
```

The first output line will be something close to:

```text
gc 1 @0.009s 2%: 0.087+1.0+0.011 ms clock, 1.0+0.72/1.7/0.040+0.13 ms cpu, 4->4->3 MB, 4 MB goal, 0 MB stacks, 0 MB globals, 12 P
```

It shows you:

- amount of time spent in each GC phase
- various heap size during the GC cycles
- timestamp of when the GC cycles completed compared to the start time of the program

If you leave the program up, you'll notice other entries in the console such as:

```text
gc 2 @27.055s 0%: 0.13+1.4+0.032 ms clock, 1.5+0.37/2.8/0.71+0.39 ms cpu, 6->6->3 MB, 6 MB goal, 0 MB stacks, 0 MB globals, 12 P
gc 3 @121.179s 0%: 0.14+1.4+0.042 ms clock, 1.6+0.57/2.3/1.1+0.51 ms cpu, 7->7->3 MB, 7 MB goal, 0 MB stacks, 0 MB globals, 12 P
gc 4 @209.288s 0%: 0.090+1.8+0.042 ms clock, 1.0+0/3.9/1.6+0.51 ms cpu, 7->7->3 MB, 7 MB goal, 0 MB stacks, 0 MB globals, 12 P
gc 5 @303.429s 0%: 0.14+2.5+0.043 ms clock, 1.7+0.50/3.7/0.73+0.51 ms cpu, 7->7->3 MB, 7 MB goal, 0 MB stacks, 0 MB globals, 12 P
gc 6 @393.573s 0%: 0.23+2.0+0.025 ms clock, 2.8+0.70/3.0/1.8+0.30 ms cpu, 7->7->3 MB, 7 MB goal, 0 MB stacks, 0 MB globals, 12 P
```

The explanation is as follows:

```text
Currently, it is:
 gc # @#s #%: #+#+# ms clock, #+#/#/#+# ms cpu, #->#-># MB, # MB goal, # MB stacks, #MB globals, # P
where the fields are as follows:
 gc #         the GC number, incremented at each GC
 @#s          time in seconds since program start
 #%           percentage of time spent in GC since program start
 #+...+#      wall-clock/CPU times for the phases of the GC
 #->#-># MB   heap size at GC start, at GC end, and live heap, or /gc/scan/heap:bytes
 # MB goal    goal heap size, or /gc/heap/goal:bytes
 # MB stacks  estimated scannable stack size, or /gc/scan/stack:bytes
 # MB globals scannable global size, or /gc/scan/globals:bytes
 # P          number of processors used, or /sched/gomaxprocs:threads
```
