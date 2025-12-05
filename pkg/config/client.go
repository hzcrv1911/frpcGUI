package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
	"github.com/samber/lo"

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
	return conf.saveTOML(path, true)
}

func (conf *ClientConfig) SavePure(path string) error {
	return conf.saveTOML(path, false)
}

func (conf *ClientConfig) saveTOML(path string, includePrivate bool) error {
	tomlData := make(map[string]interface{})

	// Server
	tomlData["serverAddr"] = conf.ServerAddress
	tomlData["serverPort"] = conf.ServerPort

	// Auth
	if conf.Token != "" || conf.OIDCClientId != "" {
		auth := make(map[string]interface{})
		if conf.Token != "" {
			auth["method"] = "token"
			auth["token"] = conf.Token
		}
		if conf.OIDCClientId != "" {
			auth["method"] = "oidc"
			auth["oidcClientId"] = conf.OIDCClientId
			auth["oidcClientSecret"] = conf.OIDCClientSecret
			auth["oidcAudience"] = conf.OIDCAudience
			auth["oidcScope"] = conf.OIDCScope
			auth["oidcTokenEndpoint"] = conf.OIDCTokenEndpoint
			if len(conf.OIDCAdditionalEndpointParams) > 0 {
				auth["oidcAdditionalEndpointParams"] = conf.OIDCAdditionalEndpointParams
			}
		}
		tomlData["auth"] = auth
	}

	// User
	if conf.User != "" {
		tomlData["user"] = conf.User
	}

	// Log
	if conf.LogFile != "" || conf.LogLevel != "" || conf.LogMaxDays > 0 {
		log := make(map[string]interface{})
		if conf.LogFile != "" {
			// Try to make the log path relative to the config file
			baseDir := filepath.Dir(path)
			if rel, err := filepath.Rel(baseDir, conf.LogFile); err == nil {
				log["to"] = filepath.ToSlash(rel)
			} else {
				log["to"] = conf.LogFile
			}
		}
		if conf.LogLevel != "" {
			log["level"] = conf.LogLevel
		}
		if conf.LogMaxDays > 0 {
			log["maxDays"] = conf.LogMaxDays
		}
		tomlData["log"] = log
	}

	// WebServer (Admin)
	if conf.AdminPort > 0 {
		webServer := make(map[string]interface{})
		if conf.AdminAddr != "" {
			webServer["addr"] = conf.AdminAddr
		}
		webServer["port"] = conf.AdminPort
		if conf.AdminUser != "" {
			webServer["user"] = conf.AdminUser
		}
		if conf.AdminPwd != "" {
			webServer["password"] = conf.AdminPwd
		}
		if conf.AssetsDir != "" {
			webServer["assetsDir"] = conf.AssetsDir
		}
		if conf.PprofEnable {
			webServer["pprofEnable"] = true
		}
		tomlData["webServer"] = webServer
	}

	// Transport
	transport := make(map[string]interface{})
	if conf.Protocol != "" {
		transport["protocol"] = conf.Protocol
	}
	if conf.TCPMux {
		transport["tcpMux"] = true
	}
	if conf.TCPMuxKeepaliveInterval > 0 {
		transport["tcpMuxKeepaliveInterval"] = conf.TCPMuxKeepaliveInterval
	}
	if conf.HeartbeatInterval > 0 {
		transport["heartbeatInterval"] = conf.HeartbeatInterval
	}
	if conf.HeartbeatTimeout > 0 {
		transport["heartbeatTimeout"] = conf.HeartbeatTimeout
	}
	if conf.UDPPacketSize > 0 {
		transport["udpPacketSize"] = conf.UDPPacketSize
	}
	if conf.TLSEnable {
		tls := make(map[string]interface{})
		tls["enable"] = true
		if conf.TLSCertFile != "" {
			tls["certFile"] = conf.TLSCertFile
		}
		if conf.TLSKeyFile != "" {
			tls["keyFile"] = conf.TLSKeyFile
		}
		if conf.TLSTrustedCaFile != "" {
			tls["trustedCaFile"] = conf.TLSTrustedCaFile
		}
		if conf.TLSServerName != "" {
			tls["serverName"] = conf.TLSServerName
		}
		if conf.DisableCustomTLSFirstByte {
			tls["disableCustomTLSFirstByte"] = true
		}
		transport["tls"] = tls
	}
	if conf.Protocol == consts.ProtoQUIC {
		quic := make(map[string]interface{})
		if conf.QUICKeepalivePeriod > 0 {
			quic["keepalivePeriod"] = conf.QUICKeepalivePeriod
		}
		if conf.QUICMaxIdleTimeout > 0 {
			quic["maxIdleTimeout"] = conf.QUICMaxIdleTimeout
		}
		if conf.QUICMaxIncomingStreams > 0 {
			quic["maxIncomingStreams"] = conf.QUICMaxIncomingStreams
		}
		transport["quic"] = quic
	}
	if len(transport) > 0 {
		tomlData["transport"] = transport
	}

	// Custom fields
	if includePrivate {
		tomlData["frpcgui_name"] = conf.Name()
		if conf.ManualStart {
			tomlData["frpcgui_manual_start"] = true
		}
		if conf.AutoDelete.DeleteMethod != "" {
			tomlData["frpcgui_delete_method"] = conf.AutoDelete.DeleteMethod
			if conf.AutoDelete.DeleteMethod == consts.DeleteAbsolute && !conf.AutoDelete.DeleteAfterDate.IsZero() {
				tomlData["frpcgui_delete_after_date"] = conf.AutoDelete.DeleteAfterDate.Format("2006-01-02T15:04:05Z")
			} else if conf.AutoDelete.DeleteMethod == consts.DeleteRelative && conf.AutoDelete.DeleteAfterDays > 0 {
				tomlData["frpcgui_delete_after_days"] = conf.AutoDelete.DeleteAfterDays
			}
		}
	}

	if len(conf.Metas) > 0 {
		tomlData["metadatas"] = conf.Metas
	}

	// Proxies
	var proxies []map[string]interface{}
	for _, p := range conf.Proxies {
		proxy := make(map[string]interface{})
		proxy["name"] = p.Name
		proxy["type"] = p.Type
		if p.LocalIP != "" {
			proxy["localIP"] = p.LocalIP
		}
		if p.LocalPort != "" {
			if port, err := strconv.Atoi(p.LocalPort); err == nil {
				proxy["localPort"] = port
			} else {
				proxy["localPort"] = p.LocalPort
			}
		}
		if p.RemotePort != "" {
			if port, err := strconv.Atoi(p.RemotePort); err == nil {
				proxy["remotePort"] = port
			} else {
				proxy["remotePort"] = p.RemotePort
			}
		}

		// Transport
		transport := make(map[string]interface{})
		if p.UseEncryption {
			transport["useEncryption"] = true
		}
		if p.UseCompression {
			transport["useCompression"] = true
		}
		if p.BandwidthLimit != "" {
			transport["bandwidthLimit"] = p.BandwidthLimit
			if p.BandwidthLimitMode != "" {
				transport["bandwidthLimitMode"] = p.BandwidthLimitMode
			}
		}
		if p.ProxyProtocolVersion != "" {
			transport["proxyProtocolVersion"] = p.ProxyProtocolVersion
		}
		if len(transport) > 0 {
			proxy["transport"] = transport
		}

		// Load Balancer
		if p.Group != "" {
			lb := make(map[string]interface{})
			lb["group"] = p.Group
			if p.GroupKey != "" {
				lb["groupKey"] = p.GroupKey
			}
			proxy["loadBalancer"] = lb
		}

		// Health Check
		if p.HealthCheckType != "" {
			hc := make(map[string]interface{})
			hc["type"] = p.HealthCheckType
			if p.HealthCheckTimeoutS > 0 {
				hc["timeoutSeconds"] = p.HealthCheckTimeoutS
			}
			if p.HealthCheckMaxFailed > 0 {
				hc["maxFailed"] = p.HealthCheckMaxFailed
			}
			if p.HealthCheckIntervalS > 0 {
				hc["intervalSeconds"] = p.HealthCheckIntervalS
			}
			if p.HealthCheckURL != "" {
				hc["path"] = p.HealthCheckURL
			}
			proxy["healthCheck"] = hc
		}

		// Plugin
		if p.Plugin != "" {
			plugin := make(map[string]interface{})
			plugin["type"] = p.Plugin
			if p.PluginLocalAddr != "" {
				plugin["localAddr"] = p.PluginLocalAddr
			}
			if p.PluginCrtPath != "" {
				plugin["crtPath"] = p.PluginCrtPath
			}
			if p.PluginKeyPath != "" {
				plugin["keyPath"] = p.PluginKeyPath
			}
			if p.PluginHostHeaderRewrite != "" {
				plugin["hostHeaderRewrite"] = p.PluginHostHeaderRewrite
			}
			if p.PluginHttpUser != "" {
				plugin["httpUser"] = p.PluginHttpUser
			}
			if p.PluginHttpPasswd != "" {
				plugin["httpPassword"] = p.PluginHttpPasswd
			}
			if p.PluginUser != "" {
				plugin["user"] = p.PluginUser
			}
			if p.PluginPasswd != "" {
				plugin["password"] = p.PluginPasswd
			}
			if p.PluginLocalPath != "" {
				plugin["localPath"] = p.PluginLocalPath
			}
			if p.PluginStripPrefix != "" {
				plugin["stripPrefix"] = p.PluginStripPrefix
			}
			if p.PluginUnixPath != "" {
				plugin["unixPath"] = p.PluginUnixPath
			}
			proxy["plugin"] = plugin
		}

		// Type specific
		switch p.Type {
		case consts.ProxyTypeHTTP, consts.ProxyTypeHTTPS:
			if p.CustomDomains != "" {
				proxy["customDomains"] = strings.Split(p.CustomDomains, ",")
			}
			if p.SubDomain != "" {
				proxy["subDomain"] = p.SubDomain
			}
			if p.Locations != "" {
				proxy["locations"] = strings.Split(p.Locations, ",")
			}
			if p.HTTPUser != "" {
				proxy["httpUser"] = p.HTTPUser
			}
			if p.HTTPPwd != "" {
				proxy["httpPassword"] = p.HTTPPwd
			}
			if p.HostHeaderRewrite != "" {
				proxy["hostHeaderRewrite"] = p.HostHeaderRewrite
			}
			if p.Multiplexer != "" {
				proxy["multiplexer"] = p.Multiplexer
			}
		case consts.ProxyTypeSTCP, consts.ProxyTypeXTCP, consts.ProxyTypeSUDP:
			if p.Role != "" {
				proxy["role"] = p.Role
			}
			if p.SK != "" {
				proxy["secretKey"] = p.SK
			}
			if p.AllowUsers != "" {
				proxy["allowUsers"] = strings.Split(p.AllowUsers, ",")
			}
			if p.IsVisitor() {
				if p.ServerUser != "" {
					proxy["serverUser"] = p.ServerUser
				}
				if p.ServerName != "" {
					proxy["serverName"] = p.ServerName
				}
				if p.BindAddr != "" {
					proxy["bindAddr"] = p.BindAddr
				}
				if p.BindPort > 0 {
					proxy["bindPort"] = p.BindPort
				}
				if p.KeepTunnelOpen {
					proxy["keepTunnelOpen"] = true
				}
				if p.MaxRetriesAnHour > 0 {
					proxy["maxRetriesAnHour"] = p.MaxRetriesAnHour
				}
				if p.MinRetryInterval > 0 {
					proxy["minRetryInterval"] = p.MinRetryInterval
				}
				if p.FallbackTo != "" {
					proxy["fallbackTo"] = p.FallbackTo
				}
				if p.FallbackTimeoutMs > 0 {
					proxy["fallbackTimeoutMs"] = p.FallbackTimeoutMs
				}
			}
		}

		if len(p.Metas) > 0 {
			proxy["metadatas"] = p.Metas
		}
		if len(p.Headers) > 0 {
			proxy["headers"] = p.Headers
		}

		proxies = append(proxies, proxy)
	}
	tomlData["proxies"] = proxies

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
	return ".toml"
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

	var tomlData map[string]interface{}
	if err := toml.Unmarshal(b, &tomlData); err != nil {
		return nil, err
	}

	conf := NewDefaultClientConfig()

	// Helper to get string
	getString := func(m map[string]interface{}, key string) string {
		if v, ok := m[key].(string); ok {
			return v
		}
		return ""
	}
	// Helper to get int
	getInt := func(m map[string]interface{}, key string) int {
		if v, ok := m[key].(int64); ok {
			return int(v)
		}
		return 0
	}
	// Helper to get bool
	getBool := func(m map[string]interface{}, key string) bool {
		if v, ok := m[key].(bool); ok {
			return v
		}
		return false
	}

	// Top level fields
	conf.ServerAddress = getString(tomlData, "serverAddr")
	conf.ServerPort = getInt(tomlData, "serverPort")
	conf.User = getString(tomlData, "user")

	// Auth
	if auth, ok := tomlData["auth"].(map[string]interface{}); ok {
		method := getString(auth, "method")
		if method == "token" {
			conf.Token = getString(auth, "token")
		} else if method == "oidc" {
			conf.OIDCClientId = getString(auth, "oidcClientId")
			conf.OIDCClientSecret = getString(auth, "oidcClientSecret")
			conf.OIDCAudience = getString(auth, "oidcAudience")
			conf.OIDCScope = getString(auth, "oidcScope")
			conf.OIDCTokenEndpoint = getString(auth, "oidcTokenEndpoint")
			if params, ok := auth["oidcAdditionalEndpointParams"].(map[string]interface{}); ok {
				conf.OIDCAdditionalEndpointParams = make(map[string]string)
				for k, v := range params {
					if s, ok := v.(string); ok {
						conf.OIDCAdditionalEndpointParams[k] = s
					}
				}
			}
		}
	}

	// Log
	if log, ok := tomlData["log"].(map[string]interface{}); ok {
		conf.LogFile = getString(log, "to")
		conf.LogLevel = getString(log, "level")
		conf.LogMaxDays = int64(getInt(log, "maxDays"))
	}

	// WebServer
	if ws, ok := tomlData["webServer"].(map[string]interface{}); ok {
		conf.AdminAddr = getString(ws, "addr")
		conf.AdminPort = getInt(ws, "port")
		conf.AdminUser = getString(ws, "user")
		conf.AdminPwd = getString(ws, "password")
		conf.AssetsDir = getString(ws, "assetsDir")
		conf.PprofEnable = getBool(ws, "pprofEnable")
	}

	// Transport
	if transport, ok := tomlData["transport"].(map[string]interface{}); ok {
		conf.Protocol = getString(transport, "protocol")
		conf.TCPMux = getBool(transport, "tcpMux")
		conf.TCPMuxKeepaliveInterval = int64(getInt(transport, "tcpMuxKeepaliveInterval"))
		conf.HeartbeatInterval = int64(getInt(transport, "heartbeatInterval"))
		conf.HeartbeatTimeout = int64(getInt(transport, "heartbeatTimeout"))
		conf.UDPPacketSize = int64(getInt(transport, "udpPacketSize"))

		if tls, ok := transport["tls"].(map[string]interface{}); ok {
			conf.TLSEnable = getBool(tls, "enable")
			conf.TLSCertFile = getString(tls, "certFile")
			conf.TLSKeyFile = getString(tls, "keyFile")
			conf.TLSTrustedCaFile = getString(tls, "trustedCaFile")
			conf.TLSServerName = getString(tls, "serverName")
			conf.DisableCustomTLSFirstByte = getBool(tls, "disableCustomTLSFirstByte")
		}

		if quic, ok := transport["quic"].(map[string]interface{}); ok {
			conf.QUICKeepalivePeriod = getInt(quic, "keepalivePeriod")
			conf.QUICMaxIdleTimeout = getInt(quic, "maxIdleTimeout")
			conf.QUICMaxIncomingStreams = getInt(quic, "maxIncomingStreams")
		}
	}

	// Custom fields
	conf.ClientCommon.Name = getString(tomlData, "frpcgui_name")
	conf.ManualStart = getBool(tomlData, "frpcgui_manual_start")
	conf.AutoDelete.DeleteMethod = getString(tomlData, "frpcgui_delete_method")
	if dateStr := getString(tomlData, "frpcgui_delete_after_date"); dateStr != "" {
		if date, err := time.Parse("2006-01-02T15:04:05Z", dateStr); err == nil {
			conf.AutoDelete.DeleteAfterDate = date
		}
	}
	conf.AutoDelete.DeleteAfterDays = int64(getInt(tomlData, "frpcgui_delete_after_days"))

	// Metas
	if metas, ok := tomlData["metadatas"].(map[string]interface{}); ok {
		conf.Metas = make(map[string]string)
		for k, v := range metas {
			if s, ok := v.(string); ok {
				conf.Metas[k] = s
			}
		}
	}

	// Proxies
	if proxies, ok := tomlData["proxies"].([]interface{}); ok {
		for _, pData := range proxies {
			if pMap, ok := pData.(map[string]interface{}); ok {
				proxy := NewDefaultProxyConfig(getString(pMap, "name"))
				proxy.Type = getString(pMap, "type")
				proxy.LocalIP = getString(pMap, "localIP")

				if v, ok := pMap["localPort"].(int64); ok {
					proxy.LocalPort = strconv.FormatInt(v, 10)
				} else if v, ok := pMap["localPort"].(string); ok {
					proxy.LocalPort = v
				}

				if v, ok := pMap["remotePort"].(int64); ok {
					proxy.RemotePort = strconv.FormatInt(v, 10)
				} else if v, ok := pMap["remotePort"].(string); ok {
					proxy.RemotePort = v
				}

				// Transport
				if transport, ok := pMap["transport"].(map[string]interface{}); ok {
					proxy.UseEncryption = getBool(transport, "useEncryption")
					proxy.UseCompression = getBool(transport, "useCompression")
					proxy.BandwidthLimit = getString(transport, "bandwidthLimit")
					proxy.BandwidthLimitMode = getString(transport, "bandwidthLimitMode")
					proxy.ProxyProtocolVersion = getString(transport, "proxyProtocolVersion")
				}

				// Load Balancer
				if lb, ok := pMap["loadBalancer"].(map[string]interface{}); ok {
					proxy.Group = getString(lb, "group")
					proxy.GroupKey = getString(lb, "groupKey")
				}

				// Health Check
				if hc, ok := pMap["healthCheck"].(map[string]interface{}); ok {
					proxy.HealthCheckType = getString(hc, "type")
					proxy.HealthCheckTimeoutS = getInt(hc, "timeoutSeconds")
					proxy.HealthCheckMaxFailed = getInt(hc, "maxFailed")
					proxy.HealthCheckIntervalS = getInt(hc, "intervalSeconds")
					proxy.HealthCheckURL = getString(hc, "path")
				}

				// Plugin
				if plugin, ok := pMap["plugin"].(map[string]interface{}); ok {
					proxy.Plugin = getString(plugin, "type")
					proxy.PluginLocalAddr = getString(plugin, "localAddr")
					proxy.PluginCrtPath = getString(plugin, "crtPath")
					proxy.PluginKeyPath = getString(plugin, "keyPath")
					proxy.PluginHostHeaderRewrite = getString(plugin, "hostHeaderRewrite")
					proxy.PluginHttpUser = getString(plugin, "httpUser")
					proxy.PluginHttpPasswd = getString(plugin, "httpPassword")
					proxy.PluginUser = getString(plugin, "user")
					proxy.PluginPasswd = getString(plugin, "password")
					proxy.PluginLocalPath = getString(plugin, "localPath")
					proxy.PluginStripPrefix = getString(plugin, "stripPrefix")
					proxy.PluginUnixPath = getString(plugin, "unixPath")
				}

				// Type specific
				if domains, ok := pMap["customDomains"].([]interface{}); ok {
					var ds []string
					for _, d := range domains {
						if s, ok := d.(string); ok {
							ds = append(ds, s)
						}
					}
					proxy.CustomDomains = strings.Join(ds, ",")
				}
				proxy.SubDomain = getString(pMap, "subDomain")
				if locs, ok := pMap["locations"].([]interface{}); ok {
					var ls []string
					for _, l := range locs {
						if s, ok := l.(string); ok {
							ls = append(ls, s)
						}
					}
					proxy.Locations = strings.Join(ls, ",")
				}
				proxy.HTTPUser = getString(pMap, "httpUser")
				proxy.HTTPPwd = getString(pMap, "httpPassword")
				proxy.HostHeaderRewrite = getString(pMap, "hostHeaderRewrite")
				proxy.Multiplexer = getString(pMap, "multiplexer")

				proxy.Role = getString(pMap, "role")
				proxy.SK = getString(pMap, "secretKey")
				if users, ok := pMap["allowUsers"].([]interface{}); ok {
					var us []string
					for _, u := range users {
						if s, ok := u.(string); ok {
							us = append(us, s)
						}
					}
					proxy.AllowUsers = strings.Join(us, ",")
				}
				proxy.ServerUser = getString(pMap, "serverUser")
				proxy.ServerName = getString(pMap, "serverName")
				proxy.BindAddr = getString(pMap, "bindAddr")
				proxy.BindPort = getInt(pMap, "bindPort")
				proxy.KeepTunnelOpen = getBool(pMap, "keepTunnelOpen")
				proxy.MaxRetriesAnHour = getInt(pMap, "maxRetriesAnHour")
				proxy.MinRetryInterval = getInt(pMap, "minRetryInterval")
				proxy.FallbackTo = getString(pMap, "fallbackTo")
				proxy.FallbackTimeoutMs = getInt(pMap, "fallbackTimeoutMs")

				// Metas
				if metas, ok := pMap["metadatas"].(map[string]interface{}); ok {
					proxy.Metas = make(map[string]string)
					for k, v := range metas {
						if s, ok := v.(string); ok {
							proxy.Metas[k] = s
						}
					}
				}
				// Headers
				if headers, ok := pMap["headers"].(map[string]interface{}); ok {
					proxy.Headers = make(map[string]string)
					for k, v := range headers {
						if s, ok := v.(string); ok {
							proxy.Headers[k] = s
						}
					}
				}

				conf.Proxies = append(conf.Proxies, proxy)
			}
		}
	} else {
		// Fallback to v1 style parsing
		if commonData, ok := tomlData["common"].(map[string]interface{}); ok {
			if conf.ServerAddress == "" {
				conf.ServerAddress = getString(commonData, "server_addr")
			}
			if conf.ServerPort == 0 {
				conf.ServerPort = getInt(commonData, "server_port")
			}
			if conf.Token == "" {
				conf.Token = getString(commonData, "token")
			}
			if conf.ClientCommon.Name == "" {
				conf.ClientCommon.Name = getString(commonData, "frpcgui_name")
			}

			for name, data := range tomlData {
				if name != "common" {
					if pMap, ok := data.(map[string]interface{}); ok {
						proxy := NewDefaultProxyConfig(name)
						proxy.Type = getString(pMap, "type")
						proxy.LocalIP = getString(pMap, "local_ip")
						proxy.LocalPort = getString(pMap, "local_port")
						proxy.RemotePort = getString(pMap, "remote_port")
						conf.Proxies = append(conf.Proxies, proxy)
					}
				}
			}
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

func NewDefaultProxyConfig(name string) *Proxy {
	return &Proxy{
		BaseProxyConf: BaseProxyConf{
			Name: name, Type: consts.ProxyTypeTCP,
		},
	}
}
