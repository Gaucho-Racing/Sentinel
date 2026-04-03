package service

import (
	"github.com/bwmarrin/discordgo"
	"github.com/gaucho-racing/sentinel/discord/config"
	"github.com/gaucho-racing/sentinel/discord/pkg/logger"
)

var Discord *discordgo.Session

func ConnectDiscord() {
	dg, err := discordgo.New("Bot " + config.DiscordToken)
	if err != nil {
		logger.SugarLogger.Errorln("Error creating Discord session:", err)
		return
	}
	Discord = dg
	logger.SugarLogger.Infoln("Created Discord session")
}

func GetChannelName(channelID string) string {
	channel, err := Discord.Channel(channelID)
	if err != nil {
		logger.SugarLogger.Errorf("Failed to get channel %s: %v", channelID, err)
		return ""
	}
	return channel.Name
}
