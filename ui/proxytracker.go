package ui

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/lxn/walk"

	"github.com/hzcrv1911/frpcgui/pkg/config"
	"github.com/hzcrv1911/frpcgui/pkg/consts"
	"github.com/hzcrv1911/frpcgui/pkg/ipc"
	"github.com/hzcrv1911/frpcgui/services"
)

type ProxyTracker struct {
	sync.RWMutex
	owner              walk.Form
	model              *ProxyModel
	cache              map[string]*config.Proxy
	ctx                context.Context
	cancel             context.CancelFunc
	refreshTimer       *time.Timer
	ticker             *time.Ticker
	lastLogPos         int64
	logFilePath        string
	rowsInsertedHandle int
	beforeRemoveHandle int
	rowEditedHandle    int
	rowRenamedHandle   int
}

func NewProxyTracker(owner walk.Form, model *ProxyModel, refresh bool) (tracker *ProxyTracker) {
	// For backward compatibility, use the config from the model
	configPath := ""
	if model != nil && model.conf != nil {
		configPath = getCurrentConf().Path
	}
	return NewProxyTrackerWithConfig(owner, model, configPath, refresh)
}

func NewProxyTrackerWithConfig(owner walk.Form, model *ProxyModel, configPath string, refresh bool) (tracker *ProxyTracker) {
	cache := make(map[string]*config.Proxy)
	ctx, cancel := context.WithCancel(context.Background())

	// Determine log file path
	logFilePath := model.conf.Data.LogFile
	if logFilePath == "" || logFilePath == "console" {
		// Default log file location
		if configPath != "" {
			configDir := filepath.Dir(configPath)
			logFilePath = filepath.Join(configDir, "logs", "frpc.log")
		} else {
			configDir := filepath.Dir(getCurrentConf().Path)
			logFilePath = filepath.Join(configDir, "logs", "frpc.log")
		}
	}

	tracker = &ProxyTracker{
		owner:       owner,
		model:       model,
		cache:       cache,
		ctx:         ctx,
		cancel:      cancel,
		logFilePath: logFilePath,
		rowsInsertedHandle: model.RowsInserted().Attach(func(from, to int) {
			tracker.Lock()
			defer tracker.Unlock()
			for i := from; i <= to; i++ {
				for _, key := range model.items[i].GetAlias() {
					cache[key] = model.items[i].Proxy
				}
			}
			// In WinSW mode, we trigger a status check
			tracker.checkProxyStatus()
		}),
		beforeRemoveHandle: model.BeforeRemove().Attach(func(i int) {
			tracker.Lock()
			defer tracker.Unlock()
			for _, key := range model.items[i].GetAlias() {
				delete(cache, key)
			}
		}),
		rowEditedHandle: model.RowEdited().Attach(func(i int) {
			// In WinSW mode, we trigger a status check
			tracker.checkProxyStatus()
		}),
		rowRenamedHandle: model.RowRenamed().Attach(func(i int) {
			tracker.buildCache()
		}),
	}
	tracker.buildCache()

	// Start log file monitoring
	go tracker.monitorLogFile()

	// If no status information is received within a certain period of time,
	// we need to refresh the view to make the icon visible.
	if refresh {
		tracker.refreshTimer = time.AfterFunc(300*time.Millisecond, func() {
			owner.Synchronize(func() {
				if ctx.Err() != nil {
					return
				}
				model.PublishRowsChanged(0, len(model.items)-1)
			})
		})
	}
	return
}

func (pt *ProxyTracker) onMessage(msg []ipc.ProxyMessage) {
	pt.RLock()
	defer pt.RUnlock()

	// Update proxy status in the model
	for _, m := range msg {
		// Convert string status to ProxyState
		proxyState := stringToProxyState(m.Status)
		for i, item := range pt.model.items {
			if item.Name == m.Name {
				// Update proxy status
				item.State = proxyState
				item.Error = m.Err
				item.RemoteAddr = m.RemoteAddr
				item.StateSource = m.Name

				// Update remote port display
				item.UpdateRemotePort()

				// Notify UI of the change
				pt.model.PublishRowChanged(i)
				break
			}
		}
	}
}

func (pt *ProxyTracker) buildCache() {
	pt.Lock()
	defer pt.Unlock()
	clear(pt.cache)
	for _, item := range pt.model.items {
		for _, name := range item.GetAlias() {
			pt.cache[name] = item.Proxy
		}
	}
}

func (pt *ProxyTracker) Close() {
	pt.model.RowsInserted().Detach(pt.rowsInsertedHandle)
	pt.model.BeforeRemove().Detach(pt.beforeRemoveHandle)
	pt.model.RowEdited().Detach(pt.rowEditedHandle)
	pt.model.RowRenamed().Detach(pt.rowRenamedHandle)
	pt.cancel()
	if pt.refreshTimer != nil {
		pt.refreshTimer.Stop()
		pt.refreshTimer = nil
	}
	if pt.ticker != nil {
		pt.ticker.Stop()
	}
}

