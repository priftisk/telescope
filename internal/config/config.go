package config

import "net"

type Opt func(*ProxyConfig) error

type ProxyConfig struct {
	Host string
	Port string
}

func WithProxyAddress(hostport string) Opt {
	return func(sc *ProxyConfig) error {
		var host, port string
		var err error
		if host, port, err = net.SplitHostPort(hostport); err != nil {
			return err
		}
		sc.Host = host
		sc.Port = port
		return nil
	}
}
