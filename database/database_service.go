package database

import (
	"fmt"
	"sentinel/config"
	"sentinel/model"
	"sentinel/utils"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

var dbRetries = 0

func InitializeDB() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=main port=%s sslmode=disable TimeZone=UTC", config.PostgresHost, config.PostgresUser, config.PostgresPassword, config.PostgresPort)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		if dbRetries < 15 {
			dbRetries++
			utils.SugarLogger.Errorln("failed to connect database, retrying in 5s... ")
			time.Sleep(time.Second * 5)
			InitializeDB()
		} else {
			utils.SugarLogger.Fatalln("failed to connect database after 15 attempts, terminating program...")
		}
	} else {
		utils.SugarLogger.Infoln("Connected to postgres database")
		db.AutoMigrate(&model.User{}, &model.Subteam{}, &model.UserSubteam{}, &model.UserRole{}, &model.UserAuth{}, &model.UserLogin{}, &model.UserActivity{}, &model.ClientApplication{}, &model.ClientApplicationRedirectURI{})
		utils.SugarLogger.Infoln("AutoMigration complete")
		DB = db
	}
}
