package db

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type TaskStatus string

type TaskType string

const (
	HMSTask     TaskType = "HMS"
	DayTimeTask TaskType = "DayTime"
)

type UserModel struct {
	gorm.Model
	UserName string
	Password string
	Token    string `gorm:"unique"`
}

type TaskModel struct {
	gorm.Model
	TaskId            string                 `gorm:"unique" json:"taskId"`
	Name              string                 `gorm:"unique" json:"name"`
	Description       string                 `gorm:"varchar(15)" json:"description"`
	ScheduleInfo      map[string]interface{} `gorm:"json" json:"scheduleInfo"`
	Schedule          string                 `gorm:"varchar(20)" json:"schedule"`
	LastExecutionTime *time.Time             `json:"lastExecutionTime"`
	NextExecutionTime *time.Time             `json:"nextExecutionTime"`
	LogPath           string                 `json:"logPath"`
	Slug              string                 `gorm:"unique" json:"slug"`
	Type              TaskType
	Executable        string
	IsRunning         bool
	IsQueued          bool
	IsError           bool
	IsDisabled        bool
	ErrorTxt          string
}

// Implement the Stringer interface for TaskModel
func (t TaskModel) String() string {
	lastExecution := "N/A"
	if t.LastExecutionTime != nil {
		lastExecution = t.LastExecutionTime.Format(time.RFC3339)
	}

	nextExecution := "N/A"
	if t.NextExecutionTime != nil {
		nextExecution = t.NextExecutionTime.Format(time.RFC3339)
	}

	return fmt.Sprintf(
		"TaskModel:\n"+
			"ID: %d\n"+
			"TaskId: %s\n"+
			"Name: %s\n"+
			"Description: %s\n"+
			"Schedule: %s\n"+
			"LastExecutionTime: %s\n"+
			"NextExecutionTime: %s\n"+
			"LogPath: %s\n"+
			"Slug: %s\n"+
			"Type: %s\n"+
			"Executable: %s\n"+
			"IsRunning: %t\n"+
			"IsQueued: %t\n"+
			"IsError: %t\n"+
			"IsDisabled: %t\n"+
			"ErrorTxt: %s\n",
		t.ID, t.TaskId, t.Name, t.Description, t.Schedule, lastExecution, nextExecution, t.LogPath, t.Slug, t.Type, t.Executable, t.IsRunning, t.IsQueued, t.IsError, t.IsDisabled, t.ErrorTxt,
	)
}

// validation hook
func (t *TaskModel) BeforeSave(tx *gorm.DB) (err error) {

	if t.NextExecutionTime != nil {
		*t.NextExecutionTime = t.NextExecutionTime.Truncate(time.Second)
	}

	if t.LastExecutionTime != nil {
		*t.LastExecutionTime = t.LastExecutionTime.Truncate(time.Second)
	}

	return nil
}
