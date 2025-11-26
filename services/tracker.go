package services

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc/mgr"

	"github.com/hzcrv1911/frpcgui/pkg/consts"
)

// Debug logging helper
func debugLog(function, format string, args ...interface{}) {
	if logFile, err := os.OpenFile(filepath.Join("logs", "tracker_debug.log"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		defer logFile.Close()
		msg := fmt.Sprintf(format, args...)
		fmt.Fprintf(logFile, "[%s] %s: %s\n", time.Now().Format("15:04:05.000"), function, msg)
	}
}

func configStateToString(state consts.ConfigState) string {
	switch state {
	case consts.ConfigStateUnknown:
		return "Unknown"
	case consts.ConfigStateStarted:
		return "Started"
	case consts.ConfigStateStopped:
		return "Stopped"
	case consts.ConfigStateStarting:
		return "Starting"
	case consts.ConfigStateStopping:
		return "Stopping"
	case consts.ConfigStateNotInstalled:
		return "NotInstalled"
	default:
		return fmt.Sprintf("Unknown(%d)", state)
	}
}

type ConfigStateCallback func(path string, state consts.ConfigState)

type tracker struct {
	service *mgr.Service
	stopCh  chan struct{}
}

var (
	trackedConfigs       = make(map[string]*tracker)
	trackedConfigsLock   = sync.Mutex{}
	cachedServiceManager *mgr.Mgr
)

func serviceManager() (*mgr.Mgr, error) {
	if cachedServiceManager != nil {
		return cachedServiceManager, nil
	}
	m, err := mgr.Connect()
	if err != nil {
		return nil, err
	}
	cachedServiceManager = m
	return cachedServiceManager, nil
}

func trackExistingConfigs(paths func() []string, cb ConfigStateCallback) error {
	m, err := serviceManager()
	if err != nil {
		return err
	}
	pathList := paths()
	pathSet := make(map[string]bool)
	for _, p := range pathList {
		pathSet[p] = true
	}

	trackedConfigsLock.Lock()
	defer trackedConfigsLock.Unlock()

	// 1. Remove trackers for paths that are no longer present
	for path, tr := range trackedConfigs {
		if !pathSet[path] {
			close(tr.stopCh)
			if tr.service != nil {
				tr.service.Close()
			}
			delete(trackedConfigs, path)
		}
	}

	// 2. Create or update trackers for current paths
	for _, path := range pathList {
		if _, exists := trackedConfigs[path]; exists {
			// Tracker already exists and is actively monitoring this config.
			// Don't re-query status here as the tracker goroutine is already
			// doing continuous monitoring and will report changes via callback.
			// Re-querying here can cause incorrect state updates due to timing issues.
			debugLog("trackExistingConfigs", "Tracker already exists for path=%s, skipping", path)
			continue
		}

		debugLog("trackExistingConfigs", "Creating new tracker for path=%s", path)

		// Create new tracker
		tr := &tracker{
			stopCh: make(chan struct{}),
		}

		if IsWinSWAvailable() {
			// WinSW mode
			tr.service = nil
			trackedConfigs[path] = tr
			go trackWinSWService(path, cb, tr)
		} else {
			// Native mode
			serviceName := ServiceNameOfClient(path)
			service, err := m.OpenService(serviceName)
			if err != nil {
				// Service not found, mark as not installed
				cb(path, consts.ConfigStateNotInstalled)
				continue
			}
			tr.service = service
			trackedConfigs[path] = tr
			go trackService(service, path, cb, tr)
		}
	}
	return nil
}

func WatchConfigServices(paths func() []string, cb ConfigStateCallback) (func() error, error) {
	m, err := serviceManager()
	if err != nil {
		return nil, err
	}
	var subscription uintptr
	err = windows.SubscribeServiceChangeNotifications(m.Handle, windows.SC_EVENT_DATABASE_CHANGE,
		windows.NewCallback(func(notification uint32, context uintptr) uintptr {
			trackExistingConfigs(paths, cb)
			return 0
		}), 0, &subscription)
	if err == nil {
		if err = trackExistingConfigs(paths, cb); err != nil {
			windows.UnsubscribeServiceChangeNotifications(subscription)
			return nil, err
		}
		return func() error {
			err := windows.UnsubscribeServiceChangeNotifications(subscription)
			trackedConfigsLock.Lock()
			for _, tr := range trackedConfigs {
				close(tr.stopCh)
				if tr.service != nil {
					tr.service.Close()
				}
			}
			// Clear map
			trackedConfigs = make(map[string]*tracker)
			trackedConfigsLock.Unlock()
			return err
		}, nil
	}
	return nil, err
}

func trackService(service *mgr.Service, path string, cb ConfigStateCallback, tr *tracker) {
	defer service.Close()

	var subscription uintptr
	lastState := consts.ConfigStateUnknown
	var updateState = func(state consts.ConfigState) {
		if state != lastState {
			cb(path, state)
			lastState = state
		}
	}
	err := windows.SubscribeServiceChangeNotifications(service.Handle, windows.SC_EVENT_STATUS_CHANGE,
		windows.NewCallback(func(notification uint32, context uintptr) uintptr {
			select {
			case <-tr.stopCh:
				return 0
			default:
			}

			configState := consts.ConfigStateUnknown
			if notification == 0 {
				status, err := service.Query()
				if err == nil {
					configState = svcStateToConfigState(uint32(status.State))
				}
			} else {
				configState = notifyStateToConfigState(notification)
			}
			updateState(configState)
			return 0
		}), 0, &subscription)
	if err == nil {
		defer windows.UnsubscribeServiceChangeNotifications(subscription)
		status, err := service.Query()
		if err == nil {
			updateState(svcStateToConfigState(uint32(status.State)))
		}
		<-tr.stopCh
	} else {
		cb(path, consts.ConfigStateStopped)
		// If we can't subscribe, maybe just poll or exit?
		// For now, just exit if subscription fails.
	}
}

// trackWinSWService tracks a WinSW-managed service
func trackWinSWService(path string, cb ConfigStateCallback, tr *tracker) {
	// Get paths
	winSWPath, err := GetWinSWPath()
	if err != nil {
		cb(path, consts.ConfigStateNotInstalled)
		return
	}

	// Create log directory
	logPath := filepath.Join(filepath.Dir(path), "logs")

	// Create WinSW service
	serviceName := ServiceNameOfClient(path)
	wsService := NewWinSWService(serviceName, path, winSWPath, "", logPath)

	lastState := consts.ConfigStateUnknown
	var updateState = func(state consts.ConfigState) {
		if state != lastState {
			// DEBUG: Log state changes
			debugLog("trackWinSWService", "path=%s, serviceName=%s, oldState=%d, newState=%d, status=%s",
				path, serviceName, lastState, state, configStateToString(state))
			cb(path, state)
			lastState = state
		}
	}

	// Check initial status
	status, err := wsService.Status()
	debugLog("trackWinSWService", "Initial status check: path=%s, serviceName=%s, status=%s, err=%v",
		path, serviceName, status, err)
	if err != nil {
		// Error checking status - mark as not installed
		updateState(consts.ConfigStateNotInstalled)
	} else {
		updateState(winSWStatusToConfigState(status))
	}

	// Poll service status periodically
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-tr.stopCh:
			return
		case <-ticker.C:
			status, err := wsService.Status()
			if err != nil {
				debugLog("trackWinSWService", "Status query error: path=%s, serviceName=%s, err=%v",
					path, serviceName, err)
				updateState(consts.ConfigStateNotInstalled)
				continue
			}
			updateState(winSWStatusToConfigState(status))
		}
	}
}

