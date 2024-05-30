package service

import "sentinel/model"

func GetAllActivities() []model.UserActivity {
	var activities []model.UserActivity
	DB.Find(&activities)
	return activities
}

func GetActivitiesForUser(userID string) []model.UserActivity {
	var activities []model.UserActivity
	DB.Where("user_id = ?", userID).Find(&activities)
	return activities
}

func GetLastActivityForUser(userID string) model.UserActivity {
	var activity model.UserActivity
	DB.Where("user_id = ?", userID).Order("created_at desc").Last(&activity)
	return activity
}

func GetActivityByID(activityID string) model.UserActivity {
	var activity model.UserActivity
	DB.Where("id = ?", activityID).Find(&activity)
	return activity
}

func CreateActivity(activity model.UserActivity) error {
	if result := DB.Create(&activity); result.Error != nil {
		return result.Error
	}
	return nil
}

func DeleteActivity(activityID string) error {
	if result := DB.Where("id = ?", activityID).Delete(&model.UserActivity{}); result.Error != nil {
		return result.Error
	}
	return nil
}
