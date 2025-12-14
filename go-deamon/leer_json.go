package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"time"
	"strings"
	"strconv"
)

const procPath = "/proc/sysinfo_so1_202300644"

type SysInfo struct {
	Totalram   uint64    `json:"Totalram"`
	Freeram    uint64    `json:"Freeram"`
	Procs      int       `json:"Procs"`
	Processes  []Process `json:"Processes"`
}

type Process struct {
	PID         int     `json:"PID"`
	Name        string  `json:"Name"`
	Cmdline     string  `json:"Cmdline"`
	VSZ         uint64  `json:"vsz"`
	RSS         uint64  `json:"rss"`
	MemoryUsage float64 `json:"Memory_Usage"`
	CPUUsage    float64 `json:"CPU_Usage"` // si lo sigues usando
	UTime       uint64  `json:"utime"`
	STime       uint64  `json:"stime"`
	CPUPercent float64 `json:"-"`
}

func main() {
	raw, err := os.ReadFile(procPath)
	if err != nil {
		fmt.Printf("ERROR leyendo %s: %v\n", procPath, err)
		os.Exit(1)
	}


	clean := strings.TrimSpace(string(raw))

	var si SysInfo
	if err := json.Unmarshal([]byte(clean), &si); err != nil {
		fmt.Printf("ERROR: el contenido NO se pudo parsear como JSON válido.\n")
		fmt.Printf("Causa: %v\n\n", err)
		fmt.Println("TIP: revisa si Cmdline/Name tienen comillas sin escapar en el módulo.")
		fmt.Println("Contenido (primeros 800 chars):")
		fmt.Println(snippet(clean, 800))
		os.Exit(1)
	}

	printSummary(si)
}

func printSummary(si SysInfo) {
	used := uint64(0)
	if si.Totalram >= si.Freeram {
		used = si.Totalram - si.Freeram
	}

	fmt.Println("=== SYSINFO (/proc/sysinfo_so1_202300644) ===")
	fmt.Printf("Total RAM (KB): %d\n", si.Totalram)
	fmt.Printf("Free  RAM (KB): %d\n", si.Freeram)
	fmt.Printf("Used  RAM (KB): %d\n", used)
	cpuTotal, err := totalCPUPercent(500 * time.Millisecond)
	if err != nil {
		fmt.Printf("CPU Total (%%): N/A (%v)\n", err)
	} else {
		fmt.Printf("CPU Total (%%): %.2f\n", cpuTotal)
	}
	fmt.Printf("Procesos contados: %d (array=%d)\n", si.Procs, len(si.Processes))
	fmt.Println()


	topMem := make([]Process, 0, len(si.Processes))
	topMem = append(topMem, si.Processes...)
	sort.Slice(topMem, func(i, j int) bool {
		return topMem[i].RSS > topMem[j].RSS
	})

	fmt.Println("Top 5 por RAM (RSS KB):")
	for i := 0; i < min(5, len(topMem)); i++ {
		p := topMem[i]
		fmt.Printf("  #%d PID=%d Name=%s RSS=%dKB VSZ=%dKB Mem%%=%.1f Cmd=%s\n",
			i+1, p.PID, p.Name, p.RSS, p.VSZ, p.MemoryUsage, safeOneLine(p.Cmdline, 60))
	}
	fmt.Println()

	
	topCPU := make([]Process, 0, len(si.Processes))
	topCPU = append(topCPU, si.Processes...)
	sort.Slice(topCPU, func(i, j int) bool {
		return topCPU[i].CPUUsage > topCPU[j].CPUUsage
	})

	fmt.Println("Top 5 por CPU (%):")
	for i := 0; i < min(5, len(topCPU)); i++ {
		p := topCPU[i]
		fmt.Printf("  #%d PID=%d Name=%s CPU%%=%.2f RSS=%dKB Cmd=%s\n",
			i+1, p.PID, p.Name, p.CPUUsage, p.RSS, safeOneLine(p.Cmdline, 60))
	}

	
	if err := validate(si); err != nil {
		fmt.Println()
		fmt.Printf("WARNING: %v\n", err)
	}
}

func validate(si SysInfo) error {
	if si.Totalram == 0 {
		return errors.New("Totalram=0 (¿módulo devolvió valores correctos?)")
	}
	if si.Procs < 0 {
		return errors.New("Procs negativo (no debería pasar)")
	}
	return nil
}

func snippet(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func safeOneLine(s string, max int) string {
	
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	s = strings.TrimSpace(s)

	if max > 0 && len(s) > max {
		return s[:max] + "..."
	}
	if s == "" {
		return "N/A"
	}
	return s
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func readCPUStat() (idle uint64, total uint64, err error) {
	b, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 0, 0, err
	}
	line := strings.SplitN(string(b), "\n", 2)[0] // primera línea: "cpu ..."
	f := strings.Fields(line)
	if len(f) < 8 || f[0] != "cpu" {
		return 0, 0, fmt.Errorf("formato inválido en /proc/stat")
	}

	// cpu user nice system idle iowait irq softirq steal guest guest_nice...
	var vals []uint64
	for i := 1; i < len(f); i++ {
		v, e := strconv.ParseUint(f[i], 10, 64)
		if e != nil {
			break
		}
		vals = append(vals, v)
	}
	if len(vals) < 5 {
		return 0, 0, fmt.Errorf("no hay suficientes campos en /proc/stat")
	}

	idleAll := vals[3] // idle
	if len(vals) > 4 {
		idleAll += vals[4] // iowait
	}

	var tot uint64
	for _, v := range vals {
		tot += v
	}

	return idleAll, tot, nil
}

func totalCPUPercent(sample time.Duration) (float64, error) {
	idle1, total1, err := readCPUStat()
	if err != nil {
		return 0, err
	}
	time.Sleep(sample)
	idle2, total2, err := readCPUStat()
	if err != nil {
		return 0, err
	}

	if total2 <= total1 {
		return 0, fmt.Errorf("delta total CPU inválido")
	}

	dTotal := float64(total2 - total1)
	dIdle := float64(idle2 - idle1)

	used := (dTotal - dIdle) / dTotal * 100.0
	if used < 0 {
		used = 0
	}
	if used > 100 {
		used = 100
	}
	return used, nil
}