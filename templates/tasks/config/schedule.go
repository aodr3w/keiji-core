package taskconfig

import (
	"log"

	"github.com/aodr3w/keiji-tasks/core"
)

func Schedule() error {
	/*
		    DEFINE FUNCTION SCHEDULE HERE
			example;
			 core.NewSchedule().Run().Every(10).Seconds().Build()
			)
	*/
	log.Println("scheduling function...")
	return core.NewSchedule().Run().Every(10).Seconds().Build()
}
