package bench

import (
    _ "github.com/docker/docker/client"
    "porto" // local link to "github.com/yandex/porto/src/api/go/porto"
    "testing"
    "fmt"
    "bench.met/bench/common"
    "os"
    "bytes"
    // "strconv"
)

var (
    containerPorto = "dork.porto"
    containerDocker = "dork.docker"

    portoProps = []string{
        "cpu_usage",
        "memory_usage",
    }

    portoContainers = []string{containerPorto}
)

func dumpPortoResult(r *map[string]map[string]porto.TPortoGetResponse) {
    // return

    for _, v := range *r {
        for _, value := range v {
            if value.Error != 0 {
                fmt.Printf("response error [%d] %s\n", value.Error, value.ErrorMsg)
            }
            // else {
            //     fmt.Printf("%s.%s %s\n", app, metric, value.Value)
            // }
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


func benchmarkPortoParallel(b *testing.B) {
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

func BenchmarkPortoPlain(b *testing.B) {
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

func BenchmarkProcFsPlain(b *testing.B) {

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
