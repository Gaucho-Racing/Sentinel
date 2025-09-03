package service

import (
	"errors"
	"sentinel/database"
	"sentinel/model"

	"github.com/google/uuid"
)

func CreateMailingListEntry(entry model.MailingList) (model.MailingList, error) {
	// Check for existing email entry
	var existingEntry model.MailingList
	err := database.DB.Where("email = ?", entry.Email).First(&existingEntry).Error
	if err == nil { //
		return model.MailingList{}, errors.New("this email is already on the mailing list")
	}

	entry.ID = uuid.NewString()

	if result := database.DB.Create(&entry); result.Error != nil {
		return model.MailingList{}, result.Error
	}

	return entry, nil
}

func GetAllMailingListEntries() []model.MailingList {
	var entries []model.MailingList
	database.DB.Find(&entries)
	return entries
}
