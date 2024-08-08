package main

import (
	"log"
	"path/filepath"

	"github.com/aodr3w/keiji-core/logging"
	"github.com/aodr3w/keiji-core/paths"
)

func main() {
	logging, err := logging.NewFileLogger(filepath.Join(paths.TASK_LOG_DIR("test"), "test"))
	if err != nil {
		log.Fatal(err)
	}
	for range 1000 {
		logging.Info("hello world")
	}

}
