package config

import (
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestUnmarshalClientConfFromIni(t *testing.T) {
	input := `
		[common]
		server_addr = example.com
		server_port = 7001
		token = 123456
		frpcgui_manual_start = true
		frpcgui_delete_method = absolute
		frpcgui_delete_after_date = 2023-03-23T00:00:00Z
		meta_1 = value
		
		[ssh]
		type = tcp
		local_ip = 192.168.1.1
		local_port = 22
		remote_port = 6000
		meta_2 = value
	`
	expected := NewDefaultClientConfig()
	expected.LegacyFormat = true
	expected.ServerAddress = "example.com"
	expected.ServerPort = 7001
	expected.Token = "123456"
	expected.ManualStart = true
	expected.Metas = map[string]string{"1": "value"}
	expected.DeleteMethod = "absolute"
	expected.DeleteAfterDate = time.Date(2023, 3, 23, 0, 0, 0, 0, time.UTC)
	expected.Proxies = append(expected.Proxies, &Proxy{
		BaseProxyConf: BaseProxyConf{
			Name:      "ssh",
			Type:      "tcp",
			LocalIP:   "192.168.1.1",
			LocalPort: "22",
			Metas:     map[string]string{"2": "value"},
		},
		RemotePort: "6000",
	})
	cc, err := UnmarshalClientConfFromIni([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cc, expected) {
		t.Errorf("Expected: %v, got: %v", expected, cc)
	}
}

func TestProxyGetAlias(t *testing.T) {
	input := `
		[range:test_tcp]
		type = tcp
		local_ip = 127.0.0.1
		local_port = 6000-6006,6007
		remote_port = 6000-6006,6007
	`
	expected := []string{"test_tcp_0", "test_tcp_1", "test_tcp_2", "test_tcp_3",
		"test_tcp_4", "test_tcp_5", "test_tcp_6", "test_tcp_7"}
	proxy, err := UnmarshalProxyFromIni([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	output := proxy.GetAlias()
	if !reflect.DeepEqual(output, expected) {
		t.Errorf("Expected: %v, got: %v", expected, output)
	}
}

func TestClientConfigSaveTOML(t *testing.T) {
	conf := NewDefaultClientConfig()
	conf.LegacyFormat = false
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
