package constants

import "strings"

type ServiceStatus string

const (
	ONLINE  ServiceStatus = "ONLINE"
	OFFLINE ServiceStatus = "OFFLINE"
)

type Service string

var (
	REPO      Service = "repo"
	SERVER    Service = "server"
	SCHEDULER Service = "scheduler"
	TCP_BUS   Service = "bus"
	SERVICES          = []Service{SERVER, SCHEDULER, TCP_BUS}
)

func IsService(name string) (serviceName Service, isService bool) {
	for _, s := range SERVICES {
		if strings.EqualFold(string(s), name) {
			return s, true
		}
	}
	return "", false
}
