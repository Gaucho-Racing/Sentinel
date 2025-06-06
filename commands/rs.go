package commands

import (
	"sentinel/config"
	"sentinel/service"
	"sentinel/utils"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func RemoveSubteam(args []string, s *discordgo.Session, m *discordgo.MessageCreate) {
	defer s.ChannelMessageDelete(m.ChannelID, m.ID)
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

	user := service.GetUserByID(guildMember.User.ID)
	if user.ID == "" || !(user.IsMember() || user.IsAlumni()) {
		// User not found
		go service.SendDisappearingMessage(m.ChannelID, "You must verify your account first! (`!verify <first name> <last name> <email>`)", 5*time.Second)
		return
	} else {
		counter := 0
		for _, arg := range args {
			ar := strings.ToLower(arg)
			a := []rune(ar)
			a[0] = []rune(strings.ToUpper(ar))[0]
			arg = string(a)
			role := service.GetSubteamByName(arg)
			if role.ID != "" {
				err = s.GuildMemberRoleRemove(m.GuildID, user.ID, role.ID)
				if err != nil {
					go service.SendDisappearingMessage(m.ChannelID, "Unexpected error occurred, please try again later!", 5*time.Second)
					utils.SugarLogger.Errorln(err)
				} else {
					counter++
				}
			} else {
				go service.SendDisappearingMessage(m.ChannelID, "Subteam `"+arg+"` not found!", 5*time.Second)
			}
		}
		if counter == 0 {
			go service.SendDisappearingMessage(m.ChannelID, "Command usage: `!rs <aero | business | chassis | data | electronics | powertrain | suspension>`", 5*time.Second)
		} else {
			go service.SendDisappearingMessage(m.ChannelID, "Removed "+strconv.Itoa(counter)+" subteam roles!", 5*time.Second)
		}
	}
}
