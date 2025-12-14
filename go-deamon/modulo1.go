package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type ModuleManager struct {
	KoPath     string 
	ModuleName string 
}

func (m ModuleManager) Load() error {
	loaded, err := isModuleLoaded(m.ModuleName)
	if err != nil {
		return err
	}
	if loaded {
		return nil
	}
	_, err = runCmd("sudo", "insmod", m.KoPath)
	if err != nil {
		return fmt.Errorf("insmod falló: %w", err)
	}
	return nil
}

func (m ModuleManager) Unload() error {
	loaded, err := isModuleLoaded(m.ModuleName)
	if err != nil {
		return err
	}
	if !loaded {
		return nil
	}
	_, err = runCmd("sudo", "rmmod", m.ModuleName)
	if err != nil {
		return fmt.Errorf("rmmod falló: %w", err)
	}
	return nil
}

func isModuleLoaded(moduleName string) (bool, error) {
	out, err := runCmd("bash", "-lc", "lsmod")
	if err != nil {
		return false, err
	}
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(line)
		if len(fields) > 0 && fields[0] == moduleName {
			return true, nil
		}
	}
	return false, nil
}

func runCmd(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("%s %v: %s", name, args, msg)
	}
	return stdout.String(), nil
}
