package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
	"github.com/samber/lo"
	"gopkg.in/ini.v1"

	"github.com/hzcrv1911/frpcgui/pkg/consts"
	"github.com/hzcrv1911/frpcgui/pkg/util"
)

type ClientAuth struct {
	AuthMethod                   string            `ini:"authentication_method,omitempty"`
	AuthenticateHeartBeats       bool              `ini:"authenticate_heartbeats,omitempty" token:"true" oidc:"true"`
	AuthenticateNewWorkConns     bool              `ini:"authenticate_new_work_conns,omitempty" token:"true" oidc:"true"`
	Token                        string            `ini:"token,omitempty" token:"true"`
	TokenSource                  string            `ini:"-" token:"true"`
	TokenSourceFile              string            `ini:"-" token:"true"`
	OIDCClientId                 string            `ini:"oidc_client_id,omitempty" oidc:"true"`
	OIDCClientSecret             string            `ini:"oidc_client_secret,omitempty" oidc:"true"`
	OIDCAudience                 string            `ini:"oidc_audience,omitempty" oidc:"true"`
	OIDCScope                    string            `ini:"oidc_scope,omitempty" oidc:"true"`
	OIDCTokenEndpoint            string            `ini:"oidc_token_endpoint_url,omitempty" oidc:"true"`
	OIDCAdditionalEndpointParams map[string]string `ini:"-" oidc:"true"`
}

func (ca ClientAuth) Complete() ClientAuth {
	authMethod := ca.AuthMethod
	if authMethod != "" {
		if auth, err := util.PruneByTag(ca, "true", authMethod); err == nil {
			ca = auth.(ClientAuth)
			ca.AuthMethod = authMethod
		}
		if authMethod == consts.AuthToken {
			if ca.TokenSource != "" {
				ca.Token = ""
			} else {
				ca.TokenSourceFile = ""
				if ca.Token == "" {
					ca.AuthMethod = ""
				}
			}
		}
	} else {
		ca = ClientAuth{}
	}
	return ca
}

type ClientCommon struct {
	ClientAuth                `ini:",extends"`
	ServerAddress             string   `ini:"server_addr,omitempty"`
	ServerPort                int      `ini:"server_port,omitempty"`
	NatHoleSTUNServer         string   `ini:"nat_hole_stun_server,omitempty"`
	DialServerTimeout         int64    `ini:"dial_server_timeout,omitempty"`
	DialServerKeepAlive       int64    `ini:"dial_server_keepalive,omitempty"`
	ConnectServerLocalIP      string   `ini:"connect_server_local_ip,omitempty"`
	HTTPProxy                 string   `ini:"http_proxy,omitempty"`
	LogFile                   string   `ini:"log_file,omitempty"`
	LogLevel                  string   `ini:"log_level,omitempty"`
	LogMaxDays                int64    `ini:"log_max_days,omitempty"`
	AdminAddr                 string   `ini:"admin_addr,omitempty"`
	AdminPort                 int      `ini:"admin_port,omitempty"`
	AdminUser                 string   `ini:"admin_user,omitempty"`
	AdminPwd                  string   `ini:"admin_pwd,omitempty"`
	AssetsDir                 string   `ini:"assets_dir,omitempty"`
	PoolCount                 int      `ini:"pool_count,omitempty"`
	DNSServer                 string   `ini:"dns_server,omitempty"`
	Protocol                  string   `ini:"protocol,omitempty"`
	QUICKeepalivePeriod       int      `ini:"quic_keepalive_period,omitempty"`
	QUICMaxIdleTimeout        int      `ini:"quic_max_idle_timeout,omitempty"`
	QUICMaxIncomingStreams    int      `ini:"quic_max_incoming_streams,omitempty"`
	LoginFailExit             bool     `ini:"login_fail_exit"`
	User                      string   `ini:"user,omitempty"`
	HeartbeatInterval         int64    `ini:"heartbeat_interval,omitempty"`
	HeartbeatTimeout          int64    `ini:"heartbeat_timeout,omitempty"`
	TCPMux                    bool     `ini:"tcp_mux"`
	TCPMuxKeepaliveInterval   int64    `ini:"tcp_mux_keepalive_interval,omitempty"`
	TLSEnable                 bool     `ini:"tls_enable"`
	TLSCertFile               string   `ini:"tls_cert_file,omitempty"`
	TLSKeyFile                string   `ini:"tls_key_file,omitempty"`
	TLSTrustedCaFile          string   `ini:"tls_trusted_ca_file,omitempty"`
	TLSServerName             string   `ini:"tls_server_name,omitempty"`
	UDPPacketSize             int64    `ini:"udp_packet_size,omitempty"`
	Start                     []string `ini:"start,omitempty"`
	PprofEnable               bool     `ini:"pprof_enable,omitempty"`
	DisableCustomTLSFirstByte bool     `ini:"disable_custom_tls_first_byte"`

	// Name of this config.
	Name string `ini:"frpcgui_name"`
	// ManualStart defines whether to start the config on system boot.
	ManualStart bool `ini:"frpcgui_manual_start,omitempty"`
	// AutoDelete is a mechanism for temporary use.
	// The config will be stopped and deleted at some point.
	AutoDelete `ini:",extends"`
	// Client meta info
	Metas map[string]string `ini:"-"`
	// Config file format
	LegacyFormat bool `ini:"-"`
}

