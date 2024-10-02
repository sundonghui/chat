package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/sundonghui/chat/auth"
	"github.com/sundonghui/chat/model"
)

const (
	defaultStrength = 8
)

type DefaultUser struct {
	Username string
	Password string
}

type DatabaseOptions struct {
	Dialect    string
	Connection string

	PasswordStrength int

	DefaultUserList []DefaultUser
}

// GormDatabase is a wrapper for the gorm framework.
type GormDatabase struct {
	DB *gorm.DB
}

// Close closes the gorm database connection.
func (d *GormDatabase) Close() {
	if d == nil || d.DB == nil {
		log.Fatal("failed to close because database not exists")
	}
	// 获取原生数据库连接
	sqlDB, err := d.DB.DB()
	if err != nil {
		log.Fatalf("failed to get DB from GORM: %v", err)
	}
	// 确保在函数退出时关闭数据库连接
	defer func() {
		if err := sqlDB.Close(); err != nil {
			log.Printf("failed to close database: %v", err)
		}
	}()
}

var mkdirAll = os.MkdirAll

// New creates a new wrapper for the gorm database framework.
func New(options DatabaseOptions) (*GormDatabase, error) {
	createDirectoryIfSqlite(options.Dialect, options.Connection)

	// 打开数据库连接
	db, err := openDatabase(options.Dialect, options.Connection)
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}

	if err := db.AutoMigrate(new(model.User), new(model.Application), new(model.Message), new(model.Client), new(model.PluginConf)); err != nil {
		return nil, err
	}

	if err := prepareBlobColumn(options.Dialect, db); err != nil {
		return nil, err
	}

	var userCount int64
	db.Find(new(model.User)).Count(&userCount)
	if len(options.DefaultUserList) > 0 && userCount == 0 {
		createList := make([]model.User, 0, len(options.DefaultUserList))
		for _, defaultUser := range options.DefaultUserList {
			createList = append(createList, model.User{
				Name:  defaultUser.Username,
				Pass:  auth.CreatePassword(defaultUser.Password, defaultStrength),
				Admin: true})
		}
		if err := db.Save(createList).Error; err != nil {
			return nil, err
		}
	}

	return &GormDatabase{DB: db}, nil
}

func openDatabase(dialect, connectionString string) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	switch dialect {
	case "mysql":
		db, err = gorm.Open(mysql.Open(connectionString), &gorm.Config{})
	case "postgres":
		db, err = gorm.Open(postgres.Open(connectionString), &gorm.Config{})
	case "sqlite", "sqlite3":
		db, err = gorm.Open(sqlite.Open(connectionString), &gorm.Config{})
	default:
		return nil, fmt.Errorf("unsupported dialect: %s", dialect)
	}

	if err != nil {
		return nil, err
	}

	configureDatabase(db, dialect)

	return db, nil
}

func configureDatabase(db *gorm.DB, dialect string) {
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get DB from GORM: %v", err)
	}

	// We normally don't need that much connections, so we limit them. F.ex. mysql complains about
	// "too many connections", while load testing Gotify.
	sqlDB.SetMaxOpenConns(10)

	if dialect == "sqlite" {
		// We use the database connection inside the handlers from the http
		// framework, therefore concurrent access occurs. Sqlite cannot handle
		// concurrent writes, so we limit sqlite to one connection.
		// see https://github.com/mattn/go-sqlite3/issues/274
		sqlDB.SetMaxOpenConns(1) // 限制为一个连接
	}

	if dialect == "mysql" {
		// Mysql has a setting called wait_timeout, which defines the duration
		// after which a connection may not be used anymore.
		// The default for this setting on mariadb is 10 minutes.
		// See https://github.com/docker-library/mariadb/issues/113
		sqlDB.SetConnMaxLifetime(9 * time.Minute) // 设置连接最大生命周期
	}
}

func prepareBlobColumn(dialect string, db *gorm.DB) error {
	blobType := ""
	switch dialect {
	case "mysql":
		blobType = "longblob"
	case "postgres":
		blobType = "bytea"
	}
	if blobType != "" {
		for _, target := range []struct {
			TableName string
			Column    string
		}{
			{"messages", "extras"},
			{"plugin_confs", "config"},
			{"plugin_confs", "storage"},
		} {
			query := fmt.Sprintf("ALTER TABLE %s MODIFY %s %s", target.TableName, target.Column, blobType)
			if err := db.Exec(query).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func createDirectoryIfSqlite(dialect, connection string) {
	if dialect == "sqlite3" || dialect == "sqlite" {
		if _, err := os.Stat(filepath.Dir(connection)); os.IsNotExist(err) {
			if err := mkdirAll(filepath.Dir(connection), 0o777); err != nil {
				panic(err)
			}
		}
	}
}
