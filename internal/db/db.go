package db

import (
	"log"
	"os"
	"time"

	"github.com/artqqwr/bookslib/internal/models"
	"github.com/artqqwr/bookslib/pkg/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func New(config *config.Config) *gorm.DB {
	cfg := config.Database

	dsn := cfg.ConnectionString // "./database/database.db"

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Millisecond * 10, // Slow SQL threshold
			LogLevel:                  logger.Info,           // Log level
			IgnoreRecordNotFoundError: false,                 // Ignore ErrRecordNotFound error for logger
			Colorful:                  true,                  // Disable color
		},
	)

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})

	if err != nil {
		log.Fatal(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConnections)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConnections)

	return db
}

func AutoMigrate(db *gorm.DB) {
	db.AutoMigrate(
		&models.User{},
		&models.Follow{},
		&models.Review{},
		&models.Tag{},
	)
}
