package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
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

func VerifyConfig(config *container.Config) (string, string, string) {
	var hostname, port, path string

	if port = config.Labels[ProxyPort]; port == "" {
		port = "80"
	}
	if hostname = config.Labels[ProxyHost]; hostname == "" {
		hostname = "localhost"
	}
	if path = config.Labels[ProxyPath]; path == "" {
		path = ""
	}
	return hostname, port, path
}

func ExtractLabels(info container.InspectResponse) (Labels, string) {
	var labels Labels = Labels{ProxyHost: "", ProxyPort: "", ProxyPath: ""}
	if info.Config == nil || info.Config.Labels == nil {
		return labels, ""
	}
	hostname, port, path := VerifyConfig(info.Config)

	if info.NetworkSettings == nil {
		return labels, ""
	}
	containerIP := GetContainerIP(info.NetworkSettings)
	labels.ProxyHost = hostname
	labels.ProxyPort = port
	labels.ProxyPath = path
	fmt.Printf("%+v\n", labels)
	return labels, containerIP
}

func onStartup(ctx context.Context, apiClient *client.Client, rt *RouteTable) {
	slog.Info("Seeding route table")
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
		labels, containerIP := ExtractLabels(info.Container)
		if labels.IsValid() == false || containerIP == "" {
			continue
		}

		rt.Register(labels, containerIP)
	}
}
