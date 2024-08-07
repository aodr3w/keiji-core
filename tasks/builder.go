package tasks

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/aodr3w/keiji-core/utils"
)

type Builder struct {
	T     *ScheduledTask
	guard *ExecutableGuard
}

func NewTaskBuilder() *Builder {
	return &Builder{
		nil,
		NewExecutableGuard(),
	}
}

/*creates the executable binary & saves task info in the database*/
func (b *Builder) Build(T *ScheduledTask) error {
	b.guard.Lock()
	defer b.guard.Unlock()
	b.T = T
	execPath, err := utils.GetExecutable(T.Name)
	if err != nil {
		return fmt.Errorf("failed to get executable task err: %v", err)
	}
	log.Println("task saved")
	sourcePath := utils.GetSourcePath(T.Name)
	log.Printf("executable path %v\n", execPath)
	log.Printf("source path %v\n", sourcePath)
	cmd := exec.Command("go", "build", "-o", execPath, sourcePath)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf(" failed to create executable: %v", err)
	}
	log.Println("executable created")

	b.T.executable = execPath
	b.T.LogsPath = T.LogsPath
	err = b.T.Save()
	if err != nil {
		return fmt.Errorf("task.save() error %v", err)
	} else {
		log.Println("task saved")
	}
	return nil
}
