package service

import (
	"github.com/gaucho-racing/sentinel/discord/database"
	"github.com/gaucho-racing/sentinel/discord/model"
	"github.com/gaucho-racing/ulid-go"
)

func GetAllDiscordReactions() ([]model.DiscordReaction, error) {
	var reactions []model.DiscordReaction
	if err := database.DB.Find(&reactions).Error; err != nil {
		return []model.DiscordReaction{}, err
	}
	return reactions, nil
}

func GetDiscordReactionByID(id string) (model.DiscordReaction, error) {
	var reaction model.DiscordReaction
	if err := database.DB.Where("id = ?", id).First(&reaction).Error; err != nil {
		return model.DiscordReaction{}, err
	}
	return reaction, nil
}

func GetDiscordReactionsByDiscordUserID(discordUserID string) ([]model.DiscordReaction, error) {
	var reactions []model.DiscordReaction
	if err := database.DB.Where("discord_user_id = ?", discordUserID).Find(&reactions).Error; err != nil {
		return []model.DiscordReaction{}, err
	}
	return reactions, nil
}

func GetDiscordReactionsByEntityID(entityID string) ([]model.DiscordReaction, error) {
	var reactions []model.DiscordReaction
	if err := database.DB.Where("entity_id = ?", entityID).Find(&reactions).Error; err != nil {
		return []model.DiscordReaction{}, err
	}
	return reactions, nil
}

func CreateDiscordReaction(reaction model.DiscordReaction) (model.DiscordReaction, error) {
	if reaction.ID == "" {
		reaction.ID = ulid.Make().Prefixed("drxn")
	}
	reaction.EntityID = GetEntityIDForDiscordUser(reaction.DiscordUserID)
	if err := database.DB.Create(&reaction).Error; err != nil {
		return model.DiscordReaction{}, err
	}
	return reaction, nil
}

func DeleteDiscordReaction(id string) error {
	if err := database.DB.Where("id = ?", id).Delete(&model.DiscordReaction{}).Error; err != nil {
		return err
	}
	return nil
}
