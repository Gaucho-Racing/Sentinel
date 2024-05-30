package service

import (
	"fmt"
	"sentinel/config"
	"sentinel/model"
	"sentinel/utils"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

var Discord *discordgo.Session

func ConnectDiscord() {
	dg, err := discordgo.New("Bot " + config.DiscordToken)
	if err != nil {
		utils.SugarLogger.Infoln("Error creating Discord session, ", err)
		return
	}
	Discord = dg
	_, err = Discord.ChannelMessageSend(config.DiscordLogChannel, ":white_check_mark: Sentinel v"+config.Version+" online! `[ENV = "+config.Env+"]` `[PREFIX = "+config.Prefix+"]`")
	if err != nil {
		utils.SugarLogger.Errorln("Error sending Discord message, ", err)
		return
	}
}

func SendMessage(channelID string, message string) {
	_, err := Discord.ChannelMessageSend(channelID, message)
	if err != nil {
		utils.SugarLogger.Errorln(err.Error())
	}
}

func SendDisappearingMessage(channelID string, message string, delay time.Duration) {
	msg, err := Discord.ChannelMessageSend(channelID, message)
	if err != nil {
		utils.SugarLogger.Errorln(err.Error())
	}
	go DelayedMessageDelete(channelID, msg.ID, delay)
}

func DelayedMessageDelete(channelID string, messageID string, delay time.Duration) {
	time.Sleep(delay)
	_ = Discord.ChannelMessageDelete(channelID, messageID)
}

func DiscordLogNewUser(user model.User) {
	var embeds []*discordgo.MessageEmbed
	var fields []*discordgo.MessageEmbedField
	fields = append(fields, &discordgo.MessageEmbedField{
		Name:   "ID",
		Value:  user.ID,
		Inline: false,
	})
	fields = append(fields, &discordgo.MessageEmbedField{
		Name:   "Name",
		Value:  user.FirstName + " " + user.LastName,
		Inline: true,
	})
	fields = append(fields, &discordgo.MessageEmbedField{
		Name:   "Email",
		Value:  user.Email,
		Inline: false,
	})
	embeds = append(embeds, &discordgo.MessageEmbed{
		Title: "New Member Verified!",
		Color: 6609663,
		Author: &discordgo.MessageEmbedAuthor{
			IconURL: user.AvatarURL,
		},
		Fields: fields,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: user.AvatarURL,
		},
	})
	_, err := Discord.ChannelMessageSendEmbeds(config.DiscordLogChannel, embeds)
	if err != nil {
		utils.SugarLogger.Errorln(err.Error())
	}
}

func DiscordUserEmbed(user model.User, channelID string) {
	guildMember, err := Discord.GuildMember(config.DiscordGuild, user.ID)
	if err != nil {
		utils.SugarLogger.Errorln("User no longer in the server: " + err.Error())
		DeleteUser(user.ID)
		return
	}
	var topRole *discordgo.Role
	var roleStrings []string
	for _, role := range guildMember.Roles {
		r, _ := Discord.State.Role(config.DiscordGuild, role)
		roleStrings = append(roleStrings, r.Name)
		if topRole == nil || r.Position > topRole.Position {
			topRole = r
		}
	}
	if topRole == nil {
		utils.SugarLogger.Errorln("User has no roles, how are they even here lmao, deleting...")
		DeleteUser(user.ID)
		return
	}
	utils.SugarLogger.Infof("%s (%d) %d", topRole.Name, topRole.Position, topRole.Color)
	color := topRole.Color
	var embeds []*discordgo.MessageEmbed
	var fields []*discordgo.MessageEmbedField
	fields = append(fields, &discordgo.MessageEmbedField{
		Name:   "ID",
		Value:  user.ID,
		Inline: false,
	})
	fields = append(fields, &discordgo.MessageEmbedField{
		Name:   "Username",
		Value:  user.Username,
		Inline: false,
	})
	fields = append(fields, &discordgo.MessageEmbedField{
		Name:   "Email",
		Value:  user.Email,
		Inline: false,
	})
	fields = append(fields, &discordgo.MessageEmbedField{
		Name:   "Roles",
		Value:  strings.Join(roleStrings, ", "),
		Inline: false,
	})
	embeds = append(embeds, &discordgo.MessageEmbed{
		Title: fmt.Sprintf("%s %s", user.FirstName, user.LastName),
		Color: color,
		Author: &discordgo.MessageEmbedAuthor{
			IconURL: user.AvatarURL,
		},
		Fields: fields,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: user.AvatarURL,
		},
	})
	_, err = Discord.ChannelMessageSendEmbeds(channelID, embeds)
	if err != nil {
		utils.SugarLogger.Errorln(err.Error())
	}
}
