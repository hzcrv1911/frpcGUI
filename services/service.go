package services

import (
	"fmt"
	"path/filepath"

	"github.com/hzcrv1911/frpcgui/pkg/util"
)

func ServiceNameOfClient(configPath string) string {
	// Use the config filename without extension as service name
	// This makes the service name readable and consistent with the config file
	filename := util.FileNameWithoutExt(configPath)
	return fmt.Sprintf("frpc_%s", filename)
}

func DisplayNameOfClient(name string) string {
	return "FRPCGUI: " + name
}

// ReloadService restarts the WinSW-managed frp service
// which triggers hot-reloading of frp configuration.
func ReloadService(configPath string) error {
	// Check if WinSW is available
	if !IsWinSWAvailable() {
		return fmt.Errorf("WinSW executable not found")
	}

	// Get paths
	winSWPath, err := GetWinSWPath()
	if err != nil {
		return err
	}

	if configPath, err = filepath.Abs(configPath); err != nil {
		return err
	}

	// Create log directory
	logPath := filepath.Join(filepath.Dir(configPath), "logs")

	// Create WinSW service
	serviceName := ServiceNameOfClient(configPath)
	wsService := NewWinSWService(serviceName, configPath, winSWPath, "", logPath)

	// Restart service to reload configuration
	return wsService.Restart()
}