// BaseProxyConf provides configuration info that is common to all types.
type BaseProxyConf struct {
	// Name is the name of this proxy.
	Name string `ini:"-"`
	// Type specifies the type of this. Valid values include tcp, udp,
	// xtcp, stcp, sudp, http, https, tcpmux. By default, this value is "tcp".
	Type string `ini:"type,omitempty"`

	// UseEncryption controls whether communication with the server will
	// be encrypted. Encryption is done using the tokens supplied in the server
	// and client configuration. By default, this value is false.
	UseEncryption bool `ini:"use_encryption,omitempty"`
	// UseCompression controls whether communication with the server
	// will be compressed. By default, this value is false.
	UseCompression bool `ini:"use_compression,omitempty"`
	// Group specifies which group the proxy is a part of. The server will use
	// this information to load balance proxies in the same group. If the value
	// is "", this will not be in a group. By default, this value is "".
	Group string `ini:"group,omitempty"`
	// GroupKey specifies a group key, which should be the same among proxies
	// of the same group. By default, this value is "".
	GroupKey string `ini:"group_key,omitempty"`

	// ProxyProtocolVersion specifies which protocol version to use. Valid
	// values include "v1", "v2", and "". If the value is "", a protocol
	// version will be automatically selected. By default, this value is "".
	ProxyProtocolVersion string `ini:"proxy_protocol_version,omitempty"`

	// BandwidthLimit limits the bandwidth.
	// 0 means no limit.
	BandwidthLimit     string `ini:"bandwidth_limit,omitempty"`
	BandwidthLimitMode string `ini:"bandwidth_limit_mode,omitempty"`

	// LocalIP specifies the IP address or host name.
	LocalIP string `ini:"local_ip,omitempty"`
	// LocalPort specifies the port.
	LocalPort string `ini:"local_port,omitempty"`

	// Plugin specifies what plugin should be used for ng. If this value
	// is set, the LocalIp and LocalPort values will be ignored. By default,
	// this value is "".
	Plugin string `ini:"plugin,omitempty"`
	// PluginParams specify parameters to be passed to the plugin, if one is
	// being used.
	PluginParams `ini:",extends"`
	// HealthCheckType specifies what protocol to use for health checking.
	HealthCheckType string `ini:"health_check_type,omitempty"` // tcp | http
	// Health checking parameters.
	HealthCheckConf `ini:",extends"`
	// Meta info for each proxy
	Metas map[string]string `ini:"-"`
	// Annotations for each proxy
	Annotations map[string]string `ini:"-"`
	// Disabled defines whether to start the proxy.
	Disabled bool `ini:"-"`
}

type PluginParams struct {
	PluginLocalAddr         string            `ini:"plugin_local_addr,omitempty" http2https:"true" http2http:"true" https2https:"true" https2http:"true" tls2raw:"true"`
	PluginCrtPath           string            `ini:"plugin_crt_path,omitempty" https2https:"true" https2http:"true" tls2raw:"true"`
	PluginKeyPath           string            `ini:"plugin_key_path,omitempty" https2https:"true" https2http:"true" tls2raw:"true"`
	PluginHostHeaderRewrite string            `ini:"plugin_host_header_rewrite,omitempty" http2https:"true" http2http:"true" https2https:"true" https2http:"true"`
	PluginHttpUser          string            `ini:"plugin_http_user,omitempty" http_proxy:"true" static_file:"true"`
	PluginHttpPasswd        string            `ini:"plugin_http_passwd,omitempty" http_proxy:"true" static_file:"true"`
	PluginUser              string            `ini:"plugin_user,omitempty" socks5:"true"`
	PluginPasswd            string            `ini:"plugin_passwd,omitempty" socks5:"true"`
	PluginLocalPath         string            `ini:"plugin_local_path,omitempty" static_file:"true"`
	PluginStripPrefix       string            `ini:"plugin_strip_prefix,omitempty" static_file:"true"`
	PluginUnixPath          string            `ini:"plugin_unix_path,omitempty" unix_domain_socket:"true"`
	PluginHeaders           map[string]string `ini:"-" http2https:"true" http2http:"true" https2https:"true" https2http:"true"`
	PluginEnableHTTP2       bool              `ini:"-" https2https:"true" https2http:"true"`
}

// HealthCheckConf configures health checking. This can be useful for load
// balancing purposes to detect and remove proxies to failing services.
type HealthCheckConf struct {
	// HealthCheckTimeoutS specifies the number of seconds to wait for a health
	// check attempt to connect. If the timeout is reached, this counts as a
	// health check failure. By default, this value is 3.
	HealthCheckTimeoutS int `ini:"health_check_timeout_s,omitempty" tcp:"true" http:"true"`
	// HealthCheckMaxFailed specifies the number of allowed failures before the
	// is stopped. By default, this value is 1.
	HealthCheckMaxFailed int `ini:"health_check_max_failed,omitempty" tcp:"true" http:"true"`
	// HealthCheckIntervalS specifies the time in seconds between health
	// checks. By default, this value is 10.
	HealthCheckIntervalS int `ini:"health_check_interval_s,omitempty" tcp:"true" http:"true"`
	// HealthCheckURL specifies the address to send health checks to if the
	// health check type is "http".
	HealthCheckURL string `ini:"health_check_url,omitempty" http:"true"`
	// HealthCheckHTTPHeaders specifies the headers to send with the http request.
	HealthCheckHTTPHeaders map[string]string `ini:"-" http:"true"`
}

type Proxy struct {
	BaseProxyConf     `ini:",extends"`
	RemotePort        string            `ini:"remote_port,omitempty" tcp:"true" udp:"true"`
	Role              string            `ini:"role,omitempty" stcp:"true" xtcp:"true" sudp:"true" visitor:"*"`
	SK                string            `ini:"sk,omitempty" stcp:"true" xtcp:"true" sudp:"true" visitor:"*"`
	AllowUsers        string            `ini:"allow_users,omitempty" stcp:"true" xtcp:"true" sudp:"true"`
	ServerUser        string            `ini:"server_user,omitempty" visitor:"*"`
	ServerName        string            `ini:"server_name,omitempty" visitor:"*"`
	BindAddr          string            `ini:"bind_addr,omitempty" visitor:"*"`
	BindPort          int               `ini:"bind_port,omitempty" visitor:"*"`
	CustomDomains     string            `ini:"custom_domains,omitempty" http:"true" https:"true" tcpmux:"true"`
	SubDomain         string            `ini:"subdomain,omitempty" http:"true" https:"true" tcpmux:"true"`
	Locations         string            `ini:"locations,omitempty" http:"true"`
	HTTPUser          string            `ini:"http_user,omitempty" http:"true" tcpmux:"true"`
	HTTPPwd           string            `ini:"http_pwd,omitempty" http:"true" tcpmux:"true"`
	HostHeaderRewrite string            `ini:"host_header_rewrite,omitempty" http:"true"`
	Headers           map[string]string `ini:"-" http:"true"`
	ResponseHeaders   map[string]string `ini:"-" http:"true"`
	Multiplexer       string            `ini:"multiplexer,omitempty" tcpmux:"true"`
	RouteByHTTPUser   string            `ini:"route_by_http_user,omitempty" http:"true" tcpmux:"true"`
	// "kcp" or "quic"
	Protocol          string `ini:"protocol,omitempty" visitor:"xtcp"`
	KeepTunnelOpen    bool   `ini:"keep_tunnel_open,omitempty" visitor:"xtcp"`
	MaxRetriesAnHour  int    `ini:"max_retries_an_hour,omitempty" visitor:"xtcp"`
	MinRetryInterval  int    `ini:"min_retry_interval,omitempty" visitor:"xtcp"`
	FallbackTo        string `ini:"fallback_to,omitempty" visitor:"xtcp"`
	FallbackTimeoutMs int    `ini:"fallback_timeout_ms,omitempty" visitor:"xtcp"`
}

