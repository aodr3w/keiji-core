package utils

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"regexp"
	"syscall"
	"time"

	"log"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/aodr3w/keiji-core/paths"
	"github.com/joho/godotenv"
	"github.com/logrusorgru/aurora"
)

func GetHomeDir() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return usr.HomeDir
}

func PathExists(path string) (bool, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func CreateServiceLog(path string) error {
	err := CreateDir(filepath.Dir(path))
	if err != nil {
		return err
	}
	fileName := filepath.Base(path)
	if strings.Count(fileName, ".log") != 1 {
		return fmt.Errorf("invalid log file name %v should end with .log", fileName)
	}
	_, err = os.Create(path)
	if err != nil {
		return err
	}
	return nil
}
func CreateDir(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}
	return nil
}

func GetExecutable(taskName string) (string, error) {
	dir := paths.TASK_EXECUTABLE
	err := CreateDir(dir)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s.bin", dir, taskName), nil
}

func GetSourcePath(task string) string {
	sourcePath := filepath.Join(os.Getenv("HOME"), "keiji", "tasks", task)
	fmt.Println("sourcePath: ", sourcePath)
	if ok, err := PathExists(sourcePath); !ok {
		log.Fatalf("failed to get source directory: %v", err)
	}
	return sourcePath
}

func GetWd() string {
	var cwd string
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("Error:", err)
	}
	return cwd
}

func CopyDir(src string, dst string) error {
	return filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(src, path)

		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}

		return CopyFile(path, dstPath)

	})
}

func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	// Ensure the copied file is saved to disk
	err = destFile.Sync()
	if err != nil {
		return fmt.Errorf("failed to sync destination file: %w", err)
	}

	// Set executable permissions on the destination file
	err = os.Chmod(dst, 0755)
	if err != nil {
		return fmt.Errorf("failed to set permissions on destination file: %w", err)
	}

	return nil
}

type TaskLog struct {
	Content []string
}

func GetLogLines(logsPath string) (*TaskLog, error) {
	taskLog := TaskLog{}
	file, err := os.Open(logsPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Get file size to ensure the offset is valid
	fileInfo, err := file.Stat()
	if err != nil {
		log.Printf("error getting file info: %v\n", err.Error())
		return nil, err
	}
	fileSize := fileInfo.Size()

	// Calculate a valid offset to avoid the "invalid argument" error
	const readSize = 10 * 1024 // 10KB
	var offset int64 = -readSize
	if fileSize < readSize {
		offset = -fileSize // If the file is smaller than the intended read size, read the whole file
	}

	// Seek to the calculated offset from the end of the file
	_, err = file.Seek(offset, io.SeekEnd)
	if err != nil {
		log.Printf("error determining startPos: %v\n", err.Error())
		return nil, err
	}

	// No need to adjust startPos here since we calculated it based on the file size

	scanner := bufio.NewScanner(file)
	var lines []string

	// Read the file into a slice
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Get the last 100 lines
	if len(lines) > 100 {
		taskLog.Content = lines[len(lines)-100:]
	} else {
		taskLog.Content = lines
	}

	return &taskLog, nil
}

func HandleStopSignal(cleanup func()) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	log.Println("waiting for termination signal")
	sig := <-sigChan
	log.Printf("signal received: %s\n", sig)
	cleanup()
}

func isValidTimeZone(tz string) bool {
	// The regex checks for the format "City/Country"
	re := regexp.MustCompile(`^[A-Za-z]+/[A-Za-z_]+$`)
	return re.MatchString(tz)
}

func isValidSettings() bool {
	err := godotenv.Load(paths.WORKSPACE_SETTINGS)
	if err != nil {
		log.Println(aurora.Red(fmt.Sprintf("settings error: %v", err)))
		return false
	}

	DB_URL := os.Getenv("DB_URL")
	TIME_ZONE := os.Getenv("TIME_ZONE")

	if len(DB_URL) == 0 || (DB_URL != "default" && !strings.Contains(DB_URL, "postgres")) {
		log.Println(aurora.Red(fmt.Sprintf("invalid dbURL, %v must be `default` or valid postgres db url", DB_URL)))
		return false
	}

	if DB_URL != "default" && !IsValidPostgresURL(DB_URL) {
		log.Println(aurora.Red("invalid DB_ENGINE must be sqlite or postgres"))
		return false
	}

	if !isValidTimeZone(TIME_ZONE) {
		log.Println(aurora.Red("invalid TIME_ZONE format, must be `City/Country`"))
		return false
	}

	return true
}

func IsValidPostgresURL(url string) bool {
	regexPattern := `^postgres(?:ql)?:\/\/(?:[a-zA-Z0-9]+)(?::[^@]+)?@(?:[a-zA-Z0-9._-]+):(?:\d+)\/(?:[a-zA-Z0-9_-]+)$`
	re := regexp.MustCompile(regexPattern)
	return re.MatchString(url)
}

func IsInit() bool {
	//confirm work space has already been created
	conf := func(exists bool, err error) bool {
		if err != nil || !exists {
			return false
		}
		return true
	}

	return conf(PathExists(paths.TASKS_PATH)) &&
		conf(PathExists(paths.SYSTEM_ROOT)) &&
		conf(PathExists(paths.WORKSPACE)) &&
		conf(PathExists(paths.WORKSPACE_SETTINGS)) &&
		conf(PathExists(paths.WORKSPACE_MODULE)) &&
		isValidSettings()
}

func ParseTimeStr(t string) (time.Time, error) {
	var layout string
	tail := string(t[len(t)-2:])
	if strings.EqualFold(tail, "AM") || strings.EqualFold(tail, "PM") {
		layout = "03:04PM"
	} else {
		layout = "15:04"
	}
	pt, err := time.Parse(layout, t)
	if err != nil {
		return time.Time{}, fmt.Errorf("incorrect time value: %v  should be in format %v ", t, layout)
	}

	return pt, nil
}

func DeleteTaskExecutable(executable string) error {
	dir := filepath.Dir(executable)
	runFile := filepath.Join(dir, fmt.Sprintf("%v_run.bin", strings.ReplaceAll(filepath.Base(executable), ".bin", "")))
	for _, f := range []string{executable, runFile} {
		err := os.Remove(f)
		if err != nil {
			return err
		}
	}
	return nil
}

func DeleteTaskLog(logsPath string) error {
	exists, err := PathExists(logsPath)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("logs Path %v not found", logsPath)
	}
	return os.Remove(logsPath)
}
