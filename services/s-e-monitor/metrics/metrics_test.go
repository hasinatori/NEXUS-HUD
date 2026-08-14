//go:build linux

package metrics

// Zuletzt geändert: 2026-08-14

import "testing"

func TestParseCPU(t *testing.T) {
	data := `cpu  100 10 20 30 40 5 6 7 8 9
cpu0 10 1 2 3 4 0 0 0 0 0
`
	total, idle := parseCPU(data)
	// Summe: 100+10+20+30+40+5+6+7+8+9 = 235; idle+iowait = 30+40 = 70
	if total != 235 || idle != 70 {
		t.Fatalf("total=%d idle=%d, want 235/70", total, idle)
	}
}

func TestParseMeminfo(t *testing.T) {
	data := `MemTotal:       16384152 kB
MemFree:          102400 kB
MemAvailable:    5000000 kB
Buffers:           1024 kB
`
	totalKB, availKB := parseMeminfo(data)
	if totalKB != 16384152 || availKB != 5000000 {
		t.Fatalf("totalKB=%v availKB=%v, want 16384152/5000000", totalKB, availKB)
	}
}

func TestReaderDeltas(t *testing.T) {
	r := NewReader()
	first, err := r.Read()
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if first.CPU != 0 {
		t.Fatalf("erster Read: CPU=%v, want 0", first.CPU)
	}
	// Ein zweiter Read innerhalb derselben Sekunde liefert reale Werte >= 0.
	second, err := r.Read()
	if err != nil {
		t.Fatalf("Read 2: %v", err)
	}
	if second.CPU < 0 || second.CPU > 100 {
		t.Fatalf("CPU=%v außerhalb [0,100]", second.CPU)
	}
	if second.TotalMB <= 0 {
		t.Fatalf("TotalMB=%v, want > 0", second.TotalMB)
	}
	if second.UsedMB < 0 || second.UsedMB > second.TotalMB {
		t.Fatalf("UsedMB=%v außerhalb [0,%v]", second.UsedMB, second.TotalMB)
	}
}
