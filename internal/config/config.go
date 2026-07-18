package config

import "net"

type Opt func(*ServerConfig) error

type ServerConfig struct {
	ProxyHost     string
	ProxyPort     string
	DashboardHost string
	DashboardPort string
}

const (
	DefaultHost          = "0.0.0.0"
	DefaultProxyPort     = "8999"
	DefaultDashboardPort = "8901"
	DefaultProxyAddr     = DefaultHost + ":" + DefaultProxyPort
	DefaultDashboardAddr = DefaultHost + ":" + DefaultDashboardPort
)

func WithLocalhostDefaults() Opt {
	return func(sc *ServerConfig) error {
		sc.ProxyHost, sc.ProxyPort, _ = net.SplitHostPort(DefaultProxyAddr)
		sc.DashboardHost, sc.DashboardPort, _ = net.SplitHostPort(DefaultDashboardAddr)
		return nil
	}
}

func WithProxyAddress(hostport string) Opt {
	return func(sc *ServerConfig) error {
		var host, port string
		var err error
		if host, port, err = net.SplitHostPort(hostport); err != nil {
			return err
		}
		sc.ProxyHost = host
		sc.ProxyPort = port
		return nil
	}
}

func WithDashboardAddress(hostport string) Opt {
	return func(sc *ServerConfig) error {
		var host, port string
		var err error
		if host, port, err = net.SplitHostPort(hostport); err != nil {
			return err
		}
		sc.DashboardHost = host
		sc.DashboardPort = port
		return nil
	}
}
