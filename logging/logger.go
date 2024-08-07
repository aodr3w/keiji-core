package logging

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"syscall"

	"github.com/aodr3w/keiji-core/paths"
)

var (
	SETTINGS = paths.WORKSPACE_SETTINGS
)

type Logger struct {
	logger   *slog.Logger
	LogsPath string
	file     *os.File
}

func NewStdoutLogger() *Logger {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	return &Logger{
		logger,
		"",
		nil,
	}
}

func NewFileLogger(out string) (*Logger, error) {
	path, err := getLogsPath(out)
	if err != nil {
		return nil, err
	}
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(file, nil))
	return &Logger{
		logger,
		path,
		file,
	}, nil
}

func (l *Logger) logWithLock(logFunc func()) {
	if err := l.lockFile(); err != nil {
		log.Printf("Failed to lock log file: %v", err)
		return
	}
	defer l.unlockFile()
	logFunc()
}

func (l *Logger) lockFile() error {
	if l.file != nil {
		return syscall.Flock(int(l.file.Fd()), syscall.LOCK_EX)
	}
	return nil
}

func (l *Logger) unlockFile() error {
	if l.file != nil {
		return syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN)
	}
	return nil
}

func (l *Logger) Info(text string, args ...interface{}) {
	l.logWithLock(func() {
		l.logger.Info(l.format(text, args...))
	})
}

func (l *Logger) Error(text string, args ...interface{}) {
	l.logWithLock(func() {
		l.logger.Error(l.format(text, args...))
	})

}

func (l *Logger) Warn(text string, args ...interface{}) {
	l.logWithLock(func() {
		l.logger.Warn(l.format(text, args...))
	})
}

func (l *Logger) Fatal(text string, args ...interface{}) {
	l.logWithLock(func() {
		l.Error(text, args...)
	})
	os.Exit(1)
}

func (l *Logger) format(text string, args ...interface{}) string {
	return fmt.Sprintf(text, args...)
}

func getLogsPath(path string) (string, error) {
	dir := filepath.Dir(path)
	//create dir here
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return "", err
	}
	if path[len(path)-3:] != "log" {
		path = fmt.Sprintf("%v.log", path)
	}
	return path, nil
}
