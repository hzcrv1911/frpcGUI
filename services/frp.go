package services

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/hzcrv1911/frpcgui/pkg/config"
	"github.com/hzcrv1911/frpcgui/pkg/util"
)

// VerifyClientConfig validates the frp client config file
// For WinSW integration, we do basic validation without importing frp packages
func VerifyClientConfig(path string) error {
	// Basic file existence and format validation
	if _, err := os.Stat(path); err != nil {
		return err
	}

	// Try to parse the config file to validate its format
	_, err := config.UnmarshalClientConf(path)
	if err != nil {
		return err
	}

	// Additional validation for essential fields
	conf, err := config.UnmarshalClientConf(path)
	if err != nil {
		return err
	}

	// Check if server address is specified
	if conf.ServerAddress == "" {
		return util.NewError("server address is required")
	}

	// Check if server port is specified
	if conf.ServerPort <= 0 || conf.ServerPort > 65535 {
		return util.NewError("invalid server port")
	}

	// Check if at least one proxy is defined
	if len(conf.Proxies) == 0 {
		return util.NewError("at least one proxy must be defined")
	}

	// Validate each proxy
	for _, proxy := range conf.Proxies {
		if proxy.Name == "" {
			return util.NewError("proxy name is required")
		}
		if proxy.Type == "" {
			return util.NewError("proxy type is required")
		}
		// For non-visitor proxies, check local IP and port
		if !proxy.IsVisitor() {
			if proxy.LocalIP == "" {
				proxy.LocalIP = "127.0.0.1"
			}
			if proxy.LocalPort == "" {
				return util.NewError("local port is required for proxy: " + proxy.Name)
			}
		}
	}

	return nil
}

// GetFrpcVersion returns the version of the frpc.exe
func GetFrpcVersion() (string, error) {
	frpcPath, err := GetFrpcPath()
	if err != nil {
		return "", err
	}

	// Execute frpc.exe with version flag
	output, err := util.ExecuteCommandWithOutput(frpcPath + " -v")
	if err != nil {
		return "", err
	}

	// Extract version from output
	lines := strings.Split(output, "\n")
	if len(lines) > 0 {
		version := strings.TrimSpace(lines[0])
		// Remove "frpc " prefix if present
		version = strings.TrimPrefix(version, "frpc ")
		return version, nil
	}

	return "", util.NewError("unable to determine frpc version")
}

// IsFrpcRunning checks if frpc.exe is running for the given config
func IsFrpcRunning(configPath string) (bool, error) {
	serviceName := ServiceNameOfClient(configPath)

	// Check if WinSW is available
	if !IsWinSWAvailable() {
		return false, util.NewError("WinSW not available")
	}

	winSWPath, err := GetWinSWPath()
	if err != nil {
		return false, err
	}

	// Create log directory
	logPath := filepath.Join(filepath.Dir(configPath), "logs")

	// Create WinSW service
	wsService := NewWinSWService(serviceName, configPath, winSWPath, "", logPath)

	// Get service status
	status, err := wsService.Status()
	if err != nil {
		return false, err
	}

	return status == "Running", nil
}
