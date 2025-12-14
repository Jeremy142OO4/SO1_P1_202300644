package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func main() {
	// Ruta del script existente
	scriptPath := "/home/jeremy-kvm/Proyecto/bash/crear_contenedores.sh"

	// 1. Hacer el script ejecutable
	hacerEjecutable(scriptPath)

	// 2. Agregar cronjob que se ejecute cada minuto
	agregarCronJob(scriptPath)

	// 3. Verificar que se agregó correctamente
	verificarCronJobs()

	log.Println("Cronjob configurado exitosamente!")
}

// hacerEjecutable cambia los permisos del archivo para que sea ejecutable
func hacerEjecutable(ruta string) {
	err := os.Chmod(ruta, 0755)
	if err != nil {
		log.Fatalf("Error haciendo ejecutable el script: %v", err)
	}

	log.Printf("Script %s ahora es ejecutable", ruta)
}

// agregarCronJob agrega una nueva entrada a crontab
func agregarCronJob(rutaScript string) {
	expresionCron := "* * * * *"
	comandoCron := fmt.Sprintf("%s %s >> %s.log 2>&1", expresionCron, rutaScript, rutaScript)

	cmd := exec.Command("bash", "-c",
		fmt.Sprintf("(crontab -l 2>/dev/null; echo \"%s\") | crontab -", comandoCron))

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error agregando cronjob: %v\nOutput: %s", err, string(output))
	}

	log.Printf("Cronjob agregado: %s", comandoCron)
}

// verificarCronJobs lista todos los cronjobs configurados
func verificarCronJobs() {
	cmd := exec.Command("crontab", "-l")
	output, err := cmd.CombinedOutput()

	if err != nil {
		log.Printf("No se pudieron listar cronjobs (puede estar vacío): %v", err)
	} else {
		log.Printf("=== Cronjobs Actuales ===\n%s=== Fin de Cronjobs ===", string(output))
	}
}
