package main

import (
    "golang.org/x/net/context"
    "fmt"
    "time"
    "math"
    dcr "bench.met/bench"
    "encoding/json"
    "flag"
    "github.com/rcrowley/go-metrics"
)

const workersCount = 100
const toWaitSec = 30 * time.Second

type Result struct { count int }

func timeTrack(start time.Time, name string) {
    elapsed := time.Since(start)
    fmt.Printf("%s took %s\n", name, elapsed)
}

func dockerProc(workers int, toRun uint64) {
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

    t := time.NewTimer(time.Duration(toRun))
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

func main() {
    workersPtr := flag.Int("workers", workersCount, "conatiners to spawn")
    secToBenchPtr := flag.Uint64("time", uint64(toWaitSec), "time to run (seconds)")
    flag.Parse()

    dockerProc(*workersPtr, *secToBenchPtr)
}