// monitorLogFile monitors the frpc log file for proxy status changes
func (pt *ProxyTracker) monitorLogFile() {
	// Create ticker for periodic checks
	pt.ticker = time.NewTicker(2 * time.Second)
	defer pt.ticker.Stop()

	for {
		select {
		case <-pt.ctx.Done():
			return
		case <-pt.ticker.C:
			pt.processLogFile()
		}
	}
}

// processLogFile processes the log file to extract proxy status information
func (pt *ProxyTracker) processLogFile() {
	// Check if file exists
	if _, err := os.Stat(pt.logFilePath); os.IsNotExist(err) {
		return
	}

	// Open file
	file, err := os.Open(pt.logFilePath)
	if err != nil {
		return
	}
	defer file.Close()

	// Seek to last position
	if _, err := file.Seek(pt.lastLogPos, 0); err != nil {
		return
	}

	// Read new lines
	scanner := bufio.NewScanner(file)
	var newLines []string
	for scanner.Scan() {
		line := scanner.Text()
		pt.lastLogPos += int64(len(line) + 1) // +1 for newline
		newLines = append(newLines, line)
	}

	// Process new lines
	if len(newLines) > 0 {
		pt.processLogLines(newLines)
	}
}

// processLogLines processes multiple log lines to extract proxy status
func (pt *ProxyTracker) processLogLines(lines []string) {
	var messages []ipc.ProxyMessage

	for _, line := range lines {
		// Look for proxy status messages in log
		// Examples:
		// "proxy [tcp:ssh] starts successfully"
		// "proxy [http:web] starts successfully"
		// "proxy [tcp:ssh] error: connection refused"

		// Check for proxy start message
		if strings.Contains(line, "starts successfully") {
			msg := pt.extractProxyStatusFromLog(line, "running", "")
			if msg != nil {
				messages = append(messages, *msg)
			}
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
			msg := pt.extractProxyStatusFromLog(line, "error", errorMsg)
			if msg != nil {
				messages = append(messages, *msg)
			}
		}

		// Check for proxy stop messages
		if strings.Contains(line, "proxy stopped") {
			msg := pt.extractProxyStatusFromLog(line, "stopped", "")
			if msg != nil {
				messages = append(messages, *msg)
			}
		}
	}

	// Send messages to UI
	if len(messages) > 0 {
		pt.onMessage(messages)
	}
}

// extractProxyStatusFromLog extracts proxy status from a log line
func (pt *ProxyTracker) extractProxyStatusFromLog(line, status, errorMsg string) *ipc.ProxyMessage {
	// Extract proxy name and type from log line
	// Format: "proxy [type:name] message"
	start := strings.Index(line, "[")
	end := strings.Index(line, "]")
	if start == -1 || end == -1 {
		return nil
	}

	proxyInfo := line[start+1 : end]
	parts := strings.SplitN(proxyInfo, ":", 2)
	if len(parts) != 2 {
		return nil
	}

	proxyType := parts[0]
	proxyName := parts[1]

	// Create proxy message
	msg := ipc.ProxyMessage{
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

	return &msg
}

// checkProxyStatus checks the current status of all proxies
func (pt *ProxyTracker) checkProxyStatus() {
	// Check if service is running
	running, err := services.IsFrpcRunning(getCurrentConf().Path)
	if err != nil {
		return
	}

	var messages []ipc.ProxyMessage

	if !running {
		// Service is not running, all proxies are stopped
		for _, proxy := range pt.model.items {
			msg := ipc.ProxyMessage{
				Name:   proxy.Name,
				Type:   proxy.Type,
				Status: "stopped",
			}
			messages = append(messages, msg)
		}
	} else {
		// Service is running, try to get detailed status from logs
		messages = pt.getProxyStatusFromLogs()
		if len(messages) == 0 {
			// If no status from logs, assume all proxies are running
			for _, proxy := range pt.model.items {
				msg := ipc.ProxyMessage{
					Name:   proxy.Name,
					Type:   proxy.Type,
					Status: "running",
				}
				messages = append(messages, msg)
			}
		}
	}

	// Send messages to UI
	if len(messages) > 0 {
		pt.onMessage(messages)
	}
}

// getProxyStatusFromLogs retrieves proxy status by parsing the entire log file
func (pt *ProxyTracker) getProxyStatusFromLogs() []ipc.ProxyMessage {
	var messages []ipc.ProxyMessage

	// Check if file exists
	if _, err := os.Stat(pt.logFilePath); os.IsNotExist(err) {
		return messages
	}

	// Read log file
	file, err := os.Open(pt.logFilePath)
	if err != nil {
		return messages
	}
	defer file.Close()

	// Scan file for proxy status
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Extract proxy status from log line
		msg := pt.extractProxyStatusFromLog(line, "", "")
		if msg != nil && msg.Status != "" {
			messages = append(messages, *msg)
		}
	}

	return messages
}

// logStatusToProxyState converts log status string to ProxyState
func logStatusToProxyState(status string) consts.ProxyState {
	switch status {
	case "running":
		return consts.ProxyStateRunning
	case "stopped":
		return consts.ProxyStateStopped
	case "error":
		return consts.ProxyStateError
	default:
		return consts.ProxyStateUnknown
	}
}

// stringToProxyState converts string status to ProxyState
func stringToProxyState(status string) consts.ProxyState {
	return logStatusToProxyState(status)
}
