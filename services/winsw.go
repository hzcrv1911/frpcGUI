package services

import (
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/hzcrv1911/frpcgui/pkg/util"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

// WinSWConfig represents the configuration for WinSW service
type WinSWConfig struct {
	XMLName     xml.Name `xml:"service"`
	ID          string   `xml:"id"`
	Name        string   `xml:"name"`
	Desc        string   `xml:"description"`
	Executable  string   `xml:"executable"`
	Arguments   string   `xml:"arguments"`
	LogPath     string   `xml:"log>directory"`
	LogMode     string   `xml:"log>mode"`
	StopTimeout string   `xml:"stoptimeout,omitempty"`
	StartMode   string   `xml:"startmode"`
}

// WinSWService represents a service managed by WinSW
type WinSWService struct {
	ServiceName string
	ConfigPath  string
	WinSWPath   string
	FrpcPath    string
	LogPath     string
}

// NewWinSWService creates a new WinSW service instance
func NewWinSWService(serviceName, configPath, winSWPath, frpcPath, logPath string) *WinSWService {
	return &WinSWService{
		ServiceName: serviceName,
		ConfigPath:  configPath,
		WinSWPath:   winSWPath,
		FrpcPath:    frpcPath,
		LogPath:     logPath,
	}
}

// GenerateConfigFile generates the WinSW configuration file
func (ws *WinSWService) GenerateConfigFile() (string, error) {
	// Ensure log directory exists
	if err := os.MkdirAll(ws.LogPath, os.ModePerm); err != nil {
		return "", err
	}

	// Convert all paths to absolute paths
	frpcPath, err := filepath.Abs(ws.FrpcPath)
	if err != nil {
		frpcPath = ws.FrpcPath
	}
	configPath, err := filepath.Abs(ws.ConfigPath)
	if err != nil {
		configPath = ws.ConfigPath
	}
	logPath, err := filepath.Abs(ws.LogPath)
	if err != nil {
		logPath = ws.LogPath
	}

	// Create WinSW configuration
	config := WinSWConfig{
		ID:          ws.ServiceName,
		Name:        ws.ServiceName,
		Desc:        "FRPC Runtime Service(" + ws.ServiceName + ")",
		Executable:  frpcPath,
		Arguments:   fmt.Sprintf("-c %s", configPath),
		LogPath:     logPath,
		LogMode:     "roll",
		StopTimeout: "15s",
		StartMode:   "Automatic",
	}

	// Marshal to XML
	data, err := xml.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", err
	}

	// Add XML header
	xmlContent := xml.Header + string(data)

	// Write to config file: winsw.xml in the same directory as WinSW exe
	winswDir := filepath.Dir(ws.WinSWPath)
	configFile := filepath.Join(winswDir, "winsw.xml")
	if err := os.WriteFile(configFile, []byte(xmlContent), 0644); err != nil {
		return "", err
	}

	return configFile, nil
}

// Install installs the service using WinSW
func (ws *WinSWService) Install() error {
	// Generate WinSW config file
	configFile, err := ws.GenerateConfigFile()
	if err != nil {
		return err
	}

	// Check if service already exists
	m, err := serviceManager()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	service, err := m.OpenService(ws.ServiceName)
	if err == nil {
		// Service exists, stop and delete it first
		service.Control(svc.Stop)
		service.Delete()
		service.Close()
	}

	// Install service using WinSW
	cmd := fmt.Sprintf("%s install %s", ws.WinSWPath, configFile)
	if err := util.ExecuteCommand(cmd); err != nil {
		return fmt.Errorf("failed to install service: %v", err)
	}

	// Start the service
	cmd = fmt.Sprintf("%s start %s", ws.WinSWPath, ws.ServiceName)
	if err := util.ExecuteCommand(cmd); err != nil {
		return fmt.Errorf("failed to start service: %v", err)
	}

	return nil
}

// Uninstall uninstalls the service using WinSW
func (ws *WinSWService) Uninstall() error {
	// Stop the service
	stopCmd := exec.Command(ws.WinSWPath, "stop", ws.ServiceName)
	stopCmd.Run() // Ignore errors when stopping, service might already be stopped

	// Uninstall the service
	uninstallCmd := exec.Command(ws.WinSWPath, "uninstall", ws.ServiceName)
	if output, err := uninstallCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to uninstall service: %v, output: %s", err, string(output))
	}

	// Remove config file
	configFile := ws.ConfigPath + ".winsw.xml"
	os.Remove(configFile)

	return nil
}

// Start starts the service using WinSW
func (ws *WinSWService) Start() error {
	// Start the service
	startCmd := exec.Command(ws.WinSWPath, "start", ws.ServiceName)
	if output, err := startCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to start service: %v, output: %s", err, string(output))
	}
	return nil
}

