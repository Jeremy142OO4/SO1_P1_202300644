package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	procSysinfo = "/proc/sysinfo_so1_202300644"

	readEvery = 3 * time.Second

	
	cleanEvery = 20 * time.Second
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


	mod := ModuleManager{
		KoPath:     "/home/jeremy-kvm/Github/SO1_P1_202300644/modulo-kernel/sysinfo.ko",
		ModuleName: "sysinfo", 
	}

	if err := mod.Load(); err != nil {
		fmt.Printf("ERROR cargando módulo: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Módulo cargado.")
	defer func() {
		_ = mod.Unload()
		fmt.Println("Módulo descargado.")
	}()


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
		_ = cron.Remove()
		fmt.Println("Cronjob eliminado.")
	}()


	if _, err := os.Stat(procSysinfo); err != nil {
		fmt.Printf("ERROR: no existe %s (¿PROC_NAME correcto y módulo cargado?)\n", procSysinfo)
		os.Exit(1)
	}

	
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	
	readTicker := time.NewTicker(readEvery)
	cleanTicker := time.NewTicker(cleanEvery)
	defer readTicker.Stop()
	defer cleanTicker.Stop()

	
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

	fmt.Printf("[sysinfo] Total=%dKB Free=%dKB Used=%dKB Procs=%d\n",
		si.Totalram, si.Freeram, used, si.Procs)
	return nil
}