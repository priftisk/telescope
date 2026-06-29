package main

import (
	"context"
	"log"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

func getContainerHostAndPort(info container.InspectResponse) (string, string) {
	if info.Config == nil || info.Config.Labels == nil {
		return "", ""
	}
	var host, port string
	if host = info.Config.Labels["proxy.host"]; host == "" {
		return "", ""
	}

	if port = info.Config.Labels["proxy.port"]; port == "" {
		port = "80"
	}

	if info.NetworkSettings == nil {
		return "", ""
	}
	networks := info.NetworkSettings.Networks
	var containerIP string

	// Get the first non-empty IP
	for _, v := range networks {
		if v.IPAddress.IsValid() && !v.IPAddress.IsUnspecified() {
			containerIP = v.IPAddress.String()
			return host, containerIP + ":" + port
		}
	}
	return "", ""
}

func onStartup(ctx context.Context, apiClient *client.Client, rt *RouteTable) {
	containers, err := apiClient.ContainerList(ctx, client.ContainerListOptions{All: true})
	if err != nil {
		log.Fatal("failed to list containers:", err)
	}
	for _, c := range containers.Items {
		info, err := apiClient.ContainerInspect(ctx, c.ID, client.ContainerInspectOptions{})
		if err != nil {
			log.Printf("inspect failed for %s: %v", c.ID, err)
			continue
		}
		host, addr := getContainerHostAndPort(info.Container)

		if host == "" {
			continue
		}
		rt.Register(host, addr)
	}
}
