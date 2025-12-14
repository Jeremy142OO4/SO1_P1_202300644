package main

import (
	"fmt"
	"strings"
)

type CronManager struct {
	ScriptPath string 
	LogPath    string 
}

func (c CronManager) InstallEveryMinute() error {
	line := fmt.Sprintf("* * * * * /bin/bash %s >> %s 2>&1", c.ScriptPath, c.LogPath)

	
	current, _ := runCmd("bash", "-lc", "crontab -l 2>/dev/null || true")
	if strings.Contains(current, c.ScriptPath) {
		return nil
	}

	
	cmd := fmt.Sprintf(`(crontab -l 2>/dev/null; echo "%s") | crontab -`, line)
	_, err := runCmd("bash", "-lc", cmd)
	if err != nil {
		return fmt.Errorf("agregar cronjob falló: %w", err)
	}
	return nil
}

func (c CronManager) Remove() error {
	
	cmd := fmt.Sprintf(`crontab -l 2>/dev/null | grep -vF "%s" | crontab -`, c.ScriptPath)
	_, err := runCmd("bash", "-lc", cmd)
	return err
}
