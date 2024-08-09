package constants

type ServiceStatus string

const (
	ONLINE  ServiceStatus = "ONLINE"
	OFFLINE ServiceStatus = "OFFLINE"
)

type Service string

var (
	REPO      Service = "repo"
	SCHEDULER Service = "scheduler"
	TCP_BUS   Service = "bus"
	SERVICES          = []Service{TCP_BUS, SCHEDULER}
)
