package main

import (
	"context"
	"log"
	"log/slog"

	"github.com/moby/moby/client"
)

func watchEvents(ctx context.Context, apiClient client.APIClient, rt *RouteTable) {
	filterArgs := client.Filters{}
	filterArgs.Add("type", "container")

	eventOptions := client.EventsListOptions{
		Filters: filterArgs,
	}

	log.Println("starting docker event watcher")

	eventChan := apiClient.Events(ctx, eventOptions)

	for {
		select {
		case event := <-eventChan.Messages:

			switch event.Action {

			case "start", "restart":

				c, err := apiClient.ContainerInspect(ctx, event.Actor.ID, client.ContainerInspectOptions{})
				if err != nil {
					slog.Error("inspect failed for container", "container", event.Actor.ID, "error", err)
					continue
				}

				labels, containerIP := ExtractLabels(c.Container)
				if labels.IsValid() == false || containerIP == "" {
					slog.Info("skipping container(no proxy host resolved)", "container", event.Actor.ID)
					continue
				}
				rt.Register(labels, containerIP)

			case "die", "kill", "stop":

				rt.Deregister(event.Actor.ID)
			}

		case err := <-eventChan.Err:
			slog.Info("docker event stream error", "error", err.Error())
			log.Fatal("event stream closed unexpectedly")

		case <-ctx.Done():
			slog.Info("context cancelled, shutting down event watcher")
			return
		}
	}
}
