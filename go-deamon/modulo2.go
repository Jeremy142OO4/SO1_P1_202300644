package main

import (
	"encoding/json"
	"fmt"
	"os"
)


type ContInfo struct {
	Count      int              `json:"Count"`
	Containers []ContainerInfo  `json:"Containers"`
}

type ContainerInfo struct {
	ContainerID string `json:"ContainerID"`
	CgroupPath  string `json:"CgroupPath"`
	RSSKB       uint64 `json:"RSS_KB"`
	CPUJiffies  uint64 `json:"CPU_Jiffies"`
	Procs       uint32 `json:"Procs"`
}


func ReadContainerInfo() (*ContInfo, error) {
	raw, err := os.ReadFile(procContinfo)
	if err != nil {
		return nil, fmt.Errorf("no se pudo leer %s: %w", procContinfo, err)
	}

	var ci ContInfo
	if err := json.Unmarshal(raw, &ci); err != nil {
		return nil, fmt.Errorf("JSON inválido en continfo: %w", err)
	}

	return &ci, nil
}


func PrintContainerInfo(ci *ContInfo) {
	fmt.Println("=== CONTINFO ===")
	fmt.Printf("Contenedores detectados: %d\n", ci.Count)

	for i, c := range ci.Containers {
		fmt.Printf(
			"#%d ID=%s RSS=%dKB CPU_Jiffies=%d Procs=%d\n",
			i+1,
			shortID(c.ContainerID),
			c.RSSKB,
			c.CPUJiffies,
			c.Procs,
		)
	}
	fmt.Println()
}

func shortID(id string) string {
	if len(id) > 12 {
		return id[:12]
	}
	return id
}
