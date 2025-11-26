package services

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hzcrv1911/frpcgui/pkg/config"
)

// GetProfileDirectory returns the profile directory path for a config
// Format: profiles/R_<server_ip>_<port> (dots and colons replaced with underscores)
func GetProfileDirectory(configPath string) (string, error) {
	// Load config to get server address and port
	cfg, err := config.UnmarshalClientConf(configPath)
	if err != nil {
		return "", fmt.Errorf("failed to load config: %v", err)
	}

	// Generate directory name: R_<ip>_<port>
	// Replace dots and colons with underscores
	serverAddr := strings.ReplaceAll(cfg.ServerAddress, ".", "_")
	serverAddr = strings.ReplaceAll(serverAddr, ":", "_")
	dirName := fmt.Sprintf("R_%s_%d", serverAddr, cfg.ServerPort)

	return filepath.Join("profiles", dirName), nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Create destination directory if not exists
	if err := os.MkdirAll(filepath.Dir(dst), os.ModePerm); err != nil {
		return err
	}

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// prepareProfileDirectory creates profile directory and copies assets
func prepareProfileDirectory(configPath string) (string, error) {
	// Get profile directory path
	profileDir, err := GetProfileDirectory(configPath)
	if err != nil {
		return "", err
	}

	// Create profile directory
	if err := os.MkdirAll(profileDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create profile directory: %v", err)
	}

	// Get assets directory
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	exeDir := filepath.Dir(exePath)

	// Try multiple possible asset locations
	possibleAssetDirs := []string{
		filepath.Join(exeDir, "assets"),
		"assets",
	}

	var assetsDir string
	for _, dir := range possibleAssetDirs {
		if _, err := os.Stat(dir); err == nil {
			assetsDir = dir
			break
		}
	}

	if assetsDir == "" {
		return "", fmt.Errorf("assets directory not found")
	}

	// Copy winsw.exe with fixed name
	winswSrc := filepath.Join(assetsDir, "winsw.exe")
	winswDst := filepath.Join(profileDir, "winsw.exe")
	if _, err := os.Stat(winswSrc); err == nil {
		if err := copyFile(winswSrc, winswDst); err != nil {
			return "", fmt.Errorf("failed to copy winsw.exe: %v", err)
		}
	}

	// Copy frpc.exe
	frpcSrc := filepath.Join(assetsDir, "frpc.exe")
	frpcDst := filepath.Join(profileDir, "frpc.exe")
	if _, err := os.Stat(frpcSrc); err == nil {
		if err := copyFile(frpcSrc, frpcDst); err != nil {
			return "", fmt.Errorf("failed to copy frpc.exe: %v", err)
		}
	}

	// Copy config file to profile directory
	// Determine config file name based on extension
	configExt := filepath.Ext(configPath)
	var configFileName string
	if configExt == ".toml" {
		configFileName = "frpc.toml"
	} else {
		configFileName = "frpc.ini"
	}

	configDst := filepath.Join(profileDir, configFileName)
	if err := copyFile(configPath, configDst); err != nil {
		return "", fmt.Errorf("failed to copy config file: %v", err)
	}

	return profileDir, nil
}

// InstallWinSWService installs the WinSW service without starting it
func InstallWinSWService(name string, configPath string, manual bool) error {
	// Check if WinSW is available and get detailed error
	if _, err := GetWinSWPath(); err != nil {
		return fmt.Errorf("WinSW executable not found: %v", err)
	}

	// Check if frpc.exe is available and get detailed error
	if _, err := GetFrpcPath(); err != nil {
		return fmt.Errorf("frpc.exe not found: %v", err)
	}

	// Get original config path for service name
	originalConfigPath, err := filepath.Abs(configPath)
	if err != nil {
		return err
	}

	// Generate service name first
	serviceName := ServiceNameOfClient(originalConfigPath)

	// Prepare profile directory and copy assets
	profileDir, err := prepareProfileDirectory(configPath)
	if err != nil {
		return fmt.Errorf("failed to prepare profile directory: %v", err)
	}

	// Convert to absolute path
	profileDir, err = filepath.Abs(profileDir)
	if err != nil {
		return err
	}

	// Use executables from profile directory
	winSWPath := filepath.Join(profileDir, "winsw.exe")
	frpcPath := filepath.Join(profileDir, "frpc.exe")

	// Use config file in profile directory
	configExt := filepath.Ext(configPath)
	var configFileName string
	if configExt == ".toml" {
		configFileName = "frpc.toml"
	} else {
		configFileName = "frpc.ini"
	}
	profileConfigPath := filepath.Join(profileDir, configFileName)

	// Create log directory in profile directory
	logPath := filepath.Join(profileDir, "logs")
	if err := os.MkdirAll(logPath, os.ModePerm); err != nil {
		return err
	}

	// Create WinSW service
	wsService := NewWinSWService(serviceName, profileConfigPath, winSWPath, frpcPath, logPath)

	// Generate config file (will be <serviceName>.xml)
	_, err = wsService.GenerateConfigFile()
	if err != nil {
		return err
	}

	// Check if service already exists
	m, err := serviceManager()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	service, err := m.OpenService(serviceName)
	if err == nil {
		// Service exists, delete it first
		service.Close()
		return fmt.Errorf("service already installed")
	}

	// Install service using WinSW (no need to specify config file, it will find <serviceName>.xml)
	// Change working directory to profile directory so winsw can find its config
	cmd := exec.Command(winSWPath, "install")
	cmd.Dir = profileDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to install service: %v, output: %s", err, string(output))
	}

	return nil
}

