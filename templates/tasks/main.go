package main

import (
	"flag"
	"fmt"
	"log"

	taskconfig "github.com/aodr3w/keiji-core/templates/tasks/config"
)

/*This file is generated. Modify catiously*/
func main() {
	var err error
	//use flags to choose between run-schedule and run-function
	schedule := flag.Bool("schedule", false, "provide true to save task's schedule")
	run := flag.Bool("run", false, "provide true to run task's function")
	flag.Parse()
	if *schedule {
		err = taskconfig.Schedule()
	} else if *run {
		err = taskconfig.Function()
	} else {
		err = fmt.Errorf("valid arguments: --schedule, --run")
	}
	if err != nil {
		log.Fatal(err)
	}
}
