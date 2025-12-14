package main

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	lowImage = "img_bajo"
	highCPU  = "img_cpu"
	highRAM  = "img_ram"
	keepLow  = 3
	keepHigh = 2
)

type Container struct {
	ID    string
	Image string
	Name  string

	CPUPerc  float64
	MemBytes uint64
}

func EnforceContainerPolicy() error {
	// ✅ 0) Limpia contenedores detenidos del proyecto (docker ps -a)
	_ = removeStoppedProjectContainers()

	// 1) contenedores corriendo
	containers, err := listRunningContainers()
	if err != nil {
		return err
	}
	if len(containers) == 0 {
		return nil
	}

	// 2) stats
	stats, err := fetchStats()
	if err != nil {
		return err
	}

	// 3) map ID -> stats
	idTo := map[string]Container{}
	for _, s := range stats {
		idTo[s.ID] = s
	}
	for i := range containers {
		if s, ok := idTo[containers[i].ID]; ok {
			containers[i].CPUPerc = s.CPUPerc
			containers[i].MemBytes = s.MemBytes
		}
	}

	// 4) decidir qué borrar (policy)
	toDelete := PickContainersToDelete(containers)

	// 5) borrar
	for _, c := range toDelete {
		_, _ = run("docker", "stop", c.ID)
		_, _ = run("docker", "rm", c.ID)
	}

	return nil
}

func removeStoppedProjectContainers() error {
	// Traemos TODOS los contenedores detenidos (exited) y filtramos por imagen del proyecto
	out, err := run("docker", "ps", "-a", "--filter", "status=exited", "--format", "{{.ID}}|{{.Image}}|{{.Names}}")
	if err != nil {
		return err
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")
	for _, ln := range lines {
		ln = strings.TrimSpace(ln)
		if ln == "" {
			continue
		}

		p := strings.Split(ln, "|")
		if len(p) != 3 {
			continue
		}

		id := p[0]
		img := p[1]
		name := p[2]

		c := Container{ID: id, Image: img, Name: name}

		// No tocar grafana
		if isGrafana(c) {
			continue
		}

		// Solo borrar imágenes del proyecto
		if strings.Contains(img, lowImage) || strings.Contains(img, highCPU) || strings.Contains(img, highRAM) {
			_, _ = run("docker", "rm", id)
		}
	}

	return nil
}

func usageScore(c Container) float64 {
	memMB := float64(c.MemBytes) / (1024.0 * 1024.0)
	return memMB + (c.CPUPerc * 10.0)
}

func isGrafana(c Container) bool {
	img := strings.ToLower(c.Image)
	name := strings.ToLower(c.Name)
	return strings.Contains(img, "grafana") || strings.Contains(name, "grafana")
}

func listRunningContainers() ([]Container, error) {
	out, err := run("docker", "ps", "--format", "{{.ID}}|{{.Image}}|{{.Names}}")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")
	var res []Container

	for _, ln := range lines {
		ln = strings.TrimSpace(ln)
		if ln == "" {
			continue
		}

		p := strings.Split(ln, "|")
		if len(p) != 3 {
			continue
		}
		res = append(res, Container{ID: p[0], Image: p[1], Name: p[2]})
	}

	return res, nil
}

func fetchStats() ([]Container, error) {
	out, err := run("docker", "stats", "--no-stream", "--format", "{{.Container}}|{{.CPUPerc}}|{{.MemUsage}}")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")
	var res []Container

	for _, ln := range lines {
		ln = strings.TrimSpace(ln)
		if ln == "" {
			continue
		}

		p := strings.Split(ln, "|")
		if len(p) != 3 {
			continue
		}

		id := p[0]
		cpu, _ := parsePercent(p[1])
		mem, _ := parseMemUsage(p[2])

		res = append(res, Container{ID: id, CPUPerc: cpu, MemBytes: mem})
	}

	return res, nil
}

func parsePercent(s string) (float64, error) {
	s = strings.TrimSpace(strings.TrimSuffix(s, "%"))
	if s == "" {
		return 0, errors.New("percent vacío")
	}
	return strconv.ParseFloat(s, 64)
}

func parseMemUsage(s string) (uint64, error) {
	s = strings.TrimSpace(s)
	parts := strings.Split(s, "/")
	if len(parts) < 1 {
		return 0, errors.New("mem inválido")
	}
	used := strings.TrimSpace(parts[0])
	return parseHumanBytes(used)
}

var reBytes = regexp.MustCompile(`^\s*([0-9]*\.?[0-9]+)\s*([KMGTP]?i?B|[KMGTP]?B)\s*$`)

func parseHumanBytes(s string) (uint64, error) {
	m := reBytes.FindStringSubmatch(strings.TrimSpace(s))
	if len(m) != 3 {
		return 0, fmt.Errorf("no pude parsear bytes: %q", s)
	}

	val, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return 0, err
	}

	unit := m[2]
	mult := float64(1)

	switch unit {
	case "B":
		mult = 1
	case "KiB", "KB":
		mult = 1024
	case "MiB", "MB":
		mult = 1024 * 1024
	case "GiB", "GB":
		mult = 1024 * 1024 * 1024
	case "TiB", "TB":
		mult = 1024 * 1024 * 1024 * 1024
	default:
		return 0, fmt.Errorf("unidad desconocida: %q", unit)
	}

	return uint64(val * mult), nil
}