// GetAlias returns the alias of this proxy.
// It's usually equal to the proxy name, but proxies that start with "range:" differ from it.
func (p *Proxy) GetAlias() []string {
	if p.IsRange() {
		localPorts, err := parseRangeNumbers(p.LocalPort)
		if err != nil {
			return []string{p.Name}
		}
		alias := make([]string, len(localPorts))
		for i := range localPorts {
			alias[i] = fmt.Sprintf("%s_%d", p.Name, i)
		}
		return alias
	}
	return []string{p.Name}
}

// parseRangeNumbers parses a range string like "1000-1002,1004" into individual numbers
func parseRangeNumbers(rangeStr string) ([]int, error) {
	if rangeStr == "" {
		return nil, fmt.Errorf("empty range string")
	}

	var result []int
	parts := strings.Split(rangeStr, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if strings.Contains(part, "-") {
			// Handle range like "1000-1002"
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("invalid range format: %s", part)
			}

			start, err := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
			if err != nil {
				return nil, fmt.Errorf("invalid start number in range: %s", rangeParts[0])
			}

			end, err := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
			if err != nil {
				return nil, fmt.Errorf("invalid end number in range: %s", rangeParts[1])
			}

			if start > end {
				return nil, fmt.Errorf("start number must be less than or equal to end number in range: %s", part)
			}

			for i := start; i <= end; i++ {
				result = append(result, i)
			}
		} else {
			// Handle single number
			num, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("invalid number: %s", part)
			}
			result = append(result, num)
		}
	}

	return result, nil
}

// IsVisitor returns a boolean indicating whether the proxy has a visitor role.
func (p *Proxy) IsVisitor() bool {
	return (p.Type == consts.ProxyTypeXTCP ||
		p.Type == consts.ProxyTypeSTCP ||
		p.Type == consts.ProxyTypeSUDP) && p.Role == "visitor"
}

func (p *Proxy) IsRange() bool {
	return (p.Type == consts.ProxyTypeTCP || p.Type == consts.ProxyTypeUDP) &&
		lo.Some([]rune(p.LocalPort+p.RemotePort), []rune{',', '-'})
}

// Complete removes redundant parameters base on the proxy type.
func (p *Proxy) Complete() {
	var base = p.BaseProxyConf
	if p.IsVisitor() {
		// Visitor
		if vp, err := util.PruneByTag(*p, p.Type, "visitor"); err == nil {
			*p = vp.(Proxy)
		}
		p.BaseProxyConf = BaseProxyConf{
			Name: base.Name, Type: base.Type, UseEncryption: base.UseEncryption,
			UseCompression: base.UseCompression, Disabled: base.Disabled,
		}
		// Reset xtcp visitor parameters
		if !p.KeepTunnelOpen {
			p.MaxRetriesAnHour = 0
			p.MinRetryInterval = 0
		}
		if p.FallbackTo == "" {
			p.FallbackTimeoutMs = 0
		}
	} else {
		// Plugins
		if base.Plugin != "" {
			base.LocalIP = ""
			base.LocalPort = ""
			if pluginParams, err := util.PruneByTag(base.PluginParams, "true", base.Plugin); err == nil {
				base.PluginParams = pluginParams.(PluginParams)
			}
		} else {
			base.PluginParams = PluginParams{}
		}
		// Health Check
		if base.HealthCheckType != "" {
			if healthCheckConf, err := util.PruneByTag(base.HealthCheckConf, "true", base.HealthCheckType); err == nil {
				base.HealthCheckConf = healthCheckConf.(HealthCheckConf)
			}
		} else {
			base.HealthCheckConf = HealthCheckConf{}
		}
		// Proxy type
		if typedProxy, err := util.PruneByTag(*p, "true", p.Type); err == nil {
			*p = typedProxy.(Proxy)
		}
		p.BaseProxyConf = base
	}
}

type ClientConfig struct {
	ClientCommon
	Proxies []*Proxy
}

// Name of this config.
func (conf *ClientConfig) Name() string {
	return conf.ClientCommon.Name
}

// AutoStart indicates whether this config should be started at boot.
func (conf *ClientConfig) AutoStart() bool {
	return !conf.ManualStart
}

func (conf *ClientConfig) DeleteProxy(index int) {
	conf.Proxies = append(conf.Proxies[:index], conf.Proxies[index+1:]...)
}

func (conf *ClientConfig) AddProxy(proxy *Proxy) {
	conf.Proxies = append(conf.Proxies, proxy)
}

func (conf *ClientConfig) Save(path string) error {
	if conf.LegacyFormat {
		return conf.saveINI(path)
	} else {
		return conf.saveTOML(path)
	}
}

