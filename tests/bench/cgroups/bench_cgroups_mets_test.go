package bench_cgroups

import (
    // "github.com/docker/docker/client"
    // "github.com/yandex/porto/blob/master/src/api/go"
    "fmt"
    "common"
)

func BenchmarkDecayMetrics1k(b *testing.B) {
    common.RunParallel(b, func() {
        _ := 1 + 1
    })
}
