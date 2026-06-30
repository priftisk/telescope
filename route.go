package main

type Route struct {
	HostName      string `json:"hostname"`
	TargetAddress string `json:"address"`
	URLPath       string `json:"url_path"`
}