func (conf *ClientConfig) saveINI(path string) error {
	cfg := ini.Empty()
	common, err := cfg.NewSection("common")
	if err != nil {
		return err
	}
	if err = common.ReflectFrom(&conf.ClientCommon); err != nil {
		return err
	}
	for k, v := range conf.Metas {
		common.Key("meta_" + k).SetValue(v)
	}
	for k, v := range conf.OIDCAdditionalEndpointParams {
		common.Key("oidc_additional_" + k).SetValue(v)
	}
	for _, proxy := range conf.Proxies {
		name := proxy.Name
		if proxy.IsRange() && !strings.HasPrefix(name, consts.RangePrefix) {
			name = consts.RangePrefix + name
		}
		p, err := cfg.NewSection(name)
		if err != nil {
			return err
		}
		if err = p.ReflectFrom(&proxy); err != nil {
			return err
		}
		for k, v := range proxy.Metas {
			p.Key("meta_" + k).SetValue(v)
		}
		for k, v := range proxy.Headers {
			p.Key("header_" + k).SetValue(v)
		}
		for k, v := range proxy.PluginHeaders {
			p.Key("plugin_header_" + k).SetValue(v)
		}
	}
	return cfg.SaveTo(path)
}

func (conf *ClientConfig) saveTOML(path string) error {
	// Create a simple TOML structure for the config
	tomlData := make(map[string]interface{})

	// Add common configuration
	common := make(map[string]interface{})
	common["server_addr"] = conf.ServerAddress
	common["server_port"] = conf.ServerPort
	common["token"] = conf.Token
	if conf.User != "" {
		common["user"] = conf.User
	}
	if conf.LogFile != "" {
		common["log_file"] = conf.LogFile
	}
	if conf.LogLevel != "" {
		common["log_level"] = conf.LogLevel
	}
	if conf.LogMaxDays > 0 {
		common["log_max_days"] = conf.LogMaxDays
	}
	if conf.AdminAddr != "" {
		common["admin_addr"] = conf.AdminAddr
	}
	if conf.AdminPort > 0 {
		common["admin_port"] = conf.AdminPort
	}
	if conf.AdminUser != "" {
		common["admin_user"] = conf.AdminUser
	}
	if conf.AdminPwd != "" {
		common["admin_pwd"] = conf.AdminPwd
	}
	if conf.TLSEnable {
		common["tls_enable"] = true
		if conf.TLSCertFile != "" {
			common["tls_cert_file"] = conf.TLSCertFile
		}
		if conf.TLSKeyFile != "" {
			common["tls_key_file"] = conf.TLSKeyFile
		}
		if conf.TLSTrustedCaFile != "" {
			common["tls_trusted_ca_file"] = conf.TLSTrustedCaFile
		}
		if conf.TLSServerName != "" {
			common["tls_server_name"] = conf.TLSServerName
		}
	}
	if conf.Protocol != "" {
		common["protocol"] = conf.Protocol
	}
	if conf.TCPMux {
		common["tcp_mux"] = true
	}
	if conf.HeartbeatInterval > 0 {
		common["heartbeat_interval"] = conf.HeartbeatInterval
	}
	if conf.HeartbeatTimeout > 0 {
		common["heartbeat_timeout"] = conf.HeartbeatTimeout
	}
	if conf.UDPPacketSize > 0 {
		common["udp_packet_size"] = conf.UDPPacketSize
	}

	// Add manager-specific fields
	common["frpcgui_name"] = conf.Name()
	if conf.ManualStart {
		common["frpcgui_manual_start"] = true
	}

	// Add auto-delete configuration if needed
	if conf.AutoDelete.DeleteMethod != "" {
		common["frpcgui_delete_method"] = conf.AutoDelete.DeleteMethod
		if conf.AutoDelete.DeleteMethod == consts.DeleteAbsolute && !conf.AutoDelete.DeleteAfterDate.IsZero() {
			common["frpcgui_delete_after_date"] = conf.AutoDelete.DeleteAfterDate.Format("2006-01-02T15:04:05Z")
		} else if conf.AutoDelete.DeleteMethod == consts.DeleteRelative && conf.AutoDelete.DeleteAfterDays > 0 {
			common["frpcgui_delete_after_days"] = conf.AutoDelete.DeleteAfterDays
		}
	}

	// Add meta information
	for k, v := range conf.Metas {
		common["meta_"+k] = v
	}

	// Add OIDC additional endpoint params
	for k, v := range conf.OIDCAdditionalEndpointParams {
		common["oidc_additional_"+k] = v
	}

	tomlData["common"] = common

	// Add proxies
	for _, proxy := range conf.Proxies {
		proxyData := make(map[string]interface{})
		proxyData["type"] = proxy.Type
		if proxy.LocalIP != "" {
			proxyData["local_ip"] = proxy.LocalIP
		}
		if proxy.LocalPort != "" {
			proxyData["local_port"] = proxy.LocalPort
		}
		if proxy.RemotePort != "" {
			proxyData["remote_port"] = proxy.RemotePort
		}
		if proxy.UseEncryption {
			proxyData["use_encryption"] = true
		}
		if proxy.UseCompression {
			proxyData["use_compression"] = true
		}
		if proxy.Group != "" {
			proxyData["group"] = proxy.Group
		}
		if proxy.GroupKey != "" {
			proxyData["group_key"] = proxy.GroupKey
		}

		// Add proxy-specific fields based on type
		switch proxy.Type {
		case consts.ProxyTypeHTTP, consts.ProxyTypeHTTPS:
			if proxy.CustomDomains != "" {
				proxyData["custom_domains"] = proxy.CustomDomains
			}
			if proxy.SubDomain != "" {
				proxyData["subdomain"] = proxy.SubDomain
			}
			if proxy.Locations != "" {
				proxyData["locations"] = proxy.Locations
			}
			if proxy.HTTPUser != "" {
				proxyData["http_user"] = proxy.HTTPUser
			}
			if proxy.HTTPPwd != "" {
				proxyData["http_pwd"] = proxy.HTTPPwd
			}
			if proxy.HostHeaderRewrite != "" {
				proxyData["host_header_rewrite"] = proxy.HostHeaderRewrite
			}
			if proxy.Multiplexer != "" {
				proxyData["multiplexer"] = proxy.Multiplexer
			}
		case consts.ProxyTypeTCP, consts.ProxyTypeUDP:
			// Already handled remote_port above
		case consts.ProxyTypeSTCP, consts.ProxyTypeXTCP, consts.ProxyTypeSUDP:
			if proxy.Role != "" {
				proxyData["role"] = proxy.Role
			}
			if proxy.SK != "" {
				proxyData["sk"] = proxy.SK
			}
			if proxy.AllowUsers != "" {
				proxyData["allow_users"] = proxy.AllowUsers
			}
			if proxy.IsVisitor() {
				if proxy.ServerUser != "" {
					proxyData["server_user"] = proxy.ServerUser
				}
				if proxy.ServerName != "" {
					proxyData["server_name"] = proxy.ServerName
				}
				if proxy.BindAddr != "" {
					proxyData["bind_addr"] = proxy.BindAddr
				}
				if proxy.BindPort > 0 {
					proxyData["bind_port"] = proxy.BindPort
				}
				if proxy.KeepTunnelOpen {
					proxyData["keep_tunnel_open"] = true
				}
				if proxy.MaxRetriesAnHour > 0 {
					proxyData["max_retries_an_hour"] = proxy.MaxRetriesAnHour
				}
				if proxy.MinRetryInterval > 0 {
					proxyData["min_retry_interval"] = proxy.MinRetryInterval
				}
				if proxy.FallbackTo != "" {
					proxyData["fallback_to"] = proxy.FallbackTo
				}
				if proxy.FallbackTimeoutMs > 0 {
					proxyData["fallback_timeout_ms"] = proxy.FallbackTimeoutMs
				}
			}
		}

		// Add plugin configuration
		if proxy.Plugin != "" {
			proxyData["plugin"] = proxy.Plugin
			if proxy.PluginLocalAddr != "" {
				proxyData["plugin_local_addr"] = proxy.PluginLocalAddr
			}
			if proxy.PluginCrtPath != "" {
				proxyData["plugin_crt_path"] = proxy.PluginCrtPath
			}
			if proxy.PluginKeyPath != "" {
				proxyData["plugin_key_path"] = proxy.PluginKeyPath
			}
			if proxy.PluginHostHeaderRewrite != "" {
				proxyData["plugin_host_header_rewrite"] = proxy.PluginHostHeaderRewrite
			}
			if proxy.PluginHttpUser != "" {
				proxyData["plugin_http_user"] = proxy.PluginHttpUser
			}
			if proxy.PluginHttpPasswd != "" {
				proxyData["plugin_http_passwd"] = proxy.PluginHttpPasswd
			}
			if proxy.PluginUser != "" {
				proxyData["plugin_user"] = proxy.PluginUser
			}
			if proxy.PluginPasswd != "" {
				proxyData["plugin_passwd"] = proxy.PluginPasswd
			}
			if proxy.PluginLocalPath != "" {
				proxyData["plugin_local_path"] = proxy.PluginLocalPath
			}
			if proxy.PluginStripPrefix != "" {
				proxyData["plugin_strip_prefix"] = proxy.PluginStripPrefix
			}
			if proxy.PluginUnixPath != "" {
				proxyData["plugin_unix_path"] = proxy.PluginUnixPath
			}
		}

		// Add health check configuration
		if proxy.HealthCheckType != "" {
			proxyData["health_check_type"] = proxy.HealthCheckType
			if proxy.HealthCheckTimeoutS > 0 {
				proxyData["health_check_timeout_s"] = proxy.HealthCheckTimeoutS
			}
			if proxy.HealthCheckMaxFailed > 0 {
				proxyData["health_check_max_failed"] = proxy.HealthCheckMaxFailed
			}
			if proxy.HealthCheckIntervalS > 0 {
				proxyData["health_check_interval_s"] = proxy.HealthCheckIntervalS
			}
			if proxy.HealthCheckURL != "" {
				proxyData["health_check_url"] = proxy.HealthCheckURL
			}
		}

		// Add meta information
		for k, v := range proxy.Metas {
			proxyData["meta_"+k] = v
		}

		// Add headers for HTTP/HTTPS proxies
		if proxy.Headers != nil {
			for k, v := range proxy.Headers {
				proxyData["header_"+k] = v
			}
		}

		// Add plugin headers
		if proxy.PluginHeaders != nil {
			for k, v := range proxy.PluginHeaders {
				proxyData["plugin_header_"+k] = v
			}
		}

		// Add bandwidth limit
		if proxy.BandwidthLimit != "" {
			proxyData["bandwidth_limit"] = proxy.BandwidthLimit
		}
		if proxy.BandwidthLimitMode != "" {
			proxyData["bandwidth_limit_mode"] = proxy.BandwidthLimitMode
		}

		// Add proxy protocol version
		if proxy.ProxyProtocolVersion != "" {
			proxyData["proxy_protocol_version"] = proxy.ProxyProtocolVersion
		}

		tomlData[proxy.Name] = proxyData
	}

	// Marshal to TOML
	b, err := toml.Marshal(tomlData)
	if err != nil {
		return err
	}

	return os.WriteFile(path, b, 0666)
}

