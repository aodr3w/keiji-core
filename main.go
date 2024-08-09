package main

import (
	"log"

	"github.com/aodr3w/keiji-core/tasks"
)

func main() {
	err := tasks.NewSchedule().On().Friday().At("10:00PM").Build()
	if err != nil {
		log.Println(err)
	}

}
