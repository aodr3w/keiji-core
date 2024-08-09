package logging

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/aodr3w/keiji-core/paths"
	"github.com/joho/godotenv"
)

var (
	SETTINGS     = paths.WORKSPACE_SETTINGS
	ROTATE_LOGS  bool
	LOG_MAX_SIZE string
)

func init() {
	err := godotenv.Load(SETTINGS)
	if err != nil {
		log.Println(err)
	}
	ROTATE_LOGS = os.Getenv("ROTATE_LOGS") == "1"
	LOG_MAX_SIZE = os.Getenv("LOG_MAX_SIZE")
}

// LogSettings provides config information
// for log rotation e.g wether to rotateLogs as well as maxLogSize
type LogSettings struct {
	Rotate  bool
	MaxSize int64
}

// NewLogSettings returns an instance of *LogSettings
// containing configuration information defined in WORKSPACE SETTINGS
func NewLogSettings() *LogSettings {
	logMaxSize, err := strconv.Atoi(LOG_MAX_SIZE)
	if err != nil {
		logMaxSize = 1024
	}
	return &LogSettings{
		Rotate:  ROTATE_LOGS,
		MaxSize: int64(logMaxSize),
	}
}

type Logger struct {
	logger         *slog.Logger
	LogsPath       string
	file           *os.File
	settings       *LogSettings
	fallbackLogger *slog.Logger
}

func NewFallbackLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, nil))
}
func NewStdoutLogger() *Logger {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	return &Logger{
		logger,
		"",
		nil,
		nil,
		NewFallbackLogger(),
	}
}

func NewFileLogger(out string) (*Logger, error) {
	path, err := getLogsPath(out)
	if err != nil {
		return nil, err
	}
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(file, nil))
	return &Logger{
		logger,
		path,
		file,
		NewLogSettings(),
		NewFallbackLogger(),
	}, nil
}

func (l *Logger) logWithLock(logFunc func()) {
	if err := l.lockFile(); err != nil {
		l.fallbackLogger.Error(fmt.Sprintf("Failed to lock log file: %v", err))
		return
	}
	defer l.unlockFile()
	if l.settings.Rotate {
		l.rotateLogs()
	}
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

func (l *Logger) rotateLogs() {
	if l.file == nil {
		return
	}
	fileInfo, err := l.file.Stat()
	if err != nil {
		l.fallbackLogger.Error(fmt.Sprintf("Failed to get log file info: %v", err))
		return
	}

	if fileInfo.Size() >= l.settings.MaxSize {
		timestamp := time.Now().Format("20060102150405")
		newPath := fmt.Sprintf("%s.%s", l.LogsPath, timestamp)
		newFile, err := os.OpenFile(newPath, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			l.fallbackLogger.Error(fmt.Sprintf("failed to create rotated log file: %v", err))
			return
		}
		defer newFile.Close()

		// Ensure all writes are completed before seeking
		if err := l.file.Sync(); err != nil {
			l.fallbackLogger.Error(fmt.Sprintf("Failed to sync log file: %v", err))
			return
		}
		if _, err := l.file.Seek(0, 0); err != nil {
			l.fallbackLogger.Error(fmt.Sprintf("failed to seek log file: %v", err))
		}
		if _, err = io.Copy(newFile, l.file); err != nil {
			l.fallbackLogger.Error(fmt.Sprintf("Failed to copy log content to rotated log file: %v", err))
			return
		}
		if err = l.file.Truncate(0); err != nil {
			l.fallbackLogger.Error(fmt.Sprintf("Failed to truncate log file: %v", err))
			return
		}

		if _, err = l.file.Seek(0, 0); err != nil {
			l.fallbackLogger.Error(fmt.Sprintf("Failed to seek log file after truncation: %v", err))
			return
		}
		return
	}
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