// Complete prunes and completes this config.
// When "read" is true, the config should be completed for a file loaded from source.
// Otherwise, it should be completed for file written to disk.
func (conf *ClientConfig) Complete(read bool) {
	// Common config
	if conf.LegacyFormat {
		conf.TokenSource = ""
	}
	conf.ClientAuth = conf.ClientAuth.Complete()
	if conf.AdminPort == 0 {
		conf.AdminUser = ""
		conf.AdminPwd = ""
		conf.AssetsDir = ""
		conf.PprofEnable = false
	}
	conf.AutoDelete = conf.AutoDelete.Complete()
	if !conf.TCPMux {
		conf.TCPMuxKeepaliveInterval = 0
	}
	if !conf.TLSEnable {
		conf.TLSServerName = ""
		conf.TLSCertFile = ""
		conf.TLSKeyFile = ""
		conf.TLSTrustedCaFile = ""
	}
	if conf.Protocol == consts.ProtoQUIC {
		conf.DialServerTimeout = 0
		conf.DialServerKeepAlive = 0
	} else {
		conf.QUICMaxIdleTimeout = 0
		conf.QUICKeepalivePeriod = 0
		conf.QUICMaxIncomingStreams = 0
	}
	// Proxies
	for _, proxy := range conf.Proxies {
		// Complete proxy
		proxy.Complete()
		// Check proxy status
		if read && len(conf.Start) > 0 {
			proxy.Disabled = !lo.Every(conf.Start, proxy.GetAlias())
		}
	}
	if !read {
		conf.Start = conf.gatherStart()
	}
}

