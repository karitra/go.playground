package bench

import (
    "porto" // local link to "github.com/yandex/porto/src/api/go/porto"
    "testing"
    "fmt"
    "bench.met/bench/common"
    "os"
    "bytes"

    "encoding/json"
    "golang.org/x/net/context"
    "io/ioutil"
)


const (
    //localVersion = "1.30"
    containerPorto = "dork.porto"
    containerDocker = "dork.docker"
    )

var (
    portoProps = []string{
        "cpu_usage",
        "memory_usage",
    }

    portoContainers = []string{containerPorto}
)

func dumpPortoResult(r *map[string]map[string]porto.TPortoGetResponse) {
    for _, v := range *r {
        for _, value := range v {
            if value.Error != 0 {
                fmt.Printf("response error [%d] %s\n", value.Error, value.ErrorMsg)
            }
        }
    }
}

func makePortoApi() (api porto.API) {
    var err error
    api, err = porto.Connect()
    if err != nil {
        panic(err)
    }

    if err := api.Create(containerPorto); err != nil {
        panic(err)
    }

    if err := api.SetProperty(containerPorto, "command", "bash"); err != nil {
        panic(err)
    }

    if err := api.Start(containerPorto); err != nil {
        panic(err)
    }

    return
}

func clearPorto(api porto.API) {
    if err := api.Stop(containerPorto); err != nil {
        panic(err)
    }

    if err := api.Destroy(containerPorto); err != nil {
        panic(err)
    }
}


func benchmarkPortoMetricsPar(b *testing.B) {
    api := makePortoApi()

    defer api.Close()
    defer clearPorto(api)

    common.RunParallel(b, func() {
        resp, err := api.Get(portoContainers, portoProps)
        if err != nil {
            fmt.Printf("failed to get metrics %v\n", err)
            b.FailNow()
        }
        dumpPortoResult(&resp)
    })

    fmt.Println("I'm done")
}

func BenchmarkPortoMetrics(b *testing.B) {
    api := makePortoApi()

    defer api.Close()
    defer clearPorto(api)

    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        resp, err := api.Get(portoContainers, portoProps)
        if err != nil {
            fmt.Printf("failed to get metrics %v\n", err)
            b.FailNow()
        }
        dumpPortoResult(&resp)
    }
}

func dummyAtoi(raw []byte) (acc int) {
    for _, ch := range raw {
        if ch == ' ' { break }
        acc = acc * 10 + int(ch - '0')
    }
    return
}

func scanMetrics(raw []byte) (utime, ctime int) {
    mets := bytes.Split(raw, []byte(" "))

    // utime, _ = strconv.Atoi(string(mets[13]))
    // ctime, _ = strconv.Atoi(string(mets[14]))

    utime = dummyAtoi(mets[13])
    ctime = dummyAtoi(mets[14])

    // fmt.Printf("utime [%s], ctime [%s]\n", mets[13], mets[14])
    // fmt.Printf("utime %d, ctime %d\n", utime, ctime)

    return
}

func BenchmarkProcMetrics(b *testing.B) {
    pid := os.Getpid()
    pageSize := os.Getpagesize()

    statFile, err := os.Open(fmt.Sprintf("/proc/%d/stat", pid))

    if err != nil {
        panic(err)
    }

    buf := make([]byte, pageSize)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := statFile.Read(buf)
        if err != nil {
            fmt.Printf("failed to read file %s %v", statFile, err)
            b.FailNow()
        }

        scanMetrics(buf)
        statFile.Seek(0, 0)
    }
}

func BenchmarkDockerMetrics(b *testing.B) {
    ctx := context.Background()

    cli, _ := MakeDockerClient()
    defer cli.Close()

    id, _ := MakeDockerContainer(ctx, cli, containerDocker)
    defer CleanupDockerContainer(ctx, cli, id)

    b.ResetTimer()
    b.Run("sequental", func (b *testing.B) {
        for i := 0; i < b.N; i++ {
            GetDockerStats(ctx, cli, containerDocker, false)
        }
    })
    b.StopTimer()
}


func BenchmarkDockerMetricsStream(b *testing.B) {
    ctx := context.Background()

    cli, _ := MakeDockerClient()
    defer cli.Close()

    id, _ := MakeDockerContainer(ctx, cli, containerDocker)
    defer CleanupDockerContainer(ctx, cli, id)

    stream := GetDockerStatsStream(ctx, cli, id)
    defer stream.Close()

    dec := json.NewDecoder(stream)
    ParseDockerStats(dec)

    b.ResetTimer()
    b.Run("sequental", func (b *testing.B) {
        for i := 0; i < b.N; i++ {
            ParseDockerStats(dec)
        }
    })

    b.StopTimer()
}


func BenchmarkDockerMetricsPar(b *testing.B) {
    ctx := context.Background()

    cli, _ := MakeDockerClient()
    defer cli.Close()

    id, _ := MakeDockerContainer(ctx, cli, containerDocker)
    defer CleanupDockerContainer(ctx, cli, id)

    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            GetDockerStats(ctx, cli, id, false)
        }
    })

    b.StopTimer()
}


func BenchmarkReadProcFile(b *testing.B) {
    for i := 0; i < b.N; i++ {
        _, err := ioutil.ReadFile("/proc/meminfo")
        if err != nil {
            panic(err)
        }
    }
}

func BenchmarkReadSysFsFile(b *testing.B) {
    for i := 0; i < b.N; i++ {
        _, err := ioutil.ReadFile("/sys/fs/cgroup/memory/memory.stat")
        if err != nil {
            panic(err)
        }
    }
}


func BenchmarkOpenClose(b *testing.B) {
    for i := 0; i < b.N; i++ {
        func () {
            statsFile, err := os.Open("/proc/meminfo")
            defer statsFile.Close()
            if err != nil {
                panic(err)
            }
        }()
    }
}
