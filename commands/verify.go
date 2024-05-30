package commands

import (
	"sentinel/model"
	"sentinel/service"
	"sentinel/utils"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func Verify(args []string, s *discordgo.Session, m *discordgo.MessageCreate) {
	defer s.ChannelMessageDelete(m.ChannelID, m.ID)
	// Get user info
	guildMember, err := s.GuildMember(m.GuildID, m.Author.ID)
	if err != nil {
		utils.SugarLogger.Errorln(err)
		go service.SendDisappearingMessage(m.ChannelID, "Unexpected error occurred, please try again later!", 5*time.Second)
		return
	}

	if len(args) < 3 {
		go service.SendDisappearingMessage(m.ChannelID, "Command usage: `!verify <first name> <last name> <email>`", 5*time.Second)
		return
	}
	emailIndex := -1
	// find email, extract first name and last name from that
	for i, arg := range args {
		if strings.Contains(arg, "@ucsb.edu") || utils.IsInnerCircle(guildMember.Roles) && strings.Contains(arg, "@") {
			emailIndex = i
		}
	}
	if emailIndex == -1 {
		go service.SendDisappearingMessage(m.ChannelID, "Email must be a valid UCSB email", 5*time.Second)
		return
	}
	id := m.Author.ID
	firstName := args[0]
	lastName := strings.Join(args[1:emailIndex], " ")
	email := args[emailIndex]
	// check if id flag is present
	if len(args) > emailIndex+1 {
		// last arg is id
		id = args[emailIndex+1]
		if !utils.IsInnerCircle(guildMember.Roles) {
			go service.SendDisappearingMessage(m.ChannelID, "You do not have permission to verify someone else!", 5*time.Second)
			return
		}
	}
	member, err := s.GuildMember(m.GuildID, id)
	if err != nil {
		utils.SugarLogger.Errorln(err)
		go service.SendDisappearingMessage(m.ChannelID, "Unexpected error occurred, please try again later!", 5*time.Second)
		return
	}
	// finally create user
	service.CreateUser(model.User{
		ID:        id,
		Username:  member.User.Username,
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		AvatarURL: member.User.AvatarURL("256"),
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
}