// Copy creates a new copy of this config.
func (conf *ClientConfig) Copy(all bool) *ClientConfig {
	newConf := NewDefaultClientConfig()
	newConf.ClientCommon = conf.ClientCommon
	// We can't share the same log file between different configs
	newConf.LogFile = ""
	if all {
		for _, proxy := range conf.Proxies {
			var newProxy = *proxy
			newConf.Proxies = append(newConf.Proxies, &newProxy)
		}
	}
	return newConf
}

// gatherStart returns a list of enabled proxies name, or a nil slice if all proxies are enabled.
func (conf *ClientConfig) gatherStart() []string {
	allStart := true
	start := make([]string, 0)
	for _, proxy := range conf.Proxies {
		if !proxy.Disabled {
			start = append(start, proxy.GetAlias()...)
		} else {
			allStart = false
		}
	}
	if allStart {
		return nil
	}
	return start
}

// CountStart returns the number of enabled proxies.
func (conf *ClientConfig) CountStart() int {
	return len(lo.Filter(conf.Proxies, func(proxy *Proxy, i int) bool { return !proxy.Disabled }))
}

// Ext is the file extension of this config.
func (conf *ClientConfig) Ext() string {
	if conf.LegacyFormat {
		return ".ini"
	} else {
		return ".toml"
	}
}

// NewProxyFromIni creates a proxy object from ini section
func NewProxyFromIni(name string, section *ini.Section) (*Proxy, error) {
	proxy := NewDefaultProxyConfig(name)
	if err := section.MapTo(&proxy); err != nil {
		return nil, err
	}
	proxy.Metas = util.GetMapWithoutPrefix(section.KeysHash(), "meta_")
	proxy.Headers = util.GetMapWithoutPrefix(section.KeysHash(), "header_")
	proxy.PluginHeaders = util.GetMapWithoutPrefix(section.KeysHash(), "plugin_header_")
	proxy.Name = strings.TrimPrefix(proxy.Name, consts.RangePrefix)
	return proxy, nil
}

// UnmarshalProxyFromIni finds a single proxy section and unmarshals it from ini source.
func UnmarshalProxyFromIni(source interface{}) (*Proxy, error) {
	cfg, err := ini.LoadSources(ini.LoadOptions{
		IgnoreInlineComment: true,
		AllowBooleanKeys:    true,
	}, source)
	if err != nil {
		return nil, err
	}
	var useName string
	var useSection *ini.Section
	// Try to find a proxy section
findSection:
	for _, section := range cfg.Sections() {
		switch section.Name() {
		case "common":
			continue
		case ini.DefaultSection:
			// Use the default section if no proxy is found
			useName, useSection = "", section
			continue
		default:
			useName, useSection = section.Name(), section
			break findSection
		}
	}
	if useSection == nil || len(useSection.Keys()) == 0 {
		return nil, ini.ErrDelimiterNotFound{}
	}
	return NewProxyFromIni(useName, useSection)
}

func UnmarshalClientConfFromIni(source interface{}) (*ClientConfig, error) {
	conf := NewDefaultClientConfig()
	cfg, err := ini.LoadSources(ini.LoadOptions{
		IgnoreInlineComment: true,
		AllowBooleanKeys:    true,
	}, source)
	if err != nil {
		return nil, err
	}
	// Load common options
	common, err := cfg.GetSection("common")
	if err != nil {
		return nil, err
	}
	if err = common.MapTo(&conf.ClientCommon); err != nil {
		return nil, err
	}
	conf.Metas = util.GetMapWithoutPrefix(common.KeysHash(), "meta_")
	conf.OIDCAdditionalEndpointParams = util.GetMapWithoutPrefix(common.KeysHash(), "oidc_additional_")
	// Load all proxies
	for _, section := range cfg.Sections() {
		name := section.Name()
		if name == ini.DefaultSection || name == "common" {
			continue
		}
		proxy, err := NewProxyFromIni(name, section)
		if err != nil {
			return nil, err
		}
		conf.Proxies = append(conf.Proxies, proxy)
	}
	conf.Complete(true)
	conf.LegacyFormat = true
	return conf, nil
}

