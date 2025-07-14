package commands

import (
	"sentinel/config"
	"sentinel/service"
	"sentinel/utils"
	"time"

	"github.com/bwmarrin/discordgo"
)

func Whois(args []string, s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.GuildID != config.DiscordGuild {
		m.GuildID = config.DiscordGuild
	}
	// Get user info
	guildMember, err := s.GuildMember(m.GuildID, m.Author.ID)
	if err != nil {
		utils.SugarLogger.Errorln(err)
		go service.SendDisappearingMessage(m.ChannelID, "Unexpected error occurred, please try again later!", 5*time.Second)
		return
	}

	user := service.GetUserByID(m.Author.ID)
	if user.ID == "" || !(user.IsMember() || user.IsAlumni()) {
		// User not found
		go service.SendDisappearingMessage(m.ChannelID, "You must verify your account first! (`!verify <first name> <last name> <email>`)", 5*time.Second)
		return
	} else {
		if len(args) < 1 {
			go service.SendDisappearingMessage(m.ChannelID, "Command usage: `!whois <id / username / email>`", 5*time.Second)
			return
		}
		if !utils.IsInnerCircle(guildMember.Roles) {
			go service.SendDisappearingMessage(m.ChannelID, "You do not have access to this command!", 5*time.Second)
			return
		}
		user := service.GetUserByID(args[0])
		if user.ID == "" {
			user = service.GetUserByUsername(args[0])
			if user.ID == "" {
				user = service.GetUserByEmail(args[0])
				if user.ID == "" {
					utils.SugarLogger.Infof("User not found: %s, attempting to search...", args[0])
					searchedUsers := service.SearchUsers(args[0])
					if len(searchedUsers) == 0 {
						go service.SendMessage(m.ChannelID, "User not found!")
						return
					} else {
						for _, u := range searchedUsers {
							utils.SugarLogger.Infof("User found: %s", u.Username)
							service.DiscordUserEmbed(u, m.ChannelID)
						}
						return
					}
				}
			}
		}
		utils.SugarLogger.Infof("User found: %s", user.ID)
		service.DiscordUserEmbed(user, m.ChannelID)
	}
}
