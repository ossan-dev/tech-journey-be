package utils

import (
	"fmt"
	"runtime"
)

func PrintMemStats() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Println()
	fmt.Printf("Alloc = %v MB\n", BytesToMegaBytes(m.Alloc))
	fmt.Printf("TotalAlloc = %v MB\n", BytesToMegaBytes(m.TotalAlloc))
	fmt.Printf("Sys = %v MB\n", BytesToMegaBytes(m.Sys))
	fmt.Printf("NumGC = %v\n", m.NumGC)
}

func BytesToMegaBytes(b uint64) uint64 {
	return b / 1000 / 1000
}
