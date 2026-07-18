package config

import "net"

type Opt func(*ServerConfig) error

type ServerConfig struct {
	ProxyHost     string
	ProxyPort     string
	DashboardHost string
	DashboardPort string
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
