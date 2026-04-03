package service

import (
	"github.com/gaucho-racing/sentinel/discord/database"
	"github.com/gaucho-racing/sentinel/discord/model"
	"github.com/gaucho-racing/ulid-go"
)

func GetAllDiscordMessages() ([]model.DiscordMessage, error) {
	var messages []model.DiscordMessage
	if err := database.DB.Find(&messages).Error; err != nil {
		return []model.DiscordMessage{}, err
	}
	return messages, nil
}

func GetDiscordMessageByID(id string) (model.DiscordMessage, error) {
	var message model.DiscordMessage
	if err := database.DB.Where("id = ?", id).First(&message).Error; err != nil {
		return model.DiscordMessage{}, err
	}
	return message, nil
}

func GetDiscordMessagesByDiscordUserID(discordUserID string) ([]model.DiscordMessage, error) {
	var messages []model.DiscordMessage
	if err := database.DB.Where("discord_user_id = ?", discordUserID).Find(&messages).Error; err != nil {
		return []model.DiscordMessage{}, err
	}
	return messages, nil
}

func GetDiscordMessagesByEntityID(entityID string) ([]model.DiscordMessage, error) {
	var messages []model.DiscordMessage
	if err := database.DB.Where("entity_id = ?", entityID).Find(&messages).Error; err != nil {
		return []model.DiscordMessage{}, err
	}
	return messages, nil
}

func CreateDiscordMessage(message model.DiscordMessage) (model.DiscordMessage, error) {
	if message.ID == "" {
		message.ID = ulid.Make().Prefixed("dmsg")
	}
	message.EntityID = GetEntityIDForDiscordUser(message.DiscordUserID)
	if err := database.DB.Create(&message).Error; err != nil {
		return model.DiscordMessage{}, err
	}
	return message, nil
}

func DeleteDiscordMessage(id string) error {
	if err := database.DB.Where("id = ?", id).Delete(&model.DiscordMessage{}).Error; err != nil {
		return err
	}
	return nil
}
