package common

import (
    "testing"
)

func RunParallel(b *testing.B, task func()) {
    b.ReportAllocs()
    b.ResetTimer()

    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            task()
        }
    })
}
