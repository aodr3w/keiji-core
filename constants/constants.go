package constants

import "strings"

type ServiceStatus string

const (
	ONLINE  ServiceStatus = "ONLINE"
	OFFLINE ServiceStatus = "OFFLINE"
)

type Service string

var (
	HTTP      Service = "http"
	SCHEDULER Service = "scheduler"
	CLEANER   Service = "cleaner"
	TCP_BUS   Service = "bus"
	SERVICES          = []Service{HTTP, SCHEDULER, CLEANER, TCP_BUS}
)

func IsService(name string) (serviceName Service, isService bool) {
	for _, s := range SERVICES {
		if strings.EqualFold(string(s), name) {
			return s, true
		}
	}
	return "", false
}
