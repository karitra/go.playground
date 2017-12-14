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
    "porto"
)

const workersCount = 100
const toWaitSec = 30

type Result struct { count int }

func timeTrack(start time.Time, name string) {
    elapsed := time.Since(start)
    fmt.Printf("%s took %s\n", name, elapsed)
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

        stream := dcr.GetDockerStatsStream(ctx, cli, id)
        go func() {
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

    for i := 0; i < workersCount; i++ {
        <- ready
    }

    fmt.Printf("All %d subroutines are ready\n", workersCount)
    close(start)

    t := time.NewTimer(toRun)
    select {
    case <- t.C:
        fmt.Println("completion timer fired")

        sum := 0

        for i := 0; i < workersCount; i++ {
            control <- true
            res := <- results
            sum += res.count
        }

        fmt.Printf("total requests: %d, avg time %.2fms\n", sum, parseTimings.Mean() / math.Pow(1000.0,2))
    }
}

func portoProc(workers int, toRun time.Duration) {
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
            apiCh, errCh := prt.MakePortoApi()

            var api porto.API

            select {
            case api = <- apiCh:
            case err := <- errCh:
                panic(err)
            }

            defer api.Close()

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

                    // fmt.Printf("cpu usage is %d\n", cpu_usage)

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

func main() {
    workersPtr := flag.Int("workers", workersCount, "conatiners to spawn")
    secToBenchPtr := flag.Uint64("time", toWaitSec, "time to run (seconds)")
    subSystemPtr := flag.String("iso", "all", "which subsystem to test (porto|docker|procfs|all)")

    flag.Parse()

    switch *subSystemPtr {
    case "porto":
        portoProc(*workersPtr, time.Duration(*secToBenchPtr) * time.Second)
    case "docker":
        dockerProc(*workersPtr, time.Duration(*secToBenchPtr) * time.Second)
    case "all":
        dockerProc(*workersPtr, time.Duration(*secToBenchPtr) * time.Second)
        portoProc(*workersPtr, time.Duration(*secToBenchPtr) * time.Second)
    }
}
