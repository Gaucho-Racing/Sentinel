package service

import (
	"sentinel/database"
	"sentinel/model"
)

func GetAllActivities() []model.UserActivity {
	var activities []model.UserActivity
	database.DB.Find(&activities)
	return activities
}

func GetActivitiesForUser(userID string) []model.UserActivity {
	var activities []model.UserActivity
	database.DB.Where("user_id = ?", userID).Find(&activities)
	return activities
}

func GetLastActivityForUser(userID string) model.UserActivity {
	var activity model.UserActivity
	database.DB.Where("user_id = ?", userID).Order("created_at asc").First(&activity)
	return activity
}

func GetActivityByID(activityID string) model.UserActivity {
	var activity model.UserActivity
	database.DB.Where("id = ?", activityID).Find(&activity)
	return activity
}

func CreateActivity(activity model.UserActivity) error {
	if result := database.DB.Create(&activity); result.Error != nil {
		return result.Error
	}
	return nil
}

func DeleteActivity(activityID string) error {
	if result := database.DB.Where("id = ?", activityID).Delete(&model.UserActivity{}); result.Error != nil {
		return result.Error
	}
	return nil
}
