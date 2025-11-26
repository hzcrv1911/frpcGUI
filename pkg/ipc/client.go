package ipc

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hzcrv1911/frpcgui/services"
)

// ProxyMessage is status information of a proxy.
type ProxyMessage struct {
	Name       string
	Type       string
	Status     string
	Err        string
	RemoteAddr string
}

// Client is used to query proxy state from frp client.
// In WinSW mode, it's a stub implementation that doesn't actually communicate with frpc.
type Client interface {
	// SetCallback changes the callback function for response message.
	SetCallback(cb func([]ProxyMessage))
	// Run client in blocking mode.
	Run(ctx context.Context)
	// Probe triggers a query request immediately.
	Probe(ctx context.Context)
}

// WinSWClient is a stub implementation for WinSW integration
// It doesn't actually communicate with frpc, but provides the same interface
type WinSWClient struct {
	callback   func([]ProxyMessage)
	configPath string
}

// NewClient creates a new client instance
// In WinSW mode, it returns a stub client
func NewClient(config interface{}, callback func([]ProxyMessage)) Client {
	// This is a placeholder for compatibility
	// In practice, we'll use NewWinSWClient instead
	return &WinSWClient{
		callback:   callback,
		configPath: "",
	}
}

// NewWinSWClient creates a new WinSW client with config path
func NewWinSWClient(configPath string, callback func([]ProxyMessage)) Client {
	return &WinSWClient{
		callback:   callback,
		configPath: configPath,
	}
}

// SetCallback sets the callback function for response messages
func (c *WinSWClient) SetCallback(cb func([]ProxyMessage)) {
	c.callback = cb
}

// Run runs the client in blocking mode
// In WinSW mode, this doesn't do anything since we don't have direct communication
func (c *WinSWClient) Run(ctx context.Context) {
	// In WinSW mode, we don't have direct IPC communication with frpc
	// The actual status updates will be done by parsing log files
	// This is just a placeholder to satisfy the interface
	<-ctx.Done()
}

// Probe triggers a query request immediately
// In WinSW mode, this doesn't do anything since we don't have direct communication
func (c *WinSWClient) Probe(ctx context.Context) {
	// In WinSW mode, we don't have direct IPC communication with frpc
	// The actual status updates will be done by parsing log files
	// This is just a placeholder to satisfy the interface
}

// GetProxyStatusFromLogs retrieves proxy status by parsing log files
// This is a helper function for WinSW integration
func GetProxyStatusFromLogs(configPath string) ([]ProxyMessage, error) {
	var messages []ProxyMessage

	// Determine log file path
	logFile := filepath.Join(filepath.Dir(configPath), "logs", "frpc.log")

	// Check if file exists
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		return messages, fmt.Errorf("log file not found: %s", logFile)
	}

	// Read log file
	file, err := os.Open(logFile)
	if err != nil {
		return messages, err
	}
	defer file.Close()

	// Scan file for proxy status
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Extract proxy status from log line
		msg, err := extractProxyStatusFromLog(line)
		if err != nil {
			continue
		}

		messages = append(messages, msg)
	}

	return messages, nil
}

// GetProxyStatusFromService retrieves proxy status from WinSW service
func GetProxyStatusFromService(configPath string) ([]ProxyMessage, error) {
	var messages []ProxyMessage

	// Check if service is running
	running, err := services.IsFrpcRunning(configPath)
	if err != nil {
		return messages, err
	}

	if !running {
		// Service is not running, all proxies are stopped
		// We can't determine the exact proxy list without parsing the config
		// So we return an empty list and let the UI handle it
		return messages, nil
	}

	// Service is running, try to get detailed status from logs
	messages, err = GetProxyStatusFromLogs(configPath)
	if err != nil {
		// If we can't get status from logs, return an empty list
		// and let the UI handle it
		return nil, nil
	}

	return messages, nil
}

// extractProxyStatusFromLog extracts proxy status from a log line
func extractProxyStatusFromLog(line string) (ProxyMessage, error) {
	// Look for proxy status messages in log
	// Examples:
	// "proxy [tcp:ssh] starts successfully"
	// "proxy [http:web] starts successfully"
	// "proxy [tcp:ssh] error: connection refused"

	// Check for proxy start message
	if strings.Contains(line, "starts successfully") {
		return extractProxyInfo(line, "running", "")
	}

	// Check for proxy error messages
	if strings.Contains(line, "error:") || strings.Contains(line, "failed") {
		// Extract error message
		var errorMsg string
		parts := strings.SplitN(line, "error:", 2)
		if len(parts) > 1 {
			errorMsg = strings.TrimSpace(parts[1])
		} else {
			parts = strings.SplitN(line, "failed", 2)
			if len(parts) > 1 {
				errorMsg = strings.TrimSpace(parts[1])
			}
		}
		return extractProxyInfo(line, "error", errorMsg)
	}

	// Check for proxy stop messages
	if strings.Contains(line, "proxy stopped") {
		return extractProxyInfo(line, "stopped", "")
	}

	return ProxyMessage{}, fmt.Errorf("no proxy status found in line")
}

// extractProxyInfo extracts proxy information from a log line
func extractProxyInfo(line, status, errorMsg string) (ProxyMessage, error) {
	// Extract proxy name and type from log line
	// Format: "proxy [type:name] message"
	start := strings.Index(line, "[")
	end := strings.Index(line, "]")
	if start == -1 || end == -1 {
		return ProxyMessage{}, fmt.Errorf("invalid proxy info format")
	}

	proxyInfo := line[start+1 : end]
	parts := strings.SplitN(proxyInfo, ":", 2)
	if len(parts) != 2 {
		return ProxyMessage{}, fmt.Errorf("invalid proxy info format")
	}

	proxyType := parts[0]
	proxyName := parts[1]

	// Create proxy message
	msg := ProxyMessage{
		Name:   proxyName,
		Type:   proxyType,
		Status: status,
		Err:    errorMsg,
	}

	// Try to extract remote address if available
	if strings.Contains(line, "remote address:") {
		addrStart := strings.Index(line, "remote address:") + len("remote address:")
		addrEnd := strings.Index(line[addrStart:], ",")
		if addrEnd == -1 {
			addrEnd = len(line[addrStart:])
		}
		remoteAddr := strings.TrimSpace(line[addrStart : addrStart+addrEnd])
		msg.RemoteAddr = remoteAddr
	}

	return msg, nil
}
