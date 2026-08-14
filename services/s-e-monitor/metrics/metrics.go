//go:build linux

// Package metrics liest Systemmetriken (CPU-Auslastung, RAM) aus /proc.
package metrics

// Zuletzt geändert: 2026-08-14

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Snapshot ist ein gemessener Systemzustand.
type Snapshot struct {
	CPU     float64
	UsedMB  float64
	TotalMB float64
	Time    time.Time
}

// Params liefert die params für event.system.metrics (Schema-konform).
func (s Snapshot) Params() map[string]any {
	return map[string]any{
		"cpu": s.CPU,
		"ram": map[string]any{"used_mb": s.UsedMB, "total_mb": s.TotalMB},
	}
}

// Reader misst CPU und RAM über aufeinanderfolgende /proc-Werte.
type Reader struct {
	prevTotal uint64
	prevIdle  uint64
	first     bool
}

// NewReader erstellt einen Reader; der erste Aufruf liefert CPU=0 (Basiswert).
func NewReader() *Reader {
	return &Reader{first: true}
}

// Read misst die aktuellen Metriken.
func (r *Reader) Read() (Snapshot, error) {
	total, idle, err := readCPU()
	if err != nil {
		return Snapshot{}, err
	}
	var cpu float64
	if !r.first && total > r.prevTotal {
		dt := total - r.prevTotal
		di := idle - r.prevIdle
		if dt > 0 {
			cpu = (1 - float64(di)/float64(dt)) * 100
		}
	}
	r.prevTotal, r.prevIdle, r.first = total, idle, false

	totalKB, availKB, err := readMem()
	if err != nil {
		return Snapshot{}, err
	}
	return Snapshot{
		CPU:     cpu,
		TotalMB: totalKB / 1024,
		UsedMB:  (totalKB - availKB) / 1024,
		Time:    time.Now().UTC(),
	}, nil
}

func readCPU() (total, idle uint64, err error) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 0, 0, err
	}
	total, idle = parseCPU(string(data))
	return total, idle, nil
}

func parseCPU(data string) (total, idle uint64) {
	for _, line := range strings.Split(data, "\n") {
		if !strings.HasPrefix(line, "cpu ") {
			continue
		}
		for i, f := range strings.Fields(line)[1:] {
			v, _ := strconv.ParseUint(f, 10, 64)
			total += v
			if i == 3 || i == 4 { // idle, iowait
				idle += v
			}
		}
		return
	}
	return 0, 0
}

func readMem() (totalKB, availKB float64, err error) {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, 0, err
	}
	totalKB, availKB = parseMeminfo(string(data))
	return totalKB, availKB, nil
}

func parseMeminfo(data string) (totalKB, availKB float64) {
	for _, line := range strings.Split(data, "\n") {
		switch {
		case strings.HasPrefix(line, "MemTotal:"):
			totalKB, _ = strconv.ParseFloat(strings.Fields(line)[1], 64)
		case strings.HasPrefix(line, "MemAvailable:"):
			availKB, _ = strconv.ParseFloat(strings.Fields(line)[1], 64)
		}
	}
	return totalKB, availKB
}