func run(cmd string, args ...string) (string, error) {
	c := exec.Command(cmd, args...)
	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr

	if err := c.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("%s %v: %s", cmd, args, msg)
	}
	return stdout.String(), nil
}

func PickContainersToDelete(containers []Container) (toDelete []Container) {
	var low, high []Container

	for _, c := range containers {
		if isGrafana(c) {
			continue
		}
		if strings.Contains(c.Image, lowImage) {
			low = append(low, c)
			continue
		}
		if strings.Contains(c.Image, highCPU) || strings.Contains(c.Image, highRAM) {
			high = append(high, c)
			continue
		}
	}

	delLow, _ := trimByUsage(low, keepLow)
	delHigh, _ := trimHighPreferTypes(high, keepHigh)

	return append(delLow, delHigh...)
}

func trimByUsage(group []Container, keep int) (del []Container, kept []Container) {
	if len(group) <= keep {
		return nil, group
	}
	sort.Slice(group, func(i, j int) bool { return usageScore(group[i]) > usageScore(group[j]) })
	del = append(del, group[:len(group)-keep]...)
	kept = append(kept, group[len(group)-keep:]...)
	return
}

func trimHighPreferTypes(high []Container, keep int) (del []Container, kept []Container) {
	if len(high) <= keep {
		return nil, high
	}

	var cpuList, ramList []Container
	for _, c := range high {
		if strings.Contains(c.Image, highCPU) {
			cpuList = append(cpuList, c)
		} else if strings.Contains(c.Image, highRAM) {
			ramList = append(ramList, c)
		}
	}

	sort.Slice(cpuList, func(i, j int) bool { return usageScore(cpuList[i]) < usageScore(cpuList[j]) })
	sort.Slice(ramList, func(i, j int) bool { return usageScore(ramList[i]) < usageScore(ramList[j]) })

	candidates := []Container{}
	if len(cpuList) > 0 {
		candidates = append(candidates, cpuList[0])
	}
	if len(ramList) > 0 && len(candidates) < keep {
		candidates = append(candidates, ramList[0])
	}

	rest := []Container{}
	if len(cpuList) > 1 {
		rest = append(rest, cpuList[1:]...)
	}
	if len(ramList) > 1 {
		rest = append(rest, ramList[1:]...)
	}
	sort.Slice(rest, func(i, j int) bool { return usageScore(rest[i]) < usageScore(rest[j]) })

	for _, r := range rest {
		if len(candidates) >= keep {
			break
		}
		candidates = append(candidates, r)
	}

	keepID := map[string]bool{}
	for _, k := range candidates {
		keepID[k.ID] = true
	}

	for _, c := range high {
		if keepID[c.ID] {
			kept = append(kept, c)
		} else {
			del = append(del, c)
		}
	}
	return
}
