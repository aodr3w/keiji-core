package main

import (
	"log"

	"github.com/aodr3w/keiji-core/db"
)

func main() {
	repo, err := db.NewRepo()
	if err != nil {
		log.Fatal(err)
	}
	// tasks, err := repo.GetAllTasks()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// log.Println("all tasks:", tasks)
	defer repo.Close()
}
