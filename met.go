package main

import (
    "golang.org/x/net/context"
    "fmt"
    "time"
    "math"
    "bench.met/bench/dcr"
    "bench.met/bench/prt"
    "encoding/json"
    "flag"
    "github.com/rcrowley/go-metrics"
    "strconv"
)

const workersCount = 100
const toWaitSec = 30

type Result struct { count int }

func metricsDump(container, prop string, val uint64) {
    fmt.Printf("%v.%v %v\n", container, prop, val)
}

func timeTrack(start time.Time, name string) {
    elapsed := time.Since(start)
    fmt.Printf("%s took %s\n", name, elapsed)
}


func dockerProcStream(workers int, toRun time.Duration) {
    fmt.Println("Running docker-stream sample")
    defer fmt.Println("done")

    s := metrics.NewExpDecaySample(1028, 0.015)
    parseTimings := metrics.NewHistogram(s)
    metrics.Register("parseDockerTimings", parseTimings)

    ctx := context.Background()
    cli, err := dcr.MakeDockerClient()
    if err != nil {
        panic(err)
    }

    defer cli.Close()

    control := make(chan bool, workers)
    results := make(chan Result, workers)
    ready := make(chan bool, workers)
    start := make(chan bool, workers)

    for i := 0; i < workers; i++ {
        id, err := dcr.MakeDockerContainer(ctx, cli, fmt.Sprintf("docker.%d", i))
        if err != nil {
            panic(err)
        }
        defer dcr.CleanupDockerContainer(ctx, cli, id)

        go func() {
            stream := dcr.GetDockerStatsStream(ctx, cli, id)
            defer stream.Close()

            dec := json.NewDecoder(stream)

            var done bool
            var count int

            ready <- true
            <- start

            for !done {
                select {
                case done = <- control:
                default:
                    now := time.Now()
                    dcr.ParseDockerStats(dec)
                    count++
                    parseTimings.Update(int64(time.Since(now)))
                }
            }

            results <- Result{count}
        }()
    }

    for i := 0; i < workers; i++ {
        <- ready
    }

    fmt.Printf("All %d subroutines are ready\n", workers)
    close(start)

    t := time.NewTimer(toRun)
    select {
    case <- t.C:
        fmt.Println("completion timer fired")

        sum := 0
        for i := 0; i < workers; i++ {
            control <- true
            res := <- results
            sum += res.count
        }

        fmt.Printf("total requests: %d, avg time %.2fms\n", sum, parseTimings.Mean() / math.Pow(1000.0,2))
    }
}

func dockerProc(workers int, toRun time.Duration) {
    fmt.Println("Running docker sample")
    defer fmt.Println("done")

    s := metrics.NewExpDecaySample(1028, 0.015)
    parseTimings := metrics.NewHistogram(s)
    metrics.Register("parseDockerTimings", parseTimings)

    ctx := context.Background()
    cli, err := dcr.MakeDockerClient()
    if err != nil {
        panic(err)
    }

    defer cli.Close()

    control := make(chan bool, workers)
    results := make(chan Result, workers)
    ready := make(chan bool, workers)
    start := make(chan bool, workers)

    for i := 0; i < workers; i++ {
        id, err := dcr.MakeDockerContainer(ctx, cli, fmt.Sprintf("docker.%d", i))
        if err != nil {
            panic(err)
        }
        defer dcr.CleanupDockerContainer(ctx, cli, id)

        go func() {

            var done bool
            var count int

            ready <- true
            <- start

            for !done {
                select {
                case done = <- control:
                default:
                    now := time.Now()
                    dcr.GetDockerStats(ctx, cli, id)
                    count++
                    parseTimings.Update(int64(time.Since(now)))
                }
            }

            results <- Result{count}
        }()
    }

    for i := 0; i < workers; i++ {
        <- ready
    }

    fmt.Printf("All %d subroutines are ready\n", workers)
    close(start)

    t := time.NewTimer(toRun)
    select {
    case <- t.C:
        fmt.Println("completion timer fired")

        sum := 0

        for i := 0; i < workers; i++ {
            control <- true
            res := <- results
            sum += res.count
        }

        fmt.Printf("total requests: %d, avg time %.2fms\n", sum, parseTimings.Mean() / math.Pow(1000.0,2))
    }
}