func winSWStatusToConfigState(status string) consts.ConfigState {
	switch status {
	case "Running":
		return consts.ConfigStateStarted
	case "Starting":
		return consts.ConfigStateStarting
	case "Stopping":
		return consts.ConfigStateStopping
	case "Stopped":
		return consts.ConfigStateStopped
	case "Not installed":
		return consts.ConfigStateNotInstalled
	default:
		return consts.ConfigStateUnknown
	}
}

func svcStateToConfigState(s uint32) consts.ConfigState {
	switch s {
	case windows.SERVICE_STOPPED:
		return consts.ConfigStateStopped
	case windows.SERVICE_START_PENDING:
		return consts.ConfigStateStarting
	case windows.SERVICE_STOP_PENDING:
		return consts.ConfigStateStopping
	case windows.SERVICE_RUNNING:
		return consts.ConfigStateStarted
	case windows.SERVICE_NO_CHANGE:
		return 0
	default:
		return 0
	}
}

func notifyStateToConfigState(s uint32) consts.ConfigState {
	if s&(windows.SERVICE_NOTIFY_STOPPED|windows.SERVICE_NOTIFY_DELETED|windows.SERVICE_NOTIFY_DELETE_PENDING) != 0 {
		return consts.ConfigStateStopped
	} else if s&windows.SERVICE_NOTIFY_STOP_PENDING != 0 {
		return consts.ConfigStateStopping
	} else if s&windows.SERVICE_NOTIFY_RUNNING != 0 {
		return consts.ConfigStateStarted
	} else if s&windows.SERVICE_NOTIFY_START_PENDING != 0 {
		return consts.ConfigStateStarting
	} else {
		return consts.ConfigStateUnknown
	}
}
