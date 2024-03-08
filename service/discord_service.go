package service

import (
	"github.com/bwmarrin/discordgo"
	"sentinel/config"
	"sentinel/model"
	"sentinel/utils"
)

var Discord *discordgo.Session

func ConnectDiscord() {
	dg, err := discordgo.New("Bot " + config.DiscordToken)
	if err != nil {
		utils.SugarLogger.Infoln("Error creating Discord session, ", err)
		return
	}
	Discord = dg
	_, err = Discord.ChannelMessageSend(config.DiscordLogChannel, ":white_check_mark: Sentinel v"+config.Version+" online! `[ENV = "+config.Env+"]`")
	if err != nil {
		utils.SugarLogger.Errorln("Error sending Discord message, ", err)
		return
	}
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
		Title: "New Account Created!",
		Color: 6609663,
		Author: &discordgo.MessageEmbedAuthor{
			IconURL: user.AvatarURL,
		},
		Fields: fields,
	})
	_, err := Discord.ChannelMessageSendEmbeds(config.DiscordLogChannel, embeds)
	if err != nil {
		utils.SugarLogger.Errorln(err.Error())
	}
}
