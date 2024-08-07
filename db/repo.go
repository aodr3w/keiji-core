package db

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/aodr3w/keiji-core/auth"
	"github.com/aodr3w/keiji-core/dto"
	"github.com/aodr3w/keiji-core/logging"
	"github.com/aodr3w/keiji-core/paths"
	"github.com/aodr3w/keiji-core/utils"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

type Repo struct {
	DB     *gorm.DB
	logger *logging.Logger
	mu     sync.Mutex
}

/*
NewRepo is a factory function that returns an instance of Repo, granting
the caller database access.
*/

func NewRepo() (*Repo, error) {
	log, err := logging.NewFileLogger(paths.REPO_LOGS)
	if err != nil {
		return nil, err
	}
	err = godotenv.Load(paths.WORKSPACE_SETTINGS)
	if err != nil {
		return nil, err
	}

	var dbType DatabaseType
	dbURL := os.Getenv("DB_URL")

	if dbURL == "default" {
		dbURL = paths.DB
		dbType = SQLite
	} else if utils.IsValidPostgresURL(dbURL) {
		dbType = Postgres
	} else {
		return nil, fmt.Errorf("database url must be either `default` or valid postgresql URL")
	}
	backend := DatabaseBackend{
		DBType: DatabaseType(dbType),
		DBURL:  dbURL,
	}

	gormDB, err := backend.Connect()
	if err != nil {
		log.Fatal("Failed to connect to the database: %v", err)
	}

	err = backend.AutoMigrate()
	if err != nil {
		return nil, err
	}
	repo := &Repo{
		gormDB,
		log,
		sync.Mutex{},
	}
	//insert defaultUser
	err = repo.insertDefaultUser()
	if err != nil {
		log.Warn("insert default user error %v", err.Error())
	}
	return repo, nil
}

/*
Close closes the DB instance/session attached to the repo
*/
func (r *Repo) Close() {
	db, err := r.DB.DB()
	if err != nil {
		r.logger.Error("failed to get db instance: %v", err)
		return
	}
	err = db.Close()
	if err != nil {
		r.logger.Error("error closing repo: %v", err)
	}
	r.logger.Info("repo closed successfully")
}

