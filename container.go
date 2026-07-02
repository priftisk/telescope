package main

import (
	"github.com/moby/moby/api/types/container"
)

func GetContainerIP(network *container.NetworkSettings) string {
	networks := network.Networks
	var containerIP string = ""
	// Get the first non-empty IP
	for _, v := range networks {
		if v.IPAddress.IsValid() && !v.IPAddress.IsUnspecified() {
			containerIP = v.IPAddress.String()
			break
		}
	}
	return containerIP
}

func GetContainerPorts(network *container.NetworkSettings) string {
	ports := network.Ports
	if len(ports) == 1 {
		for port := range ports {
			return port.Port()
		}
	}

	return ""
}

func VerifyConfig(config *container.Config) (string, string, string) {
	var hostname, port, path string

	port = config.Labels[ProxyPort]
	if hostname = config.Labels[ProxyHost]; hostname == "" {
		hostname = "localhost"
	}
	if path = config.Labels[ProxyPath]; path == "" {
		path = "/"
	}
	return hostname, port, path
}
