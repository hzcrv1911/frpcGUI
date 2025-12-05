package config

import (
	"path/filepath"
	"testing"
)

func TestClientConfigSaveTOML(t *testing.T) {
	conf := NewDefaultClientConfig()
	conf.ClientCommon.Name = "test"
	conf.ClientCommon.ServerAddress = "example.com"
	conf.ClientCommon.Token = "token"
	conf.ClientCommon.ServerPort = 7000
	conf.Complete(false)

	path := filepath.Join(t.TempDir(), "test.conf")
	if err := conf.Save(path); err != nil {
		t.Fatalf("%T: %v", err, err)
	}
}
