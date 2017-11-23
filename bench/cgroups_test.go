package bench

import (
    _ "github.com/docker/docker/client"
    "porto" // local link to "github.com/yandex/porto/src/api/go/porto"
    "testing"
    "fmt"
    "bench.met/bench/common"
)

var (
    containerPorto = "dork.porto"
    containerDocker = "dork.docker"
)


func dumpPortoResult(r *map[string]map[string]porto.TPortoGetResponse) {
    for app, v := range *r {
        for metric, value := range v {
            if value.Error != 0 {
                fmt.Printf("response error [%d] %s\n", value.Error, value.ErrorMsg)
            } else {
                fmt.Printf("%s.%s %s\n", app, metric, value.Value)
            }
        }
    }
}

func BenchmarkPorto(b *testing.B) {

    containers := []string{containerPorto}

    props := []string{
        "cpu_usage",
        "memory_usage",
    }

    api, err := porto.Connect()
    if err != nil {
        panic(err)
    }

    common.RunParallel(b, func() {
        resp, err := api.Get(containers, props)
        if err != nil {
            fmt.Printf("failed to get metrics %v\n", err)
            b.FailNow()
        }
        dumpPortoResult(&resp)
    })
}
