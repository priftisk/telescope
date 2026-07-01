package main

import "github.com/moby/moby/api/types/container"

type ContainerInfo struct {
	ContainerID     string
	ContainerIPAddr string
	Labels          Labels
}

func ExtractContainerData(info container.InspectResponse) (ContainerInfo, bool) {
	containerInfo := ContainerInfo{
		ContainerID: info.ID,
	}

	if info.Config == nil || info.Config.Labels == nil {
		return containerInfo, false
	}

	hostname, port, path := VerifyConfig(info.Config)

	// Container must have proxy.host label to be valid
	if hostname == "" {
		return containerInfo, false
	}

	// Default port if not specified in labels
	if port == "" {
		port = GetContainerPorts(info.NetworkSettings)
	}

	containerInfo.Labels = Labels{
		ProxyHost: hostname,
		ProxyPort: port,
		ProxyPath: path,
	}

	// Extract container IP if network settings exist
	if info.NetworkSettings != nil {
		containerInfo.ContainerIPAddr = GetContainerIP(info.NetworkSettings)

		// Container must have an IP address
		if containerInfo.ContainerIPAddr == "" {
			return containerInfo, false
		}
	}

	return containerInfo, true
}
