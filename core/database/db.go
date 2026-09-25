package database

import (
	"fmt"
	"time"

	"github.com/gaucho-racing/sentinel/core/config"
	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/sentinel/core/pkg/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

var dbRetries = 0

func Init() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC", config.DatabaseHost, config.DatabaseUser, config.DatabasePassword, config.DatabaseName, config.DatabasePort)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
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
		if err := db.AutoMigrate(
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
		); err != nil {
			logger.SugarLogger.Fatalf("AutoMigration failed: %v", err)
		}
		if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_github_external_id ON auth_entity_external_auth (external_id) WHERE provider = 'GITHUB'").Error; err != nil {
			logger.SugarLogger.Fatalf("GitHub identity index migration failed: %v", err)
		}
		logger.SugarLogger.Infoln("AutoMigration complete")
		DB = db
	}
}