func portoProc(workers int, toRun time.Duration) {
    fmt.Println("Running porto sample")
    defer fmt.Println("done")

    var containers []string

    control := make(chan bool, workers)
    results := make(chan Result, workers)

    s := metrics.NewExpDecaySample(1028, 0.015)
    parseTimings := metrics.NewHistogram(s)
    metrics.Register("parsePortoTimings", parseTimings)

    // useless for now
    s = metrics.NewExpDecaySample(1028, 0.015)
    cpuTimings := metrics.NewHistogram(s)
    metrics.Register("cpuTimings", cpuTimings)

    s = metrics.NewExpDecaySample(1028, 0.015)
    memoryTimings := metrics.NewHistogram(s)
    metrics.Register("memoryTimings", memoryTimings)

    for i := 0; i < workers; i++ {

        name := fmt.Sprintf("porto.%d", i)

        go func(name string) {
            api := prt.MakePortoApiWithPanic()

            defer func() {
                if err := api.Close(); err != nil {
                    panic(err)
                }
            }()

            err := prt.MakePortoContainer(api, name)
            if err != nil {
                panic(err)
            }

            err = prt.RunPortoContainer(api, name)
            if err != nil {
                panic(err)
            }
            containers = append(containers, name)
            // startTs := time.Now()

            defer prt.CleanupPortoContainer(api, name)

            var done bool
            var count int

            for !done {
                select {
                case done = <-control:
                default:
                    now := time.Now()

                    cpu_usage, _ := prt.GetPortoCpu(api, name)
                    memory_usage, _ := prt.GetPortoMem(api, name)

                    parseTimings.Update(int64(time.Since(now)))

                    cpuTimings.Update(int64(cpu_usage))
                    memoryTimings.Update(int64(memory_usage))

                    count++
                }
            }

            results <- Result{ count }
        } (name)
    }

    t := time.NewTimer(toRun)
    select {
    case <- t.C:
            fmt.Println("completion timer expired")

            var sum int
            for i := 0; i < workers; i++ {
                control <- true
                sum += (<- results).count
            }

            fmt.Printf("total requests: %d, avg time %.2fms\nmax cpu_usage %d, memory_usage %d\n",
                sum,
                parseTimings.Mean() / math.Pow(1000.0,2),
                cpuTimings.Max(),
                memoryTimings.Max(),
            )
    }
}

//
// If container in stop state it returns empty string!
//
func portoProcVec(workers int, toRun time.Duration, nonblock bool, dumpMetrics bool) {
    fmt.Println("Running porto-vec sample")
    defer fmt.Println("done")

    var containers []string

    control := make(chan bool, workers)
    results := make(chan Result, workers)

    parseTimings := metrics.NewTimer()
    // defer parseTimings.Stop()
    metrics.Register("parsePortoTimings", parseTimings)

    // useless for now
    s := metrics.NewExpDecaySample(1028, 0.015)
    cpuTimings := metrics.NewHistogram(s)
    metrics.Register("cpuTimings", cpuTimings)

    s = metrics.NewExpDecaySample(1028, 0.015)
    memoryTimings := metrics.NewHistogram(s)
    metrics.Register("memoryTimings", memoryTimings)

    api := prt.MakePortoApiWithPanic()

    for i := 0; i < workers; i++ {
        name := fmt.Sprintf("porto.%d", i)

        err := prt.MakePortoContainer(api, name)
        if err != nil {
            panic(err)
        }

        err = prt.RunPortoContainer(api, name)
        if err != nil {
            panic(err)
        }
        containers = append(containers, name)
    }

    defer func() {
        for i := 0; i < workers; i++ {
            prt.CleanupPortoContainer(api, containers[i])
        }

        api.Close()
    }()

    var procMetrics func(string, string, uint64)
    if dumpMetrics {
        procMetrics = metricsDump
    } else {
        procMetrics = func(string, string, uint64) {}
    }

    go func() {
        api := prt.MakePortoApiWithPanic()
        defer api.Close()

        metrics := []string {
            "time",
            "cpu_usage",
            "memory_usage",
            "net_tx_bytes",
            "net_rx_bytes",
        }

        var count int
        var done bool

        for !done {
            select {
            case done = <- control:
            default:
                now := time.Now()

                res, err := api.Get3(containers, metrics, nonblock)
                if err != nil {
                    panic(err)
                }

                for cont, props := range res {

                    var cpu_usage uint64
                    var uptime uint64

                    for name, value := range props {

                        // fmt.Printf("%v.%v [%v]\n", cont, name, value.Value)

                        switch name {
                        case "net_tx_bytes", "net_rx_bytes":
                            for _, v := range prt.ParseNetValues(value.Value) {
                                procMetrics(cont, fmt.Sprintf("%s.%s", name, v.Name), v.BytesCount)
                            }
                        default:
                            if v, err := strconv.ParseUint(value.Value, 10, 64); err == nil {
                                procMetrics(cont, name, v)

                                if name == "cpu_usage" {
                                    cpu_usage = v
                                } else if name == "time" {
                                    uptime = v
                                }
                            } // if ...
                        }

                        if uptime > 0 {
                            cpu_load := float64(cpu_usage) / float64(uptime * 1000000000) // uint64(time.Nanosecond))
                            procMetrics(cont, "load", uint64(cpu_load * 100.0))
                        }
                    } // for name, value

                    parseTimings.UpdateSince(now)
                    count ++
                } // for cont, props
            } // select
        } // for !done

        results <- Result{ count }
    }()

    t := time.NewTimer(toRun)
    select {
    case <- t.C:
            fmt.Println("completion timer expired")

            control <- true
            sum := (<- results).count

            fmt.Printf("total requests: %d, avg time %.2fms\nmax cpu_usage %d, memory_usage %d\n",
                sum,
                parseTimings.Mean() / math.Pow(1000.0,2),
                cpuTimings.Max(),
                memoryTimings.Max(),
            )
    }
}

