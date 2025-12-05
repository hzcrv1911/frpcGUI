package services

import (
	"fmt"
	"path/filepath"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"

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

// stopServiceSafe stops the service and waits for it to stop
func stopServiceSafe(serviceName string) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		// Service not found, assume stopped
		return nil
	}
	defer s.Close()

	status, err := s.Query()
	if err != nil {
		return err
	}

	if status.State == svc.Stopped {
		return nil
	}

	if _, err := s.Control(svc.Stop); err != nil {
		return fmt.Errorf("failed to send stop control: %v", err)
	}

	// Wait for service to stop
	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for service to stop")
		case <-ticker.C:
			status, err := s.Query()
			if err != nil {
				return err
			}
			if status.State == svc.Stopped {
				return nil
			}
		}
	}
}

// ReloadService restarts the WinSW-managed frp service
// which triggers hot-reloading of frp configuration.
func ReloadService(configPath string) error {
	// Check if WinSW is available
	if !IsWinSWAvailable() {
		return fmt.Errorf("WinSW executable not found")
	}

	var err error
	configPath, err = filepath.Abs(configPath)
	if err != nil {
		return err
	}

	serviceName := ServiceNameOfClient(configPath)

	// 1. Stop the service first to release file locks
	if err := stopServiceSafe(serviceName); err != nil {
		return fmt.Errorf("failed to stop service: %v", err)
	}

	// 2. Refresh profile directory (copies binaries and config)
	// This ensures frpc.exe and winsw.exe are present and correct
	profileDir, configFileName, err := prepareProfileDirectory(configPath)
	if err != nil {
		return fmt.Errorf("failed to prepare profile directory: %v", err)
	}

	// Convert to absolute path
	profileDir, err = filepath.Abs(profileDir)
	if err != nil {
		return err
	}

	winSWPath := filepath.Join(profileDir, "winsw.exe")
	logPath := filepath.Join(profileDir, "logs")
	profileConfigPath := filepath.Join(profileDir, configFileName)
	frpcPath := filepath.Join(profileDir, "frpc.exe")

	// 3. Create WinSW service configuration
	wsService := NewWinSWService(serviceName, profileConfigPath, winSWPath, frpcPath, logPath)

	// Regenerate config file to ensure it exists and matches this service
	if _, err := wsService.GenerateConfigFile(); err != nil {
		return fmt.Errorf("failed to regenerate winsw config: %v", err)
	}

	// 4. Start service
	return wsService.Start()
}