func UnmarshalClientConf(source interface{}) (*ClientConfig, error) {
	var b []byte
	var err error
	if path, ok := source.(string); ok {
		b, err = os.ReadFile(path)
		if err != nil {
			return nil, err
		}
	} else {
		b = source.([]byte)
	}
	if DetectLegacyINIFormat(b) {
		return UnmarshalClientConfFromIni(source)
	}

	// Parse TOML directly without using frp's internal structures
	var tomlData map[string]interface{}
	if err := toml.Unmarshal(b, &tomlData); err != nil {
		return nil, err
	}

	conf := NewDefaultClientConfig()

	// Parse common configuration
	if commonData, ok := tomlData["common"].(map[string]interface{}); ok {
		if addr, ok := commonData["server_addr"].(string); ok {
			conf.ServerAddress = addr
		}
		if port, ok := commonData["server_port"].(int64); ok {
			conf.ServerPort = int(port)
		}
		if token, ok := commonData["token"].(string); ok {
			conf.Token = token
		}
		if user, ok := commonData["user"].(string); ok {
			conf.User = user
		}
		if logFile, ok := commonData["log_file"].(string); ok {
			conf.LogFile = logFile
		}
		if logLevel, ok := commonData["log_level"].(string); ok {
			conf.LogLevel = logLevel
		}
		if logMaxDays, ok := commonData["log_max_days"].(int64); ok {
			conf.LogMaxDays = logMaxDays
		}
		if adminAddr, ok := commonData["admin_addr"].(string); ok {
			conf.AdminAddr = adminAddr
		}
		if adminPort, ok := commonData["admin_port"].(int64); ok {
			conf.AdminPort = int(adminPort)
		}
		if adminUser, ok := commonData["admin_user"].(string); ok {
			conf.AdminUser = adminUser
		}
		if adminPwd, ok := commonData["admin_pwd"].(string); ok {
			conf.AdminPwd = adminPwd
		}
		if tlsEnable, ok := commonData["tls_enable"].(bool); ok {
			conf.TLSEnable = tlsEnable
		}
		if tlsCertFile, ok := commonData["tls_cert_file"].(string); ok {
			conf.TLSCertFile = tlsCertFile
		}
		if tlsKeyFile, ok := commonData["tls_key_file"].(string); ok {
			conf.TLSKeyFile = tlsKeyFile
		}
		if tlsTrustedCaFile, ok := commonData["tls_trusted_ca_file"].(string); ok {
			conf.TLSTrustedCaFile = tlsTrustedCaFile
		}
		if tlsServerName, ok := commonData["tls_server_name"].(string); ok {
			conf.TLSServerName = tlsServerName
		}
		if protocol, ok := commonData["protocol"].(string); ok {
			conf.Protocol = protocol
		}
		if tcpMux, ok := commonData["tcp_mux"].(bool); ok {
			conf.TCPMux = tcpMux
		}
		if heartbeatInterval, ok := commonData["heartbeat_interval"].(int64); ok {
			conf.HeartbeatInterval = heartbeatInterval
		}
		if heartbeatTimeout, ok := commonData["heartbeat_timeout"].(int64); ok {
			conf.HeartbeatTimeout = heartbeatTimeout
		}
		if udpPacketSize, ok := commonData["udp_packet_size"].(int64); ok {
			conf.UDPPacketSize = udpPacketSize
		}

		// Parse manager-specific fields
		if name, ok := commonData["frpcgui_name"].(string); ok {
			conf.ClientCommon.Name = name
		}
		if manualStart, ok := commonData["frpcgui_manual_start"].(bool); ok {
			conf.ManualStart = manualStart
		}

		// Parse auto-delete configuration
		if deleteMethod, ok := commonData["frpcgui_delete_method"].(string); ok {
			conf.AutoDelete.DeleteMethod = deleteMethod
			switch deleteMethod {
			case consts.DeleteAbsolute:
				if deleteAfterDate, ok := commonData["frpcgui_delete_after_date"].(string); ok {
					if date, err := time.Parse("2006-01-02T15:04:05Z", deleteAfterDate); err == nil {
						conf.AutoDelete.DeleteAfterDate = date
					}
				}
			case consts.DeleteRelative:
				if deleteAfterDays, ok := commonData["frpcgui_delete_after_days"].(int64); ok {
					conf.AutoDelete.DeleteAfterDays = deleteAfterDays
				}
			}
		}

		// Parse meta information
		conf.Metas = make(map[string]string)
		for k, v := range commonData {
			if strings.HasPrefix(k, "meta_") {
				if vStr, ok := v.(string); ok {
					metaKey := strings.TrimPrefix(k, "meta_")
					conf.Metas[metaKey] = vStr
				}
			}
		}
	}

	// Parse proxies
	for name, data := range tomlData {
		if name != "common" {
			proxyData, ok := data.(map[string]interface{})
			if !ok {
				continue
			}
			proxy := NewDefaultProxyConfig(name)

			if proxyType, ok := proxyData["type"].(string); ok {
				proxy.Type = proxyType
			}
			if localIP, ok := proxyData["local_ip"].(string); ok {
				proxy.LocalIP = localIP
			}
			if localPort, ok := proxyData["local_port"].(string); ok {
				proxy.LocalPort = localPort
			}
			if remotePort, ok := proxyData["remote_port"].(string); ok {
				proxy.RemotePort = remotePort
			}
			if useEncryption, ok := proxyData["use_encryption"].(bool); ok {
				proxy.UseEncryption = useEncryption
			}
			if useCompression, ok := proxyData["use_compression"].(bool); ok {
				proxy.UseCompression = useCompression
			}
			if group, ok := proxyData["group"].(string); ok {
				proxy.Group = group
			}
			if groupKey, ok := proxyData["group_key"].(string); ok {
				proxy.GroupKey = groupKey
			}

			// Parse proxy-specific fields based on type
			switch proxy.Type {
			case consts.ProxyTypeHTTP, consts.ProxyTypeHTTPS:
				if customDomains, ok := proxyData["custom_domains"].(string); ok {
					proxy.CustomDomains = customDomains
				}
				if subDomain, ok := proxyData["subdomain"].(string); ok {
					proxy.SubDomain = subDomain
				}
				if locations, ok := proxyData["locations"].(string); ok {
					proxy.Locations = locations
				}
				if httpUser, ok := proxyData["http_user"].(string); ok {
					proxy.HTTPUser = httpUser
				}
				if httpPwd, ok := proxyData["http_pwd"].(string); ok {
					proxy.HTTPPwd = httpPwd
				}
				if hostHeaderRewrite, ok := proxyData["host_header_rewrite"].(string); ok {
					proxy.HostHeaderRewrite = hostHeaderRewrite
				}
				if multiplexer, ok := proxyData["multiplexer"].(string); ok {
					proxy.Multiplexer = multiplexer
				}
			case consts.ProxyTypeSTCP, consts.ProxyTypeXTCP, consts.ProxyTypeSUDP:
				if role, ok := proxyData["role"].(string); ok {
					proxy.Role = role
				}
				if sk, ok := proxyData["sk"].(string); ok {
					proxy.SK = sk
				}
				if allowUsers, ok := proxyData["allow_users"].(string); ok {
					proxy.AllowUsers = allowUsers
				}
				if serverUser, ok := proxyData["server_user"].(string); ok {
					proxy.ServerUser = serverUser
				}
				if serverName, ok := proxyData["server_name"].(string); ok {
					proxy.ServerName = serverName
				}
				if bindAddr, ok := proxyData["bind_addr"].(string); ok {
					proxy.BindAddr = bindAddr
				}
				if bindPort, ok := proxyData["bind_port"].(int64); ok {
					proxy.BindPort = int(bindPort)
				}
				if keepTunnelOpen, ok := proxyData["keep_tunnel_open"].(bool); ok {
					proxy.KeepTunnelOpen = keepTunnelOpen
				}
				if maxRetriesAnHour, ok := proxyData["max_retries_an_hour"].(int64); ok {
					proxy.MaxRetriesAnHour = int(maxRetriesAnHour)
				}
				if minRetryInterval, ok := proxyData["min_retry_interval"].(int64); ok {
					proxy.MinRetryInterval = int(minRetryInterval)
				}
				if fallbackTo, ok := proxyData["fallback_to"].(string); ok {
					proxy.FallbackTo = fallbackTo
				}
				if fallbackTimeoutMs, ok := proxyData["fallback_timeout_ms"].(int64); ok {
					proxy.FallbackTimeoutMs = int(fallbackTimeoutMs)
				}
			}

			// Parse plugin configuration
			if plugin, ok := proxyData["plugin"].(string); ok {
				proxy.Plugin = plugin
				if pluginLocalAddr, ok := proxyData["plugin_local_addr"].(string); ok {
					proxy.PluginLocalAddr = pluginLocalAddr
				}
				if pluginCrtPath, ok := proxyData["plugin_crt_path"].(string); ok {
					proxy.PluginCrtPath = pluginCrtPath
				}
				if pluginKeyPath, ok := proxyData["plugin_key_path"].(string); ok {
					proxy.PluginKeyPath = pluginKeyPath
				}
				if pluginHostHeaderRewrite, ok := proxyData["plugin_host_header_rewrite"].(string); ok {
					proxy.PluginHostHeaderRewrite = pluginHostHeaderRewrite
				}
				if pluginHttpUser, ok := proxyData["plugin_http_user"].(string); ok {
					proxy.PluginHttpUser = pluginHttpUser
				}
				if pluginHttpPasswd, ok := proxyData["plugin_http_passwd"].(string); ok {
					proxy.PluginHttpPasswd = pluginHttpPasswd
				}
				if pluginUser, ok := proxyData["plugin_user"].(string); ok {
					proxy.PluginUser = pluginUser
				}
				if pluginPasswd, ok := proxyData["plugin_passwd"].(string); ok {
					proxy.PluginPasswd = pluginPasswd
				}
				if pluginLocalPath, ok := proxyData["plugin_local_path"].(string); ok {
					proxy.PluginLocalPath = pluginLocalPath
				}
				if pluginStripPrefix, ok := proxyData["plugin_strip_prefix"].(string); ok {
					proxy.PluginStripPrefix = pluginStripPrefix
				}
				if pluginUnixPath, ok := proxyData["plugin_unix_path"].(string); ok {
					proxy.PluginUnixPath = pluginUnixPath
				}
			}

			// Parse health check configuration
			if healthCheckType, ok := proxyData["health_check_type"].(string); ok {
				proxy.HealthCheckType = healthCheckType
				if healthCheckTimeoutS, ok := proxyData["health_check_timeout_s"].(int64); ok {
					proxy.HealthCheckTimeoutS = int(healthCheckTimeoutS)
				}
				if healthCheckMaxFailed, ok := proxyData["health_check_max_failed"].(int64); ok {
					proxy.HealthCheckMaxFailed = int(healthCheckMaxFailed)
				}
				if healthCheckIntervalS, ok := proxyData["health_check_interval_s"].(int64); ok {
					proxy.HealthCheckIntervalS = int(healthCheckIntervalS)
				}
				if healthCheckURL, ok := proxyData["health_check_url"].(string); ok {
					proxy.HealthCheckURL = healthCheckURL
				}
			}

			// Parse meta information
			proxy.Metas = make(map[string]string)
			for k, v := range proxyData {
				if strings.HasPrefix(k, "meta_") {
					if vStr, ok := v.(string); ok {
						metaKey := strings.TrimPrefix(k, "meta_")
						proxy.Metas[metaKey] = vStr
					}
				}
			}

			// Parse headers for HTTP/HTTPS proxies
			proxy.Headers = make(map[string]string)
			for k, v := range proxyData {
				if strings.HasPrefix(k, "header_") {
					if vStr, ok := v.(string); ok {
						headerKey := strings.TrimPrefix(k, "header_")
						proxy.Headers[headerKey] = vStr
					}
				}
			}

			// Parse plugin headers
			proxy.PluginHeaders = make(map[string]string)
			for k, v := range proxyData {
				if strings.HasPrefix(k, "plugin_header_") {
					if vStr, ok := v.(string); ok {
						headerKey := strings.TrimPrefix(k, "plugin_header_")
						proxy.PluginHeaders[headerKey] = vStr
					}
				}
			}

			// Parse bandwidth limit
			if bandwidthLimit, ok := proxyData["bandwidth_limit"].(string); ok {
				proxy.BandwidthLimit = bandwidthLimit
			}
			if bandwidthLimitMode, ok := proxyData["bandwidth_limit_mode"].(string); ok {
				proxy.BandwidthLimitMode = bandwidthLimitMode
			}

			// Parse proxy protocol version
			if proxyProtocolVersion, ok := proxyData["proxy_protocol_version"].(string); ok {
				proxy.ProxyProtocolVersion = proxyProtocolVersion
			}

			conf.Proxies = append(conf.Proxies, proxy)
		}
	}

	conf.Complete(true)
	return conf, nil
}

func NewDefaultClientConfig() *ClientConfig {
	return &ClientConfig{
		ClientCommon: ClientCommon{
			ClientAuth:                ClientAuth{AuthMethod: consts.AuthToken},
			ServerPort:                consts.DefaultServerPort,
			LogLevel:                  consts.LogLevelInfo,
			LogMaxDays:                consts.DefaultLogMaxDays,
			TCPMux:                    true,
			TLSEnable:                 true,
			DisableCustomTLSFirstByte: true,
			AutoDelete:                AutoDelete{DeleteMethod: consts.DeleteRelative},
		},
		Proxies: make([]*Proxy, 0),
	}
}

// DetectLegacyINIFormat checks if the configuration is in INI format
func DetectLegacyINIFormat(data []byte) bool {
	// Simple heuristic: if the file contains [common] section, it's likely INI format
	return strings.Contains(string(data), "[common]")
}

func NewDefaultProxyConfig(name string) *Proxy {
	return &Proxy{
		BaseProxyConf: BaseProxyConf{
			Name: name, Type: consts.ProxyTypeTCP,
		},
	}
}
