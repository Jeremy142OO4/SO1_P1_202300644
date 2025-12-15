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
	procSysinfo  = "/proc/sysinfo_so1_202300644"
	

	loopEvery = 20 * time.Second 
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
	UTime       uint64  `json:"utime"`
	STime       uint64  `json:"stime"`
}



func main() {
	fmt.Println("Daemon iniciado...")

	// 0) DB
	if err := InitDB(); err != nil {
		fmt.Printf("ERROR InitDB: %v\n", err)
		os.Exit(1)
	}
	defer CloseDB()

	if err := ResetDB(); err != nil {
		fmt.Printf("ERROR ResetDB: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("DB inicializada y reseteada.")

	// 1) Módulos kernel
	loadModules()
	defer unloadModules()

	// 2) Cron
	cron := CronManager{
		ScriptPath: "/home/jeremy-kvm/Github/SO1_P1_202300644/bash/crear_contenedores.sh",
		LogPath:    "/home/jeremy-kvm/Github/SO1_P1_202300644/bash/crear_contenedores.log",
	}
	if err := cron.InstallEveryMinute(); err != nil {
		fmt.Printf("ERROR cron: %v\n", err)
		os.Exit(1)
	}
	defer cron.Remove()

	// 3) Verificar /proc
	checkProc()

	// 4) Señales
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// 5) LOOP PRINCIPAL
	ticker := time.NewTicker(loopEvery)
	defer ticker.Stop()

	fmt.Println("Loop principal iniciado (cada 20s)")

	for {
		select {
		case <-stop:
			fmt.Println("\nCerrando daemon...")
			removeStoppedProjectContainers()
			return

		case <-ticker.C:
			if err := loopOnce(); err != nil {
				fmt.Printf("WARNING loop: %v\n", err)
			}
		}
	}
}



func loopOnce() error {
	
	raw, err := os.ReadFile(procSysinfo)
	if err != nil {
		return err
	}

	var si SysInfo
	if err := json.Unmarshal(raw, &si); err != nil {
		return err
	}

	used := si.Totalram - si.Freeram
	cpuTotal, _ := totalCPUPercent(200 * time.Millisecond)

	fmt.Printf(
		"[sysinfo] Total=%dKB Free=%dKB Used=%dKB Procs=%d CPU=%.2f%%\n",
		si.Totalram, si.Freeram, used, si.Procs, cpuTotal,
	)


	ci, err := ReadContainerInfo()
	if err != nil {
		fmt.Printf("WARNING continfo: %v\n", err)
		ci = nil
	}

	
	idLote, err := CrearLote()
	if err != nil {
		return err
	}

	if err := InsertarProcesosSnapshot(idLote, si.Processes); err != nil {
		return err
	}

	if ci != nil {
		if err := InsertarContenedoresSnapshot(idLote, ci); err != nil {
			return err
		}
	}

	
	if err := EnforceContainerPolicy(); err != nil {
		return err
	}

	fmt.Printf(
		"[LOTE %d] procesos=%d contenedores=%d\n",
		idLote,
		len(si.Processes),
		func() int {
			if ci == nil {
				return 0
			}
			return len(ci.Containers)
		}(),
	)

	return nil
}


func readCPUStat() (idle, total uint64, err error) {
	b, err := os.ReadFile("/proc/stat")
	if err != nil {
		return
	}
	f := strings.Fields(strings.SplitN(string(b), "\n", 2)[0])

	var vals []uint64
	for _, v := range f[1:] {
		n, e := strconv.ParseUint(v, 10, 64)
		if e != nil {
			break
		}
		vals = append(vals, n)
	}

	idle = vals[3] + vals[4]
	for _, v := range vals {
		total += v
	}
	return
}

func totalCPUPercent(sample time.Duration) (float64, error) {
	i1, t1, _ := readCPUStat()
	time.Sleep(sample)
	i2, t2, _ := readCPUStat()

	dt := float64(t2 - t1)
	di := float64(i2 - i1)

	return (dt-di)/dt*100.0, nil
}


func checkProc() {
	if _, err := os.Stat(procSysinfo); err != nil {
		fmt.Println("ERROR:", procSysinfo)
		os.Exit(1)
	}
	if _, err := os.Stat(procContinfo); err != nil {
		fmt.Println("ERROR:", procContinfo)
		os.Exit(1)
	}
}

func loadModules() {
	ModuleManager{
		KoPath:     "/home/jeremy-kvm/Github/SO1_P1_202300644/modulo-kernel/sysinfo/sysinfo.ko",
		ModuleName: "sysinfo",
	}.Load()

	ModuleManager{
		KoPath:     "/home/jeremy-kvm/Github/SO1_P1_202300644/modulo-kernel/continfo/continfo.ko",
		ModuleName: "continfo",
	}.Load()
}

func unloadModules() {
	ModuleManager{ModuleName: "continfo"}.Unload()
	ModuleManager{ModuleName: "sysinfo"}.Unload()
}