func portoProcReconnect(workers int, toRun time.Duration) {
    fmt.Println("Running porto-reconnect sample")
    defer fmt.Println("done")

    var containers []string

    control := make(chan bool, workers)
    results := make(chan Result, workers)

    s := metrics.NewExpDecaySample(1028, 0.015)
    parseTimings := metrics.NewHistogram(s)
    metrics.Register("parsePortoTimings", parseTimings)

    // useless for now
    s = metrics.NewExpDecaySample(1028, 0.015)
    cpuTimings := metrics.NewHistogram(s)
    metrics.Register("cpuTimings", cpuTimings)

    s = metrics.NewExpDecaySample(1028, 0.015)
    memoryTimings := metrics.NewHistogram(s)
    metrics.Register("memoryTimings", memoryTimings)

    for i := 0; i < workers; i++ {

        name := fmt.Sprintf("porto.%d", i)

        go func(name string) {
            api := prt.MakePortoApiWithPanic()

            err := prt.MakePortoContainer(api, name)
            if err != nil {
                panic(err)
            }

            err = prt.RunPortoContainer(api, name)
            if err != nil {
                panic(err)
            }
            containers = append(containers, name)
            // startTs := time.Now()

            var done bool
            var count int

            for !done {
                select {
                case done = <-control:
                default:
                    now := time.Now()

                    api := prt.MakePortoApiWithPanic()

                    cpu_usage, _ := prt.GetPortoCpu(api, name)
                    memory_usage, _ := prt.GetPortoMem(api, name)

                    if err := api.Close(); err != nil {
                        panic(err)
                    }

                    parseTimings.Update(int64(time.Since(now)))

                    cpuTimings.Update(int64(cpu_usage))
                    memoryTimings.Update(int64(memory_usage))

                    count++
                }
            }

            prt.CleanupPortoContainer(api, name)
            if err := api.Close(); err != nil {
                panic(err)
            }

            results <- Result{ count }
        } (name)
    }

    t := time.NewTimer(toRun)
    select {
    case <- t.C:
            var sum int
            for i := 0; i < workers; i++ {
                control <- true
                sum += (<- results).count
            }

            fmt.Printf("total requests: %d, avg time %.2fms\nmax cpu_usage %d, memory_usage %d\n",
                sum,
                parseTimings.Mean() / math.Pow(1000.0,2),
                cpuTimings.Max(),
                memoryTimings.Max(),
            )
    }
}


func main() {
    workersPtr := flag.Int("workers", workersCount, "conatiners to spawn")
    secToBenchPtr := flag.Uint64("time", toWaitSec, "time to run (seconds)")
    nonBlockPtr := flag.Bool("nonblock", false, "block or not in Porto API Get3 call")
    dumpMetricPtr := flag.Bool("dump", false, "dump petrics on console")

    subSystemPtr := flag.String("iso", "all",
        "which subsystem to test " +
        "(porto|proto-reco|proto-vec|docker-stream|docker|procfs|all)")

    flag.Parse()

    switch *subSystemPtr {
    case "porto-vec":
        portoProcVec(*workersPtr, time.Duration(*secToBenchPtr) * time.Second, *nonBlockPtr, *dumpMetricPtr)
    case "porto-reco":
        portoProcReconnect(*workersPtr, time.Duration(*secToBenchPtr) * time.Second)
    case "porto":
        portoProc(*workersPtr, time.Duration(*secToBenchPtr) * time.Second)
    case "docker":
        dockerProc(*workersPtr, time.Duration(*secToBenchPtr) * time.Second)
    case "docker-stream":
        dockerProcStream(*workersPtr, time.Duration(*secToBenchPtr) * time.Second)
    case "all":
        dockerProc(*workersPtr, time.Duration(*secToBenchPtr) * time.Second)
        dockerProcStream(*workersPtr, time.Duration(*secToBenchPtr) * time.Second)
        portoProc(*workersPtr, time.Duration(*secToBenchPtr) * time.Second)
    }
}