/*
SaveTask attempts to save a task to the database.If
it already exists, it will try to update the existing record
*/
func (r *Repo) SaveTask(task *TaskModel) error {
	// Check if the task already exists in the database
	r.mu.Lock()
	defer r.mu.Unlock()
	tx := r.DB.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	existingTask, err := r.GetTaskByName(task.Name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		r.logger.Error("Error checking existing task %v", err)
		tx.Rollback()
		return err
	}

	if existingTask != nil {
		r.logger.Info("Task already exists, updating: %v", task.Name)
		// Update the existing task fields
		existingTask.ScheduleInfo = task.ScheduleInfo
		existingTask.Schedule = task.Schedule
		existingTask.Description = task.Description
		existingTask.Type = task.Type
		existingTask.Executable = task.Executable
		existingTask.NextExecutionTime = task.NextExecutionTime
		existingTask.LastExecutionTime = task.LastExecutionTime
		existingTask.LogPath = task.LogPath
		// Update the task in the database
		if err := tx.Save(existingTask).Error; err != nil {
			r.logger.Error("Error occurred updating task %v: %v", task.Name, err)
			tx.Rollback()
			return err
		}
		r.logger.Info("Task updated successfully")
	} else {
		r.logger.Info("Task does not exist, creating new: %v", task.Name)
		// Create a new task in the database
		if err := tx.Create(task).Error; err != nil {
			r.logger.Fatal("Error occurred creating task %v: %v", task.Name, err)
			tx.Rollback()
			return err
		}
		r.logger.Info("Task created successfully")
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

/*HMSScheduleChanged returns true if the schedule info on an HMSTask has changed*/
func (r *Repo) HMSScheduleChanged(taskInfo *dto.TaskInfo, existingTask *TaskModel) bool {
	s1 := taskInfo.Schedule
	s2 := existingTask.ScheduleInfo
	return s1["interval"] != s2["interval"]
}

/*DayTimeTaskScheduleChanged returns true if the schedule info on a DayTimeTask has changed*/
func (r *Repo) DayTimeTaskScheduleChanged(taskInfo *dto.TaskInfo, existingTask *TaskModel) bool {
	s1 := taskInfo.Schedule
	s2 := existingTask.ScheduleInfo
	return s1["day"] != s2["day"] || s1["time"] != s2["time"]
}

/*IsHMSTask returns True if task is of type HMSTask*/
func (r *Repo) IsHMSTask(taskInfo *dto.TaskInfo) bool {
	_, ok := taskInfo.Schedule["interval"]
	return ok || taskInfo.Type == string(HMSTask)
}

/*IsDayTimeTask returns True if task is of type DayTime*/
func (r *Repo) IsDayTimeTask(taskInfo *dto.TaskInfo) bool {
	_, timeOk := taskInfo.Schedule["time"]
	_, dayOk := taskInfo.Schedule["day"]
	return timeOk && dayOk || taskInfo.Type == string(DayTimeTask)
}

/*
GetTaskByName queries the database for a record where
task.Name = Name
*/
func (r *Repo) GetTaskByName(name string) (*TaskModel, error) {
	var task TaskModel
	result := r.DB.First(&task, "name = ?", name) // Find first task with the given name
	if result.Error != nil {
		r.logger.Error("an error occured getting task %v", result.Error)
		return nil, result.Error
	}
	return &task, nil

}

/*
GetTaskByID queries the database for a record where
task.taskID = taskID
*/
func (r *Repo) GetTaskByID(taskID string) (*TaskModel, error) {
	var task TaskModel
	result := r.DB.First(&task, "task_id = ?", taskID)
	if result.Error != nil {
		r.logger.Error("an error occured getting task %v", result.Error)
		return nil, result.Error
	}
	return &task, nil
}

/*
ResetIsQueued sets the IsQueued field to false for all tasks in the database
*/
func (r *Repo) ResetIsQueued() {
	r.DB.Model(&TaskModel{}).Where("is_queued = ?", true).Update("is_queued", false)
}

/*
GetRunnableTask queries the database for all runnableTasks i.e where
is_queued=False, is_running=False & is_disabled=False
*/
func (r *Repo) GetRunnableTasks() ([]*TaskModel, error) {
	tasks := make([]*TaskModel, 0)
	if err := r.DB.Model(&TaskModel{}).Where(
		"is_error = ? AND is_queued = ? AND is_running = ? AND is_disabled = ?", false, false, false, false,
	).Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

/*
GetRunningTasks queries the database for all tasks where isRunning = True
*/
func (r *Repo) GetRunningTasks() ([]*TaskModel, error) {
	tasks := make([]*TaskModel, 0)
	if err := r.DB.Model(&TaskModel{}).Where(
		"is_running = ?", true,
	).Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

/*
GetAllTasks returns all tasks stored in the database
*/
func (r *Repo) GetAllTasks() ([]*TaskModel, error) {
	tasks := make([]*TaskModel, 0)
	if err := r.DB.Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

/*insertDefaultUser creates the default admin-user*/
func (r *Repo) insertDefaultUser() error {
	du := "admin"
	var count int64
	r.DB.Model(&UserModel{}).Where("user_name = ?", du).Count(&count)
	if count == 0 {
		hashedPassword, err := auth.HashPassword(du)
		if err != nil {
			r.logger.Fatal("failed to hash password %v", err)
		}
		token, err := auth.GenerateToken()
		if err != nil {
			r.logger.Fatal("failed to generate token %v", err)
		} else {
			r.logger.Info("token generated: %s", token)
		}
		defaultUser := UserModel{
			UserName: du,
			Password: hashedPassword,
			Token:    token,
		}

		result := r.DB.Create(&defaultUser)

		if result.Error != nil {
			return result.Error
		}
		r.logger.Info("default user created.")
	}
	return nil
}

/*
GetUserByName checks the database for a record where user.userName = userName
*/
func (r *Repo) GetUserByName(userName string) (*UserModel, error) {
	dbUser := &UserModel{}
	result := r.DB.Model(&UserModel{}).Where("user_name  = ? ", userName).First(dbUser)
	if result.Error != nil {
		return nil, result.Error
	}
	return dbUser, nil
}

/*UpdateUser updates user details in the database if the provided information is non empty*/
func (r *Repo) UpdateUser(currentUserName, currentUserPassword string, newUserInfo *dto.UserInfo) error {
	_, err := r.AuthUser(&dto.UserInfo{
		UserName: currentUserName,
		Password: currentUserPassword,
	})
	if err != nil {
		return err
	}
	existingUser, err := r.GetUserByName(currentUserName)
	if err != nil {
		return err
	}
	if len(newUserInfo.UserName) > 0 {
		existingUser.UserName = newUserInfo.UserName
	}
	if len(newUserInfo.Password) > 0 {
		h, err := auth.HashPassword(newUserInfo.Password)
		if err != nil {
			return err
		}
		existingUser.Password = h
		newToken, err := auth.GenerateToken()
		if err != nil {
			return err
		}
		existingUser.Token = newToken
	}
	v := r.DB.Save(existingUser)
	return v.Error
}

/*
AuthUser verifies the authentication information against the database
*/
func (r *Repo) AuthUser(userInfo *dto.UserInfo) (dbUser *UserModel, err error) {
	dbUser = &UserModel{}
	result := r.DB.Model(&UserModel{}).Where("user_name  = ? ", userInfo.UserName).First(dbUser)
	if result.Error != nil {
		r.logger.Error("IsUser error: %v", result.Error)
		return nil, result.Error
	}
	r.logger.Info("IsUser: %v", result)
	isValid := auth.IsValidPassword([]byte(dbUser.Password), userInfo.Password)
	if isValid {
		return dbUser, nil
	}
	return nil, fmt.Errorf("invalid username ,password combination")
}

/*VerifyToken checks if the provided token string exists in the database*/
func (r *Repo) VerifyToken(token string) (user *UserModel, err error) {
	result := r.DB.Model(&UserModel{}).Where("token = ?", token).First(&user)
	return user, result.Error
}

/*
SetIsRunning sets task.IsRunning field to value,
returning an updated *TaskModel and an error
*/
func (r *Repo) SetIsRunning(taskName string, value bool) (*TaskModel, error) {
	var task TaskModel
	//Find the task by taskName
	if err := r.DB.Where("name = ?", taskName).First(&task).Error; err != nil {
		fmt.Println("task not found:", err)
		return nil, err
	}

	if value {
		task.IsQueued = false
	}
	task.IsRunning = value
	if err := r.DB.Save(&task).Error; err != nil {
		fmt.Println("Failed to update task: ", err)
		return nil, err
	}
	return &task, nil
}

/*
SetIsError sets task.IsError field to value ,
returning *TaskModel and an error
*/
func (r *Repo) SetIsError(taskName string, value bool, err string) (*TaskModel, error) {
	var task TaskModel
	//Find the task by taskName
	if err := r.DB.Where("name = ?", taskName).First(&task).Error; err != nil {
		return nil, err
	}

	if value {
		//if true, set isRunning to false
		task.IsRunning = false
		task.IsQueued = false
	}
	task.IsError = value
	task.ErrorTxt = err
	if err := r.DB.Save(&task).Error; err != nil {
		fmt.Println("Failed to update task: ", err)
		return nil, err
	}
	return &task, nil
}

/*
SetIsQueued sets task.IsQueued field to value,
returning *TaskModel and an error
*/
func (r *Repo) SetIsQueued(taskName string, value bool) (*TaskModel, error) {
	var task TaskModel
	//Find the task by taskName
	if err := r.DB.Where("name = ?", taskName).First(&task).Error; err != nil {
		return nil, err
	}

	if value {
		//if true, set isRunning to false
		task.IsRunning = false
		task.IsError = false
	}
	task.IsQueued = value
	if err := r.DB.Save(&task).Error; err != nil {
		fmt.Println("Failed to update task: ", err)
		return nil, err
	}
	return &task, nil
}

/*
SetIsDisabled sets task.IsDisabled field to value,
returning *TaskModel and an error
*/
func (r *Repo) SetIsDisabled(taskName string, value bool) (*TaskModel, error) {
	var task TaskModel
	//Find the task by taskName
	if err := r.DB.Where("name = ?", taskName).First(&task).Error; err != nil {
		return nil, err
	}

	if value {
		//if true, set all other statusFields to false
		task.IsRunning = false
		task.IsError = false
		task.IsQueued = false
	}
	task.IsDisabled = value
	if err := r.DB.Save(&task).Error; err != nil {
		fmt.Println("Failed to update task: ", err)
		return nil, err
	}
	return &task, nil
}

/*DeleteTask attempts to delete the provided task from the database and returns an error*/
func (r *Repo) DeleteTask(task *TaskModel) error {
	return r.DB.Delete(&task, task.ID).Error
}

/*
UpdateExecutionTime sets NextExecutionTime to t and LastExecutionTime to sub (if
LastExecutionTime is currently nil)
*/
func (r *Repo) UpdateExecutionTime(task *TaskModel, t *time.Time, sub *time.Time) error {
	task.LastExecutionTime = task.NextExecutionTime
	if task.LastExecutionTime == nil {
		task.LastExecutionTime = sub
	}
	task.NextExecutionTime = t
	return r.SaveTask(task)
}
