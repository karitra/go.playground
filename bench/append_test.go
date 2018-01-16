package bench

import "testing"

func BenchmarkPreallocAppend100(b *testing.B) {
    slice := make([]int, 0, 100)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        slice = append(slice, i)
    }
}

func BenchmarkPreallocAppend1k(b *testing.B) {
    slice := make([]int, 0, 1000)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        slice = append(slice, i)
    }
}

func BenchmarkPreallocAppend10k(b *testing.B) {
    slice := make([]int, 0, 10000)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        slice = append(slice, i)
    }
}

func BenchmarkPreallocAppend100k(b *testing.B) {
    slice := make([]int, 0, 100000)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        slice = append(slice, i)
    }
}

func BenchmarkAppend(b *testing.B) {
    var slice []int

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        slice = append(slice, i)
    }
}


