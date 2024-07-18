package dto

import (
	"time"
)

type TaskInfo struct {
	TaskID            string
	Name              string
	Description       string                 `json:"description"`
	Schedule          map[string]interface{} `json:"scheduleInfo"`
	NextExecutionTime *time.Time
	LastExecutionTime *time.Time
	Ch                chan bool
	RetryPolicy       map[string]int
	Status            string
	LogPath           string
	Slug              string
	Type              string
}

type UserInfo struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}
