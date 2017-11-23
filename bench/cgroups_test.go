package bench

import (
    // "github.com/docker/docker/client"
    // "github.com/yandex/porto/blob/master/src/api/go"
    // "fmt"
    "testing"
    "bench.met/bench/common"
)

func BenchmarkDecayMetrics1k(b *testing.B) {
    common.RunParallel(b, func() {
        _ := 1 + 1
    })
}
