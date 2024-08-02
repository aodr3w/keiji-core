package db

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/aodr3w/keiji-core/auth"
	"github.com/aodr3w/keiji-core/dto"
	"github.com/aodr3w/keiji-core/paths"
	"github.com/aodr3w/logger"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

type Repo struct {
	DB     *gorm.DB
	logger *logger.Logger
	mu     sync.Mutex
}

/*
NewRepo is a factory function that returns an instance of the repo. Granting the caller
database access
*/
func NewRepo() (*Repo, error) {
	log, err := logger.NewFileLogger(paths.REPO_LOGS)
	if err != nil {
		return nil, err
	}
	err = godotenv.Load(paths.WORKSPACE_SETTINGS)
	if err != nil {
		return nil, err
	}

	dbURL := os.Getenv("DBURL")
	dbType := os.Getenv("DB_ENGINE")

	if len(dbURL) == 0 {
		dbURL = paths.DB
	}
	if dbType != string(SQLite) && dbType != string(Postgres) {
		log.Warn("unsuppored DB_ENGINE provided, defaulting to sqllite")
		dbType = string(SQLite)
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
	err = repo.InsertDefaultUser()
	if err != nil {
		log.Warn("insert default user error %v", err.Error())
	}
	return repo, nil
}

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
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *Repo) HMSScheduleChanged(taskInfo *dto.TaskInfo, existingTask *TaskModel) bool {
	s1 := taskInfo.Schedule
	s2 := existingTask.ScheduleInfo
	return s1["interval"] != s2["interval"] || s1["intervalUnits"] != s2["intervalUnits"]
}

func (r *Repo) DayTimeTaskScheduleChanged(taskInfo *dto.TaskInfo, existingTask *TaskModel) bool {
	s1 := taskInfo.Schedule
	s2 := existingTask.ScheduleInfo
	return s1["Day"] != s2["Day"] || s1["time"] != s2["time"]
}

func (r *Repo) IsHMSTask(taskInfo *dto.TaskInfo) bool {
	_, ok := taskInfo.Schedule["interval"]
	return ok || taskInfo.Type == "HMS"
}

func (r *Repo) IsDayTimeTask(taskInfo *dto.TaskInfo) bool {
	_, ok := taskInfo.Schedule["time"]
	return ok || taskInfo.Type == "DayTime"
}

func (r *Repo) GetTaskByName(name string) (*TaskModel, error) {
	var task TaskModel
	result := r.DB.First(&task, "name = ?", name) // Find first task with the given name
	if result.Error != nil {
		r.logger.Error("an error occured getting task %v", result.Error)
		return nil, result.Error
	}
	return &task, nil

}

func (r *Repo) GetTaskByID(taskID string) (*TaskModel, error) {
	var task TaskModel
	result := r.DB.First(&task, "task_id = ?", taskID)
	if result.Error != nil {
		r.logger.Error("an error occured getting task %v", result.Error)
		return nil, result.Error
	}
	return &task, nil
}

func (r *Repo) ResetIsQueued() {
	r.DB.Model(&TaskModel{}).Where("is_queued = ?", true).Update("is_queued", false)
}

func (r *Repo) GetRunnableTasks() ([]*TaskModel, error) {
	tasks := make([]*TaskModel, 0)
	if err := r.DB.Model(&TaskModel{}).Where(
		"is_error = ? AND is_queued = ? AND is_running = ? AND is_disabled = ?", false, false, false, false,
	).Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *Repo) GetRunningTasks() ([]*TaskModel, error) {
	tasks := make([]*TaskModel, 0)
	if err := r.DB.Model(&TaskModel{}).Where(
		"is_running = ?", true,
	).Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *Repo) GetAllTasks() ([]*TaskModel, error) {
	tasks := make([]*TaskModel, 0)
	if err := r.DB.Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *Repo) InsertDefaultUser() error {
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
func (r *Repo) GetUserByName(userName string) (*UserModel, error) {
	dbUser := &UserModel{}
	result := r.DB.Model(&UserModel{}).Where("user_name  = ? ", userName).First(dbUser)
	if result.Error != nil {
		return nil, result.Error
	}
	return dbUser, nil
}

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

func (r *Repo) ValidateToken(token string) (user *UserModel, err error) {
	result := r.DB.Model(&UserModel{}).Where("token = ?", token).First(&user)
	return user, result.Error
}

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

func (r *Repo) DeleteTask(task *TaskModel) error {
	return r.DB.Delete(&task, task.ID).Error
}

func (r *Repo) UpdateExecutionTime(task *TaskModel, t *time.Time, sub *time.Time) error {
	task.LastExecutionTime = task.NextExecutionTime
	if task.LastExecutionTime == nil {
		task.LastExecutionTime = sub
	}
	task.NextExecutionTime = t
	return r.SaveTask(task)
}
