package main

import (
	"log"

	"github.com/aodr3w/keiji-core/tasks"
)

func Schedule() error {
	/*
		    DEFINE FUNCTION SCHEDULE HERE
			example;
			 core.NewSchedule().Run().Every(10).Seconds().Build()
			)
	*/
	log.Println("scheduling function...")
	return tasks.NewSchedule().Run().Every(10).Seconds().Build()
}
