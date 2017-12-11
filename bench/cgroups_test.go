package bench

import (

    docker "github.com/docker/docker/client"

    "github.com/docker/docker/api/types"
    "github.com/docker/docker/api/types/container"

    "porto" // local link to "github.com/yandex/porto/src/api/go/porto"
    "testing"
    "fmt"
    "bench.met/bench/common"
    "os"
    "bytes"
    //"time"
    // "strconv"

    "encoding/json"
    // "log"
    "golang.org/x/net/context"
    "io/ioutil"

    //"github.com/docker/docker/daemon/stats"
    )


const (
    //localVersion = "1.30"
    containerPorto = "dork.porto"
    containerDocker = "dork.docker"

    dockerImage = "ubuntu"
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

func makeDockerClient(ctx context.Context) (cli *docker.Client, id string) {
    cli, err := docker.NewEnvClient()
    if err != nil {
        panic(err)
    }

    conf := container.Config{}

    conf.Image = dockerImage
    conf.Tty = true

    hostConf := container.HostConfig{}
    hostConf.AutoRemove = true

    cont, err := cli.ContainerCreate(ctx, &conf, &hostConf, nil, containerDocker)
    if err != nil {
        panic(err)
    }

    fmt.Println("start")
    id = cont.ID
    options := types.ContainerStartOptions{}
    cli.ContainerStart(ctx, id, options)

    return
}

func cleanupDockerContainer(ctx context.Context, cli *docker.Client, id string) {
    if err := cli.ContainerKill(ctx, id, "KILL"); err != nil {
        panic(err)
    }
}


func getDockerStats(ctx context.Context, client *docker.Client, id string, stream bool) (out types.Stats) {
    stats, err := client.ContainerStats(ctx, id, stream)
    if err != nil {
        panic(err)
    }

    // fmt.Printf("stats %V\n", stats.Body)
    defer stats.Body.Close()

    dec := json.NewDecoder(stats.Body)
    if err := dec.Decode(out); err != nil {
        fmt.Printf("error %v\n", err)
    }

    return
}


func BenchmarkDockerMetrics(b *testing.B) {
    ctx := context.Background()

    cli, id := makeDockerClient(ctx)
    defer cleanupDockerContainer(ctx, cli, id)

    b.ResetTimer()

    b.Run("sequental", func (b *testing.B) {
        for i := 0; i < b.N; i++ {
            getDockerStats(ctx, cli, containerDocker, false)
        }
    })

    b.StopTimer()
}


// TODO: broken
// func BenchmarkDockerMetricsPar(b *testing.B) {
//     ctx := context.Background()
//
//     cli, id := makeDockerClient(ctx)
//     defer cleanupDockerContainer(ctx, cli, id)
//
//     b.ResetTimer()
//     b.RunParallel(func(pb *testing.PB) {
//         for pb.Next() {
//             getDockerStats(ctx, cli, containerDocker, false)
//         }
//     })
//
//     b.StopTimer()
// }


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
        _, err := ioutil.ReadFile("/proc/meminfo")
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
