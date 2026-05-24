package service

import (
	"time"

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

// SendDisappearingMessage posts a channel message and schedules its deletion
// after the given delay. Returns immediately; deletion happens in a goroutine.
func SendDisappearingMessage(channelID, content string, delay time.Duration) {
	msg, err := Discord.ChannelMessageSend(channelID, content)
	if err != nil {
		logger.SugarLogger.Errorf("Failed to send disappearing message in %s: %v", channelID, err)
		return
	}
	go delayedMessageDelete(channelID, msg.ID, delay)
}

func delayedMessageDelete(channelID, messageID string, delay time.Duration) {
	time.Sleep(delay)
	if err := Discord.ChannelMessageDelete(channelID, messageID); err != nil {
		logger.SugarLogger.Errorf("Failed to delete message %s in %s: %v", messageID, channelID, err)
	}
}

// SendDirectMessage opens a DM channel with the user and posts the content.
// Returns the sent message so callers can reference it (e.g., build a jump link).
func SendDirectMessage(userID, content string) (*discordgo.Message, error) {
	channel, err := Discord.UserChannelCreate(userID)
	if err != nil {
		return nil, err
	}
	return Discord.ChannelMessageSend(channel.ID, content)
}

// DMJumpURL returns a Discord jump link to a message in a DM channel.
func DMJumpURL(channelID, messageID string) string {
	return "https://discord.com/channels/@me/" + channelID + "/" + messageID
}
