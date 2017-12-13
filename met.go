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
    metrics.Register("parseTimings", parseTimings)

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
    api, err := prt.MakePortoApi()
    if err != nil {
        panic(err)
    }

    defer api.Close()

    var containers []string

    control := make(chan bool, workers)

    for i := 0; i < workers; i++ {
        var err error

        name := fmt.Sprintf("porto.%d", i)

        err = prt.MakePortoContainer(api, name)
        if err != nil {
            panic(err)
        }

        defer prt.CleanupPortoContainer(api, name)

        err = prt.RunPortoContainer(api, name)
        if err != nil {
            panic(err)
        }
        containers = append(containers, name)

        go func() {
            var done bool
            for !done {
                // switch {
                //     case done = <- control:
                //         break
                // }
            }
        }()
    }

    t := time.NewTimer(toRun)
    select {
    case <- t.C:
            fmt.Println("completion timer expired")
            for i := 0; i < workers; i++ {
                control <- true
            }
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
