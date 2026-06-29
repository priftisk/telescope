package main

import (
	"context"
	"log"

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
			log.Printf(
				"docker event received: action=%s container=%s",
				event.Action,
				event.Actor.ID,
			)

			switch event.Action {

			case "start", "restart":
				log.Printf("processing container start/restart: %s", event.Actor.ID)

				c, err := apiClient.ContainerInspect(ctx, event.Actor.ID, client.ContainerInspectOptions{})
				if err != nil {
					log.Printf("inspect failed for container=%s error=%v", event.Actor.ID, err)
					continue
				}

				host, addr := getContainerHostAndPort(c.Container)
				if host == "" {
					log.Printf("skipping container=%s (no proxy host resolved)", event.Actor.ID)
					continue
				}

				rt.Register(host, addr)
				log.Printf("registered route: host=%s addr=%s container=%s", host, addr, event.Actor.ID)

			case "die", "kill", "stop":
				log.Printf("deregistering container route: container=%s action=%s", event.Actor.ID, event.Action)

				rt.Deregister(event.Actor.ID)
			}

		case err := <-eventChan.Err:
			log.Printf("docker event stream error: %v", err)
			log.Fatal("event stream closed unexpectedly")

		case <-ctx.Done():
			log.Println("context cancelled, shutting down event watcher")
			return
		}
	}
}
