package bench

import (
    "os"
    "testing"
    "fmt"

    psu "github.com/shirou/gopsutil/process"
)

func BenchmarkGopsCpuStat(b *testing.B) {
    selfPid := os.Getpid()

    p, err := psu.NewProcess(int32(selfPid))
    if err != nil {
        fmt.Printf("failed get process struct for pid %v", selfPid)
        return
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        p.CPUPercent()
    }
}

func BenchmarkGopsMemStat(b *testing.B) {
    selfPid := os.Getpid()

    p, err := psu.NewProcess(int32(selfPid))
    if err != nil {
        fmt.Printf("failed get process struct for pid %v", selfPid)
        return
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        p.MemoryInfo()
    }
}

func BenchmarkGopsAllStat(b *testing.B) {
    selfPid := os.Getpid()

    m := make(map[string]string, 0)

    m["z"] = "y"

    p, err := psu.NewProcess(int32(selfPid))
    if err != nil {
        fmt.Printf("failed get process struct for pid %v", selfPid)
        return
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        p.MemoryInfo()
        p.CPUPercent()
        n, _ := p.NetIOCounters(true)
        for _, nn := range n {
            fmt.Printf("%v\n", nn)
        }
    }
}
