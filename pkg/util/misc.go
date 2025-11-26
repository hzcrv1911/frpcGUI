package util

import (
	"fmt"
	"os/exec"
	"reflect"
	"strings"
)

// PruneByTag returns a copy of "in" that only contains fields with the given tag and value
func PruneByTag(in interface{}, value string, tag string) (interface{}, error) {
	inValue := reflect.ValueOf(in)

	ret := reflect.New(inValue.Type()).Elem()

	if err := prune(inValue, ret, value, tag); err != nil {
		return nil, err
	}
	return ret.Interface(), nil
}

func prune(inValue reflect.Value, ret reflect.Value, value string, tag string) error {
	switch inValue.Kind() {
	case reflect.Ptr:
		if inValue.IsNil() {
			return nil
		}
		if ret.IsNil() {
			// init ret and go to next level
			ret.Set(reflect.New(inValue.Type().Elem()))
		}
		return prune(inValue.Elem(), ret.Elem(), value, tag)
	case reflect.Struct:
		var fValue reflect.Value
		var fRet reflect.Value
		// search tag that has key equal to value
		for i := 0; i < inValue.NumField(); i++ {
			f := inValue.Type().Field(i)
			if key, ok := f.Tag.Lookup(tag); ok {
				if key == "*" || key == value {
					fValue = inValue.Field(i)
					fRet = ret.Field(i)
					fRet.Set(fValue)
				}
			}
		}
	}
	return nil
}

func GetMapWithoutPrefix(set map[string]string, prefix string) map[string]string {
	m := make(map[string]string)

	for key, value := range set {
		if strings.HasPrefix(key, prefix) {
			m[strings.TrimPrefix(key, prefix)] = value
		}
	}

	if len(m) == 0 {
		return nil
	}

	return m
}

// MoveSlice moves the element s[i] to index j in s.
func MoveSlice[S ~[]E, E any](s S, i, j int) {
	x := s[i]
	if i < j {
		copy(s[i:j], s[i+1:j+1])
	} else if i > j {
		copy(s[j+1:i+1], s[j:i])
	}
	s[j] = x
}

// ByteCountIEC converts a size in bytes to a human-readable string in IEC (binary) format.
func ByteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}

// ExecuteCommand executes a command in the system shell
func ExecuteCommand(command string) error {
	// Simple implementation using os/exec
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	cmd := parts[0]
	args := parts[1:]

	return executeCommand(cmd, args...)
}

// executeCommand is the internal implementation that can be mocked for testing
var executeCommand = func(cmd string, args ...string) error {
	c := exec.Command(cmd, args...)
	return c.Run()
}

// ExecuteCommandWithOutput executes a command and returns its output
func ExecuteCommandWithOutput(command string) (string, error) {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "", fmt.Errorf("empty command")
	}

	cmd := parts[0]
	args := parts[1:]

	c := exec.Command(cmd, args...)
	output, err := c.CombinedOutput()
	if err != nil {
		return "", err
	}

	return string(output), nil
}

// NewError creates a new error with the given message
func NewError(message string) error {
	return fmt.Errorf("%s", message)
}
