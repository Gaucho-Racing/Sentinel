package database

import (
	"fmt"
	"time"

	"github.com/gaucho-racing/sentinel/core/config"
	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/sentinel/core/observability"
	"github.com/gaucho-racing/sentinel/core/pkg/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

var dbRetries = 0

func Init() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC", config.DatabaseHost, config.DatabaseUser, config.DatabasePassword, config.DatabaseName, config.DatabasePort)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: observability.NewDatabaseLogger(config.DatabaseSlowQueryThreshold),
	})
	if err != nil {
		if dbRetries < 5 {
			dbRetries++
			logger.SugarLogger.Errorln("failed to connect database, retrying in 5s... ")
			time.Sleep(time.Second * 5)
			Init()
		} else {
			logger.SugarLogger.Fatalf("failed to connect database after 5 attempts")
		}
	} else {
		logger.SugarLogger.Infoln("Connected to database")
		if config.DatabaseEnableQueryStatistics {
			enableQueryStatistics(db)
		}
		db.AutoMigrate(
			&model.Entity{},
			&model.EntityEmail{},
			&model.EntityPhone{},
			&model.EntityExternalAuth{},
			&model.PhoneLoginCode{},
			&model.EmailLoginCode{},
			&model.Token{},
			&model.User{},
			&model.Application{},
			&model.ApplicationGroup{},
			&model.ApplicationRedirectURI{},
			&model.EntityLogin{},
			&model.ServiceAccount{},
			&model.Group{},
			&model.GroupMember{},
			&model.GroupJoinRequest{},
			&model.GroupJoinRequestComment{},
			&model.GroupOwner{},
			&model.GroupConditionalBinding{},
			&model.SigningKey{},
			&model.AuditEvent{},
		)
		logger.SugarLogger.Infoln("AutoMigration complete")
		DB = db
	}
}

func enableQueryStatistics(db *gorm.DB) {
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS pg_stat_statements").Error; err != nil {
		logger.SugarLogger.Warnf("Failed to enable pg_stat_statements: %v", err)
		return
	}
	var available int
	if err := db.Raw("SELECT 1 FROM pg_stat_statements LIMIT 1").Scan(&available).Error; err != nil {
		logger.SugarLogger.Warnf("pg_stat_statements is installed but unavailable: %v", err)
		return
	}
	logger.SugarLogger.Infoln("pg_stat_statements is enabled")
}
