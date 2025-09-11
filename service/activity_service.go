package service

import (
	"sentinel/database"
	"sentinel/model"
	"time"
)

func GetAllActivities() []model.UserActivity {
	var activities []model.UserActivity
	database.DB.Find(&activities)
	return activities
}

func GetActivitiesForUser(userID string) []model.UserActivity {
	var activities []model.UserActivity
	database.DB.Where("user_id = ?", userID).Order("created_at asc").Find(&activities)
	return activities
}

func GetLastActivityForUser(userID string) model.UserActivity {
	var activity model.UserActivity
	database.DB.Where("user_id = ?", userID).Order("created_at desc").First(&activity)
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

// ActivityCount represents aggregated counts for charting
type ActivityCount struct {
	Date   string `json:"date"`
	Action string `json:"action"`
	Count  int    `json:"count"`
}

// GetActivityCountsByDayForUser aggregates last 90 days of activity (messages/reactions)
func GetActivityCountsByDayForUser(userID string) []ActivityCount {
	end := time.Now()
	start := end.AddDate(0, 0, -89)

	buckets := make(map[string]map[string]int)
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		key := d.Format("2006-01-02")
		buckets[key] = map[string]int{"message": 0, "reaction": 0}
	}

	var activities []model.UserActivity
	database.DB.Where("user_id = ? AND created_at >= ?", userID, start).Find(&activities)
	for _, a := range activities {
		key := a.CreatedAt.Format("2006-01-02")
		if _, ok := buckets[key]; !ok {
			buckets[key] = map[string]int{}
		}
		buckets[key][a.Action] = buckets[key][a.Action] + 1
	}

	out := make([]ActivityCount, 0, len(buckets)*2)
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		key := d.Format("2006-01-02")
		counts := buckets[key]
		for action, count := range counts {
			out = append(out, ActivityCount{Date: key, Action: action, Count: count})
		}
	}
	return out
}
