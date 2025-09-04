package service

import (
	"sentinel/database"
	"sentinel/model"
	"sentinel/utils"
)

func CreateMailingListEntry(entry model.MailingList) error {
	if database.DB.Where("email = ?", entry.Email).Updates(&entry).RowsAffected == 0 {
		utils.SugarLogger.Infoln("New entry created with email: " + entry.Email)
		if result := database.DB.Create(&entry); result.Error != nil {
			return result.Error
		}
	} else {
		utils.SugarLogger.Infoln("Entry with email: " + entry.Email + " has been updated!")
	}

	return nil
}

func GetAllMailingListEntries() []model.MailingList {
	var entries []model.MailingList
	database.DB.Find(&entries)

	// Merge with sentinel users
	users := GetAllUsers()
	for _, user := range users {
		var entry model.MailingList
		entry.Email = user.Email
		entry.FirstName = user.FirstName
		entry.LastName = user.LastName
		entry.Role = user.GetHighestRole()
		entry.Organization = "Gaucho Racing"

		entries = append(entries, entry)
	}

	return entries
}

func GetExternalMailingListEntries() []model.MailingList {
	var entries []model.MailingList
	database.DB.Where("organization != ?", "Gaucho Racing").Find(&entries)

	return entries
}