// Stop stops the service using WinSW
func (ws *WinSWService) Stop() error {
	// Stop the service
	stopCmd := exec.Command(ws.WinSWPath, "stop", ws.ServiceName)
	if output, err := stopCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to stop service: %v, output: %s", err, string(output))
	}
	return nil
}

// Restart restarts the service using WinSW
func (ws *WinSWService) Restart() error {
	// Restart the service
	restartCmd := exec.Command(ws.WinSWPath, "restart", ws.ServiceName)
	if output, err := restartCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to restart service: %v, output: %s", err, string(output))
	}
	return nil
}

// Status returns the status of the service
func (ws *WinSWService) Status() (string, error) {
	// For WinSW status checks, always use a fresh connection to avoid stale data
	// when services are installed/uninstalled frequently
	m, err := mgr.Connect()
	if err != nil {
		return "", err
	}
	defer m.Disconnect()

	// Always try to open with read-only access first for reliability
	// This approach is more robust than relying on full access
	h, err := windows.OpenService(m.Handle, windows.StringToUTF16Ptr(ws.ServiceName), windows.SERVICE_QUERY_STATUS)
	if err != nil {
		// Service not found - genuinely not installed
		return "Not installed", nil
	}
	defer windows.CloseServiceHandle(h)

	var status windows.SERVICE_STATUS
	if err := windows.QueryServiceStatus(h, &status); err != nil {
		// Can't query status - treat as unknown
		return "Unknown", nil
	}

	switch status.CurrentState {
	case windows.SERVICE_STOPPED:
		return "Stopped", nil
	case windows.SERVICE_START_PENDING:
		return "Starting", nil
	case windows.SERVICE_STOP_PENDING:
		return "Stopping", nil
	case windows.SERVICE_RUNNING:
		return "Running", nil
	case windows.SERVICE_CONTINUE_PENDING:
		return "Continuing", nil
	case windows.SERVICE_PAUSE_PENDING:
		return "Pausing", nil
	case windows.SERVICE_PAUSED:
		return "Paused", nil
	default:
		return "Unknown", nil
	}
}

// IsWinSWAvailable checks if WinSW is available in the system
func IsWinSWAvailable() bool {
	_, err := GetWinSWPath()
	return err == nil
}

// GetWinSWPath returns the path to WinSW executable
func GetWinSWPath() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", err
	}
	dir := filepath.Dir(path)

	// Get working directory for fallback
	wd, _ := os.Getwd()

	// Try multiple possible locations for winsw.exe and winsw-x64.exe
	possiblePaths := []string{
		filepath.Join(dir, "assets", "winsw.exe"),     // Release: assets subdirectory relative to exe
		filepath.Join(dir, "assets", "winsw-x64.exe"), // Release: assets subdirectory (alternative name)
		filepath.Join(wd, "assets", "winsw.exe"),      // Debug: assets relative to working directory
		filepath.Join(wd, "assets", "winsw-x64.exe"),  // Debug: assets (alternative name)
		filepath.Join(dir, "winsw.exe"),               // Legacy: same directory as exe
		filepath.Join(dir, "winsw-x64.exe"),           // Legacy: same directory (alternative name)
	}

	var triedPaths []string
	for _, winswPath := range possiblePaths {
		if absPath, err := filepath.Abs(winswPath); err == nil {
			triedPaths = append(triedPaths, absPath)
			if _, err := os.Stat(absPath); err == nil {
				return absPath, nil
			}
		}
	}

	return "", fmt.Errorf("WinSW executable not found. Tried paths: %v. Exe dir: %s, Working dir: %s", triedPaths, dir, wd)
}

// IsFrpcAvailable checks if frpc.exe is available in the system
func IsFrpcAvailable() bool {
	_, err := GetFrpcPath()
	return err == nil
}

// GetFrpcPath returns the path to frpc.exe
func GetFrpcPath() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", err
	}
	dir := filepath.Dir(path)

	// Get working directory for fallback
	wd, _ := os.Getwd()

	// Try multiple possible locations
	possiblePaths := []string{
		filepath.Join(dir, "assets", "frpc.exe"), // Release: assets subdirectory relative to exe
		filepath.Join(wd, "assets", "frpc.exe"),  // Debug: assets relative to working directory
		filepath.Join(dir, "frpc.exe"),           // Legacy: same directory as exe
	}

	var triedPaths []string
	for _, frpcPath := range possiblePaths {
		if absPath, err := filepath.Abs(frpcPath); err == nil {
			triedPaths = append(triedPaths, absPath)
			if _, err := os.Stat(absPath); err == nil {
				return absPath, nil
			}
		}
	}

	return "", fmt.Errorf("frpc.exe not found. Tried paths: %v. Exe dir: %s, Working dir: %s", triedPaths, dir, wd)
}
