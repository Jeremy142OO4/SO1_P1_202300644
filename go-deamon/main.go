package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	procSysinfo = "/proc/sysinfo_so1_202300644"
	procContinfo = "/proc/continfo_so1_202300644"

	readEvery  = 3 * time.Second  // lectura del /proc
	cleanEvery = 20 * time.Second // aplicar regla 3 low + 2 high
)

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
	MemoryUsage float64 `json:"Memory_Usage"`
	CPUUsage    float64 `json:"CPU_Usage"`
}

func main() {
	fmt.Println("Daemon iniciado...")

	// 1) Cargar módulos del kernel
	modSys := ModuleManager{
		KoPath:     "/home/jeremy-kvm/Github/SO1_P1_202300644/modulo-kernel/sysinfo/sysinfo.ko",
		ModuleName: "sysinfo",
	}
	if err := modSys.Load(); err != nil {
		fmt.Printf("ERROR cargando sysinfo: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Módulo sysinfo cargado.")

	defer func() {
		if err := modSys.Unload(); err != nil {
			fmt.Printf("WARNING descargando sysinfo: %v\n", err)
		} else {
			fmt.Println("Módulo sysinfo descargado.")
		}
	}()

	modCont := ModuleManager{
		KoPath:     "/home/jeremy-kvm/Github/SO1_P1_202300644/modulo-kernel/continfo/continfo.ko",
		ModuleName: "continfo", // <-- verifica con: lsmod | grep continfo
	}
	if err := modCont.Load(); err != nil {
		fmt.Printf("ERROR cargando continfo: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Módulo continfo cargado.")

	defer func() {
		if err := modCont.Unload(); err != nil {
			fmt.Printf("WARNING descargando continfo: %v\n", err)
		} else {
			fmt.Println("Módulo continfo descargado.")
		}
	}()

	// 2) Instalar cronjob para crear contenedores cada minuto
	cron := CronManager{
		ScriptPath: "/home/jeremy-kvm/Github/SO1_P1_202300644/bash/crear_contenedores.sh",
		LogPath:    "/home/jeremy-kvm/Github/SO1_P1_202300644/bash/crear_contenedores.log",
	}
	if err := cron.InstallEveryMinute(); err != nil {
		fmt.Printf("ERROR instalando cronjob: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Cronjob instalado.")

	defer func() {
		if err := cron.Remove(); err != nil {
			fmt.Printf("WARNING eliminando cronjob: %v\n", err)
		} else {
			fmt.Println("Cronjob eliminado.")
			fmt.Println("Eliminando contenedores del proyecto...")
			RemoveAllProjectContainers()
		}
	}()

	// 3) Verificar que existan los /proc (ya con módulos cargados)
	if _, err := os.Stat(procSysinfo); err != nil {
		fmt.Printf("ERROR: no existe %s (¿PROC_NAME correcto en sysinfo.c?)\n", procSysinfo)
		os.Exit(1)
	}
	if _, err := os.Stat(procContinfo); err != nil {
		fmt.Printf("ERROR: no existe %s (¿PROC_NAME correcto en continfo.c? ¿módulo cargado?)\n", procContinfo)
		os.Exit(1)
	}

	// 4) Manejar cierre limpio (CTRL+C / SIGTERM)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// 5) Tickers
	readTicker := time.NewTicker(readEvery)
	cleanTicker := time.NewTicker(cleanEvery)
	defer readTicker.Stop()
	defer cleanTicker.Stop()

	// limpieza inicial para aplicar regla desde el inicio
	_ = EnforceContainerPolicy()

	for {
		select {
		case <-stop:
			fmt.Println("\nSeñal recibida, cerrando daemon...")
			return

		case <-readTicker.C:
			if err := readAndPrintSysinfo(); err != nil {
				fmt.Printf("WARNING sysinfo: %v\n", err)
			}

		case <-cleanTicker.C:
			if err := EnforceContainerPolicy(); err != nil {
				fmt.Printf("WARNING limpieza: %v\n", err)
			}
		}
	}
}

func readAndPrintSysinfo() error {
	raw, err := os.ReadFile(procSysinfo)
	if err != nil {
		return err
	}

	var si SysInfo
	if err := json.Unmarshal(raw, &si); err != nil {
		return fmt.Errorf("JSON inválido desde %s: %w", procSysinfo, err)
	}

	used := uint64(0)
	if si.Totalram >= si.Freeram {
		used = si.Totalram - si.Freeram
	}

	cpuTotal, err := totalCPUPercent(200 * time.Millisecond)
	if err != nil {
		fmt.Printf("[sysinfo] Total=%dKB Free=%dKB Used=%dKB Procs=%d CPU_Total=N/A (%v)\n",
			si.Totalram, si.Freeram, used, si.Procs, err)
	} else {
		fmt.Printf("[sysinfo] Total=%dKB Free=%dKB Used=%dKB Procs=%d CPU_Total=%.2f%%\n",
			si.Totalram, si.Freeram, used, si.Procs, cpuTotal)
	}

	ci, err := ReadContainerInfo()
	if err != nil {
		fmt.Printf("WARNING continfo: %v\n", err)
		return nil
	}
	PrintContainerInfo(ci)

	return nil
}

func readCPUStat() (idle uint64, total uint64, err error) {
	b, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 0, 0, err
	}
	line := strings.SplitN(string(b), "\n", 2)[0]
	f := strings.Fields(line)
	if len(f) < 8 || f[0] != "cpu" {
		return 0, 0, fmt.Errorf("formato inválido en /proc/stat")
	}

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

	idleAll := vals[3]
	if len(vals) > 4 {
		idleAll += vals[4]
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

func RemoveAllProjectContainers() {
	containers, err := listRunningContainers()
	if err != nil {
		return
	}

	for _, c := range containers {
		if strings.Contains(c.Image, lowImage) ||
			strings.Contains(c.Image, highCPU) ||
			strings.Contains(c.Image, highRAM) {

			_, _ = run("docker", "stop", c.ID)
			_, _ = run("docker", "rm", c.ID)
		}
	}
}
