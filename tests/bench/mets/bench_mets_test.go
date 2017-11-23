package bench_cgroups

import (
    "testing"
    "fmt"
    "github.com/rcrowley/go-metrics"
    // "common"
)

func init() {
    fmt.Println("Starting benchs")
}

func makeMetrics(size int) (m metrics.Histogram) {
    s := metrics.NewExpDecaySample(size, 0.015)
    m = metrics.NewHistogram(s)

    metrics.Register("CPU", m)

    for i := 0; i < size; i++ {
        m.Update(int64(i * i))
    }

    return
}

func BenchmarkDecayMetrics1k(b *testing.B) {
    h := makeMetrics(1024)
    common.RunParallel(b, h, func() {
        h.Update(int64(100500))
    })
}

func BenchmarkDecayMetrics128(b *testing.B) {
    h := makeMetrics(128)
    common.RunParallel(b, h, func() {
        h.Update(int64(100500))
    })
}

func BenchmarkDecayMetricsMean1k(b *testing.B) {
    h := makeMetrics(1024)
    common.RunParallel(b, h, func() {
        h.Mean()
    })
}

func BenchmarkDecayMetricsMean128(b *testing.B) {
    h := makeMetrics(128)
    common.RunParallel(b, h, func() {
        h.Mean()
    })
}

func BenchmarkDecayMetricsMean16(b *testing.B) {
    h := makeMetrics(16)
    common.RunParallel(b, h, func() {
        h.Mean()
    })
}

func BenchmarkDecayMetricsMax1k(b *testing.B) {
    h := makeMetrics(1024)
    common.RunParallel(b, h, func() {
        h.Max()
    })
}

func BenchmarkDecayMetricsMax128(b *testing.B) {
    h := makeMetrics(128)
    common.RunParallel(b, h, func() {
        h.Max()
    })
}

func BenchmarkDecayMetricsMax16(b *testing.B) {
    h := makeMetrics(16)
    common.RunParallel(b, h, func() {
        h.Max()
    })
}
