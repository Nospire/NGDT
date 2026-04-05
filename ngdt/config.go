package main

import "time"

const (
	OrchestratorURL   = "https://gdt.geekcom.org"
	SingBoxPath       = "./sing-box"
	SingBoxConfigPath = "/tmp/gdt-singbox.json"
	TunInterface      = "gdt0"
	TunnelCheckURL    = "https://ipinfo.io/json"
	HeartbeatInterval = 10 * time.Minute
)
