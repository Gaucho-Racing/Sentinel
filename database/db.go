package database

import (
	"fmt"
	"sentinel/config"
	"sentinel/model"
	"sentinel/utils"
	"time"

	singlestore "github.com/singlestore-labs/gorm-singlestore"
	"gorm.io/gorm"
)

var DB *gorm.DB

var dbRetries = 0

func InitializeDB() error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=UTC", config.DatabaseUser, config.DatabasePassword, config.DatabaseHost, config.DatabasePort, config.DatabaseName)
	db, err := gorm.Open(singlestore.Open(dsn), &gorm.Config{})
	if err != nil {
		if dbRetries < 5 {
			dbRetries++
			utils.SugarLogger.Errorln("failed to connect database, retrying in 5s... ")
			time.Sleep(time.Second * 5)
			InitializeDB()
		} else {
			return fmt.Errorf("failed to connect database after 5 attempts")
		}
	} else {
		utils.SugarLogger.Infoln("Connected to database")
		db.AutoMigrate(
			&model.User{},
			&model.Subteam{},
			&model.UserSubteam{},
			&model.UserRole{},
			&model.UserAuth{},
			&model.UserLogin{},
			&model.UserActivity{},
			&model.ClientApplication{},
			&model.ClientApplicationRedirectURI{},
			&model.AuthorizationCode{},
			&model.RefreshToken{},
		)
		utils.SugarLogger.Infoln("AutoMigration complete")
		DB = db
	}
	return nil
}
