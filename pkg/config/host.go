package config

import "net"

type Host struct {
	IP   net.IP `toml:"ip"`
	Name string `toml:"name"`
}