// StartWinSWService starts an already installed WinSW service
func StartWinSWService(configPath string) error {
	if !IsWinSWAvailable() {
		return fmt.Errorf("WinSW not available")
	}

	var err error
	configPath, err = filepath.Abs(configPath)
	if err != nil {
		return err
	}

	// Get profile directory
	profileDir, err := GetProfileDirectory(configPath)
	if err != nil {
		return err
	}

	// Convert to absolute path
	profileDir, err = filepath.Abs(profileDir)
	if err != nil {
		return err
	}

	serviceName := ServiceNameOfClient(configPath)
	winSWPath := filepath.Join(profileDir, "winsw.exe")
	logPath := filepath.Join(profileDir, "logs")

	// Get config file in profile directory
	configExt := filepath.Ext(configPath)
	var configFileName string
	if configExt == ".toml" {
		configFileName = "frpc.toml"
	} else {
		configFileName = "frpc.ini"
	}
	profileConfigPath := filepath.Join(profileDir, configFileName)

	wsService := NewWinSWService(serviceName, profileConfigPath, winSWPath, "", logPath)

	return wsService.Start()
}

// StopWinSWService stops a running WinSW service
func StopWinSWService(configPath string) error {
	if !IsWinSWAvailable() {
		return fmt.Errorf("WinSW not available")
	}

	var err error
	configPath, err = filepath.Abs(configPath)
	if err != nil {
		return err
	}

	// Get profile directory
	profileDir, err := GetProfileDirectory(configPath)
	if err != nil {
		return err
	}

	// Convert to absolute path
	profileDir, err = filepath.Abs(profileDir)
	if err != nil {
		return err
	}

	serviceName := ServiceNameOfClient(configPath)
	winSWPath := filepath.Join(profileDir, "winsw.exe")
	logPath := filepath.Join(profileDir, "logs")

	// Get config file in profile directory
	configExt := filepath.Ext(configPath)
	var configFileName string
	if configExt == ".toml" {
		configFileName = "frpc.toml"
	} else {
		configFileName = "frpc.ini"
	}
	profileConfigPath := filepath.Join(profileDir, configFileName)

	wsService := NewWinSWService(serviceName, profileConfigPath, winSWPath, "", logPath)

	return wsService.Stop()
}

// InstallService runs the program as Windows service using WinSW (installs and starts)
func InstallService(name string, configPath string, manual bool) error {
	// Install the service
	if err := InstallWinSWService(name, configPath, manual); err != nil {
		return err
	}

	// Start the service
	return StartWinSWService(configPath)
}

// UninstallService stops and removes the given service using WinSW
func UninstallService(configPath string, wait bool) error {
	// Check if WinSW is available
	if !IsWinSWAvailable() {
		return fmt.Errorf("WinSW executable not found. Please place winsw.exe in the same directory as the application.")
	}

	var err error
	configPath, err = filepath.Abs(configPath)
	if err != nil {
		return err
	}

	// Get profile directory
	profileDir, err := GetProfileDirectory(configPath)
	if err != nil {
		return err
	}

	// Convert to absolute path
	profileDir, err = filepath.Abs(profileDir)
	if err != nil {
		return err
	}

	serviceName := ServiceNameOfClient(configPath)
	winSWPath := filepath.Join(profileDir, "winsw.exe")
	logPath := filepath.Join(profileDir, "logs")

	// Get config file in profile directory
	configExt := filepath.Ext(configPath)
	var configFileName string
	if configExt == ".toml" {
		configFileName = "frpc.toml"
	} else {
		configFileName = "frpc.ini"
	}
	profileConfigPath := filepath.Join(profileDir, configFileName)

	// Create WinSW service
	wsService := NewWinSWService(serviceName, profileConfigPath, winSWPath, "", logPath)

	// Uninstall service
	if err := wsService.Uninstall(); err != nil {
		return err
	}

	// Clean up profile directory without touching config files or R* folders
	cleanupProfileArtifacts(profileDir)

	return nil
}

// cleanupProfileArtifacts removes WinSW artifacts but keeps user configs and R* directories.
func cleanupProfileArtifacts(profileDir string) {
	entries, err := os.ReadDir(profileDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return
		}
		return
	}

	for _, entry := range entries {
		name := entry.Name()
		fullPath := filepath.Join(profileDir, name)

		if entry.IsDir() {
			if hasRPrefix(name) {
				continue
			}
			os.RemoveAll(fullPath)
			continue
		}

		if strings.EqualFold(filepath.Ext(name), ".conf") {
			continue
		}

		os.Remove(fullPath)
	}
}

func hasRPrefix(name string) bool {
	if name == "" {
		return false
	}
	first := name[0]
	return first == 'R' || first == 'r'
}

// QueryStartInfo returns the start type and process id of the given service.
// For WinSW-managed services, we return default values since WinSW handles the service lifecycle.
func QueryStartInfo(configPath string) (uint32, uint32, error) {
	// Check if WinSW is available
	if !IsWinSWAvailable() {
		return 0, 0, fmt.Errorf("WinSW executable not found")
	}

	// Get paths
	winSWPath, err := GetWinSWPath()
	if err != nil {
		return 0, 0, err
	}

	if configPath, err = filepath.Abs(configPath); err != nil {
		return 0, 0, err
	}

	// Create log directory
	logPath := filepath.Join(filepath.Dir(configPath), "logs")

	// Create WinSW service
	serviceName := ServiceNameOfClient(configPath)
	wsService := NewWinSWService(serviceName, configPath, winSWPath, "", logPath)

	// Get service status
	status, err := wsService.Status()
	if err != nil {
		return 0, 0, err
	}

	// For WinSW services, we return default start type and 0 for PID
	// since WinSW manages the service lifecycle
	var startType uint32 = 2 // SERVICE_AUTO_START
	var pid uint32 = 0

	if status == "Stopped" {
		return startType, pid, nil
	}

	return startType, pid, nil
}
