package main

import (
	"context"
	"log"
	"log/slog"

	"github.com/moby/moby/client"
)

func (s *Server) watchEvents(ctx context.Context) {
	filterArgs := client.Filters{}
	filterArgs.Add("type", "container")

	eventOptions := client.EventsListOptions{
		Filters: filterArgs,
	}

	log.Println("starting docker event watcher")

	eventChan := s.dockerClient.Events(ctx, eventOptions)

	for {
		select {
		case event := <-eventChan.Messages:

			switch event.Action {

			case "start", "restart":

				c, err := s.dockerClient.ContainerInspect(ctx, event.Actor.ID, client.ContainerInspectOptions{})
				if err != nil {
					slog.Error("inspect failed for container", "container", event.Actor.ID, "error", err)
					continue
				}

				container, valid := ExtractContainerData(c.Container)
				if !valid {
					slog.Info("skipping container(no proxy host resolved)", "container", event.Actor.ID)
					continue
				}
				s.routeTable.Register(container)

			case "die", "kill", "stop":

				s.routeTable.Deregister(event.Actor.ID)
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
