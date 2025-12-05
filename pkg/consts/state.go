package consts

// ConfigState is the state of FRP daemon service
type ConfigState int

const (
	ConfigStateUnknown ConfigState = iota
	ConfigStateStarted
	ConfigStateStopped
	ConfigStateStarting
	ConfigStateStopping
	ConfigStateNotInstalled
)

// ProxyState is the state of a proxy.
type ProxyState int

const (
	ProxyStateUnknown ProxyState = iota
	ProxyStateRunning
	ProxyStateError
	ProxyStateStopped
)
