package version

// Version represents the application version
var Version = "v0.0.0-dev"

// GetVersion returns the current version of the application
func GetVersion() string {
	return Version
}

// GetFrpcVersion returns the version of the frpc.exe
// This will be determined at runtime by checking the frpc.exe file
func GetFrpcVersion() (string, error) {
	// This function will be implemented in services/frp.go
	// as it needs access to the frpc.exe path
	return "", nil
}

var (
	Number = "1.0.1"
	// FRPVersion is the version of FRP used by this program
	FRPVersion = ""
	// BuildDate is the day that this program was built
	BuildDate = ""
)
