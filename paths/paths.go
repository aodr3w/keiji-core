package paths

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aodr3w/keiji-core/constants"
)

var (
	SYSTEM_ROOT        = fmt.Sprintf("%v/%v", os.Getenv("HOME"), ".keiji")
	TASK_LOG           = fmt.Sprintf("%v/logs/tasks", SYSTEM_ROOT)
	DB                 = fmt.Sprintf("%v/db/keiji.db", SYSTEM_ROOT)
	SERVICE_LOGS       = fmt.Sprintf("%v/logs/services", SYSTEM_ROOT)
	TASK_EXECUTABLE    = fmt.Sprintf("%v/exec/tasks", SYSTEM_ROOT)
	SERVICE_EXECUTABLE = fmt.Sprintf("%v/exec/services", SYSTEM_ROOT)
	REPO_LOGS          = fmt.Sprintf("%v/%v.log", SERVICE_LOGS, "repo")
	TCP_BUS_LOGS       = fmt.Sprintf("%v/%v.log", SERVICE_LOGS, "tcp-bus")
	SCHEDULER_LOGS     = fmt.Sprintf("%v/%v.log", SERVICE_LOGS, "scheduler")
	EXECUTOR_LOGS      = fmt.Sprintf("%v/%v.log", SERVICE_LOGS, "executor")
	CLEANER_LOGS       = fmt.Sprintf("%v/%v.log", SERVICE_LOGS, "cleaner")
	HTTP_SERVER_LOGS   = fmt.Sprintf("%v/%v.log", SERVICE_LOGS, "http")
	PID_PATH           = func(name constants.Service) string {
		return fmt.Sprintf("%v/%v.pid", SERVICE_EXECUTABLE, name)
	}
	WORKSPACE          = filepath.Join(os.Getenv("HOME"), "keiji")
	TASKS_PATH         = filepath.Join(WORKSPACE, "tasks")
	WORKSPACE_SETTINGS = filepath.Join(WORKSPACE, "settings.conf")
	WORKSPACE_MODULE   = filepath.Join(WORKSPACE, "go.mod")
)
