package bench

import (
	"time"

	"github.com/klauspost/cpuid/v2"
)

type Stats struct {
	Time         time.Time `json:"time"`
	CPUBrandName string    `json:"cpu_brand_name"`
	LogicalCores int       `json:"logical_cores"`
	HasAESNI     bool      `json:"has_aesni"`
}

func FetchStats() Stats {
	return Stats{
		Time:         time.Now(),
		CPUBrandName: cpuid.CPU.BrandName,
		LogicalCores: cpuid.CPU.LogicalCores,
		HasAESNI:     cpuid.CPU.Supports(cpuid.AESNI),
	}
}
