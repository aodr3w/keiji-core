package main

import "github.com/aodr3w/keiji-tasks/core"

func main() {
	/*
		    DEFINE FUNCTION SCHEDULE HERE
			example;
			 core.NewSchedule().Run().Every(10).Seconds().Build()
			)
	*/
	core.NewSchedule().Run().Every(10).Seconds().Build()
}
