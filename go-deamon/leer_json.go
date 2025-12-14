package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
)

const procPath = "/proc/sysinfo_so1_202300644"

type SysInfo struct {
	Totalram  uint64    `json:"Totalram"`
	Freeram   uint64    `json:"Freeram"`
	Procs     int       `json:"Procs"`
	Processes []Process `json:"Processes"`
}

type Process struct {
	PID         int     `json:"PID"`
	Name        string  `json:"Name"`
	Cmdline     string  `json:"Cmdline"`
	VSZ         uint64  `json:"vsz"`
	RSS         uint64  `json:"rss"`
	MemoryUsage float64 `json:"Memory_Usage"` // ej: 12.3
	CPUUsage    float64 `json:"CPU_Usage"`    // ej: 1.25
}

func main() {
	raw, err := os.ReadFile(procPath)
	if err != nil {
		fmt.Printf("ERROR leyendo %s: %v\n", procPath, err)
		os.Exit(1)
	}

	// /proc a veces trae cosas raras; limpiamos espacios nulos (por si acaso)
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
	fmt.Printf("Procesos contados: %d (array=%d)\n", si.Procs, len(si.Processes))
	fmt.Println()

	// Top 5 por RSS (memoria)
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

	// Top 5 por CPU
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

	// Validaciones simples
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
	// Quita saltos y tabs para que el print no se rompa
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