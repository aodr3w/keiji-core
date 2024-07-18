package db

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/aodr3w/keiji-core/utils"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DatabaseType string

const (
	SQLite   DatabaseType = "sqlite"
	Postgres DatabaseType = "postgres"
)

type IDatabaseBackend interface {
	Connect() (*gorm.DB, error)
	AutoMigrate() error
}

type DatabaseBackend struct {
	DBType DatabaseType
	DBURL  string
}

func NewDatabaseBackend(dbType DatabaseType, dbURL string) *DatabaseBackend {
	return &DatabaseBackend{
		DBType: dbType,
		DBURL:  dbURL,
	}
}

func (dbBackend *DatabaseBackend) Connect() (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	switch dbBackend.DBType {
	case SQLite:
		err = utils.CreateDir(filepath.Dir(dbBackend.DBURL))
		if err != nil {
			log.Fatal(err)
		}
		db, err = gorm.Open(sqlite.Open(dbBackend.DBURL), &gorm.Config{})
	case Postgres:
		db, err = gorm.Open(postgres.Open(dbBackend.DBURL), &gorm.Config{})
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbBackend.DBType)
	}

	if err != nil {
		log.Printf("error opening %s db: %v", dbBackend.DBType, err)
	}
	return db, err
}

func (dbBackend *DatabaseBackend) AutoMigrate() error {
	db, err := dbBackend.Connect()
	if err != nil {
		return err
	}
	return db.AutoMigrate(&TaskModel{}, &UserModel{})
}
