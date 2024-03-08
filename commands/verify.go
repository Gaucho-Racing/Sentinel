package commands

import (
	"github.com/bwmarrin/discordgo"
	"sentinel/model"
	"sentinel/service"
	"sentinel/utils"
	"strings"
	"time"
)

func Verify(args []string, s *discordgo.Session, m *discordgo.MessageCreate) {
	defer s.ChannelMessageDelete(m.ChannelID, m.ID)
	// Get user info
	guildMember, err := s.GuildMember(m.GuildID, m.Author.ID)
	if err != nil {
		utils.SugarLogger.Errorln(err)
		return
	}
	isOfficer := false
	for _, role := range guildMember.Roles {
		if role == "812948550819905546" {
			isOfficer = true
			break
		}
	}

	user := service.GetUserByID(m.Author.ID)
	if user.ID == "" {
		// User not found
		if len(args) < 3 {
			go service.SendDisappearingMessage(m.ChannelID, "Command usage: `!verify <first name> <last name> <email>`", 5*time.Second)
			return
		}
		emailIndex := -1
		// find email, extract first name and last name from that
		for i, arg := range args {
			if strings.Contains(arg, "@ucsb.edu") || isOfficer && strings.Contains(arg, "@") {
				emailIndex = i
			}
		}
		if emailIndex == -1 {
			go service.SendDisappearingMessage(m.ChannelID, "Email must be a valid UCSB email", 5*time.Second)
			return
		}
		id := m.Author.ID
		firstName := strings.Join(args[:emailIndex-1], " ")
		lastName := args[emailIndex-1]
		email := args[emailIndex]
		// check if id flag is present
		if len(args) > emailIndex+1 {
			// last arg is id
			id = args[emailIndex+1]
			if !isOfficer {
				go service.SendDisappearingMessage(m.ChannelID, "You must be an officer to verify someone else!", 5*time.Second)
				return
			}
		}
		// finally create user
		service.CreateUser(model.User{
			ID:        id,
			Username:  m.Author.Username,
			FirstName: firstName,
			LastName:  lastName,
			Email:     email,
			AvatarURL: m.Author.AvatarURL("256"),
			Verified:  false,
			UpdatedAt: time.Time{},
			CreatedAt: time.Time{},
		})
		// rename user
		err = s.GuildMemberNickname(m.GuildID, id, firstName)
		if err != nil {
			utils.SugarLogger.Errorln(err)
		}
		// TODO: google drive access

		// assign member role
		err = s.GuildMemberRoleAdd(m.GuildID, id, "820467859477889034")
		if err != nil {
			utils.SugarLogger.Errorln(err)
		}
		go service.SendDisappearingMessage(m.ChannelID, "You have been verified! Welcome to the server <@"+id+">!", 5*time.Second)
	} else {
		//s.ChannelMessageSend(m.ChannelID, "You are already verified!")
		service.DeleteUser(m.Author.ID)
		Verify(args, s, m)
	}
}
