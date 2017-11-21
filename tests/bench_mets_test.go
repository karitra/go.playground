package bench_mets_test

import (
    "testing"
    "github.com/rcrowley/go-metrics"
)

func makeMetrics(size int) (m metrics.Histogram) {
    s := metrics.NewExpDecaySample(size, 0.015)
    m = metrics.NewHistogram(s)

    metrics.Register("CPU", m)

    for i := 0; i < size; i++ {
        m.Update(int64(i * i))
    }

    return
}

func runParallel(b *testing.B, h metrics.Histogram, task func()) {
    b.ReportAllocs()
    b.ResetTimer()

    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            task()
        }
    })
}

func BenchmarkDecayMetrics1k(b *testing.B) {
    h := makeMetrics(1024)
    runParallel(b, h, func() {
        h.Update(int64(100500))
    })
}

func BenchmarkDecayMetrics128(b *testing.B) {
    h := makeMetrics(128)
    runParallel(b, h, func() {
        h.Update(int64(100500))
    })
}

func BenchmarkDecayMetricsMean1k(b *testing.B) {
    h := makeMetrics(1024)
    runParallel(b, h, func() {
        h.Mean()
    })
}

func BenchmarkDecayMetricsMean128(b *testing.B) {
    h := makeMetrics(128)
    runParallel(b, h, func() {
        h.Mean()
    })
}

func BenchmarkDecayMetricsMax1k(b *testing.B) {
    h := makeMetrics(1024)
    runParallel(b, h, func() {
        h.Max()
    })
}

func BenchmarkDecayMetricsMax128(b *testing.B) {
    h := makeMetrics(128)
    runParallel(b, h, func() {
        h.Max()
    })
}
