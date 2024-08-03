package main

import (
	"flag"
	"fmt"
	"log"
)

/*This file is generated. Modify catiously*/
func main() {
	var err error
	//use flags to choose between run-schedule and run-function
	schedule := flag.Bool("schedule", false, "provide true to save task's schedule")
	run := flag.Bool("run", false, "provide true to run task's function")
	flag.Parse()
	if *schedule {
		err = Schedule()
	} else if *run {
		err = Function()
	} else {
		err = fmt.Errorf("valid arguments: --schedule, --run")
	}
	if err != nil {
		log.Fatal(err)
	}
}
