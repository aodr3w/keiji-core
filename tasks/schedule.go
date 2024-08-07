package tasks

/*
this file should not be modified
*/
import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aodr3w/keiji-core/db"
	"github.com/aodr3w/keiji-core/logging"
	"github.com/aodr3w/keiji-core/paths"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

var (
	TASK_NAME        string
	TASK_DESCRIPTION string
)

/*
capture scheduling info
*/
type Schedule struct{}
type RunnableTask struct{}
type IntervalTask struct{ n int64 }

type TaskTime struct {
	day string
}
type TaskDay struct{}

type TaskType string

const (
	HMSTask     TaskType = "HMS"
	DayTimeTask TaskType = "DayTime"
)

type Action struct {
	n          int64
	unit       string
	day        string
	execTime   string
	taskType   TaskType
	executable string
}
type ScheduledTask struct {
	*Action
	Name         string
	Description  string
	scheduleInfo map[string]interface{}
	schedule     string
	Slug         string
	*logging.Logger
}

func (p *RetryPolicy) Map() map[string]int64 {
	return map[string]int64{
		"tries":   p.tries,
		"backoff": p.backoff,
		"delay":   p.delay,
	}
}
func NewSchedule() *Schedule {
	return &Schedule{}
}
func (st *ScheduledTask) Save() error {
	repo, err := db.NewRepo()
	if err != nil {
		return err
	}
	task_obj := db.TaskModel{}
	task_obj.ScheduleInfo = st.scheduleInfo
	task_obj.Name = st.Name
	task_obj.Description = st.Description
	task_obj.Type = db.TaskType(st.taskType)
	task_obj.TaskId = uuid.New().String()
	task_obj.Slug = strings.Join(strings.Split(strings.ToLower(st.Name), " "), "-")
	task_obj.LogPath = st.LogsPath
	task_obj.Schedule = st.schedule
	task_obj.Executable = st.E()
	return repo.SaveTask(&task_obj)
}

func (s *Schedule) Run() *RunnableTask {
	return &RunnableTask{}
}

func (s *Schedule) On() *TaskDay {
	return &TaskDay{}
}

func (d *TaskDay) Monday() *TaskTime {
	return &TaskTime{
		day: "Monday",
	}
}
func (d *TaskDay) Tuesday() *TaskTime {
	return &TaskTime{
		day: "Tuesday",
	}
}
func (d *TaskDay) Wednesday() *TaskTime {
	return &TaskTime{
		day: "Wednesday",
	}
}
func (d *TaskDay) Thursday() *TaskTime {
	return &TaskTime{
		day: "Thursday",
	}
}
func (d *TaskDay) Friday() *TaskTime {
	return &TaskTime{
		day: "Friday",
	}
}
func (d *TaskDay) Saturday() *TaskTime {
	return &TaskTime{
		day: "Saturday",
	}
}
func (d *TaskDay) Sunday() *TaskTime {
	return &TaskTime{
		day: "Sunday",
	}
}

func (tt *TaskTime) At(t string) *Action {
	layout := "15:04"
	if _, err := time.Parse(layout, t); err != nil {
		panic(fmt.Sprintf("incorrect time value: %v  should be in format %v ", t, layout))
	}
	return &Action{
		day:      tt.day,
		execTime: t,
		taskType: DayTimeTask,
	}
}

func (s *RunnableTask) Every(n int64) *IntervalTask {
	if n <= 0 {
		log.Fatal("n should be a non negative number")
	}
	return &IntervalTask{
		n,
	}
}

func (d *IntervalTask) Seconds() *Action {
	return &Action{
		n:        d.n,
		unit:     "seconds",
		taskType: HMSTask,
	}
}

func (d *IntervalTask) Minutes() *Action {
	return &Action{
		n:        d.n,
		unit:     "minutes",
		taskType: HMSTask,
	}
}

func (d *IntervalTask) Hours() *Action {
	return &Action{
		n:        d.n,
		unit:     "hours",
		taskType: HMSTask,
	}
}

func (a *Action) Build() error {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	TASK_NAME = os.Getenv("TASK_NAME")
	TASK_DESCRIPTION = os.Getenv("TASK_DESCRIPTION")
	name := TASK_NAME
	description := TASK_DESCRIPTION
	slug := strings.Join(strings.Split(strings.ToLower(name), " "), "-")
	log, err := logging.NewFileLogger(fmt.Sprintf("%v/%v", paths.TASK_LOG, slug))
	if err != nil {
		return err
	}
	var schedule []string
	var scheduledTask *ScheduledTask
	var scheduleInfo map[string]interface{}
	if a.taskType == HMSTask {
		scheduleInfo = getIntervalTaskSchedule(a)
	} else if a.taskType == DayTimeTask {
		scheduleInfo = getDayTimeTaskSchedule(a)
	}
	for k, v := range scheduleInfo {
		switch value := v.(type) {
		case int64:
			schedule = append(schedule, fmt.Sprintf("%s:%s", k, strconv.Itoa(int(value))))
		default:
			schedule = append(schedule, fmt.Sprintf("%s:%s", k, value))
		}
	}
	scheduledTask = &ScheduledTask{
		a,
		name,
		description,
		scheduleInfo,
		strings.Join(schedule, ","),
		slug,
		log,
	}
	return NewTaskBuilder().Build(scheduledTask)
}

func (a *Action) E() string {
	return a.executable
}
func (a *Action) N() int64 {
	return a.n
}

func (a *Action) Unit() string {
	return a.unit
}

func (a *Action) Type() TaskType {
	return a.taskType
}

func getIntervalTaskSchedule(a *Action) map[string]interface{} {
	scheduleInfo := make(map[string]interface{})
	scheduleInfo["units"] = a.unit
	scheduleInfo["interval"] = a.n
	return scheduleInfo
}

func getDayTimeTaskSchedule(a *Action) map[string]interface{} {
	scheduleInfo := make(map[string]interface{})
	scheduleInfo["day"] = a.day
	scheduleInfo["time"] = a.execTime
	return scheduleInfo
}
