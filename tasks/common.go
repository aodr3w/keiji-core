package tasks

import "sync"

type ExecutableGuard struct {
	mu sync.Mutex
}

var (
	executableGuardInstance *ExecutableGuard
	once                    sync.Once
)

func NewExecutableGuard() *ExecutableGuard {
	once.Do(func() {
		executableGuardInstance = &ExecutableGuard{}
	})
	return executableGuardInstance
}

func (eg *ExecutableGuard) Lock() {
	eg.mu.Lock()
}

func (eg *ExecutableGuard) Unlock() {
	eg.mu.Unlock()
}
