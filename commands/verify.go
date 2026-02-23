package commands

import (
	"sentinel/config"
	"sentinel/model"
	"sentinel/service"
	"sentinel/utils"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func Verify(args []string, s *discordgo.Session, m *discordgo.MessageCreate) {
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

	if len(args) < 3 {
		go service.SendDisappearingMessage(m.ChannelID, "Command usage: `!verify <first name> <last name> <email>`", 5*time.Second)
		return
	}
	emailIndex := -1
	// find email, extract first name and last name from that
	for i, arg := range args {
		if strings.Contains(arg, "@ucsb.edu") || strings.Contains(arg, "@pipeline.sbcc.edu") || utils.IsInnerCircle(guildMember.Roles) && strings.Contains(arg, "@") {
			emailIndex = i
		}
	}
	if emailIndex == -1 {
		go service.SendDisappearingMessage(m.ChannelID, "Email must be a valid UCSB or SBCC Pipeline email", 5*time.Second)
		return
	}

	id := m.Author.ID
	firstName := args[0]
	lastName := strings.Join(args[1:emailIndex], " ")
	email := args[emailIndex]

	msg, _ := service.Discord.ChannelMessageSend(m.ChannelID, "we are checking...")
	defer service.Discord.ChannelMessageDelete(m.ChannelID, msg.ID)

	// check if id flag is present
	if len(args) > emailIndex+1 {
		// last arg is id
		id = args[emailIndex+1]
		if !utils.IsInnerCircle(guildMember.Roles) {
			go service.SendDisappearingMessage(m.ChannelID, "You do not have permission to verify someone else!", 5*time.Second)
			return
		}
	}

	// check if user is already verified
	if service.GetUserByID(id).ID != "" && service.GetUserByID(id).IsMember() {
		go service.SendDisappearingMessage(m.ChannelID, "You are already verified!", 5*time.Second)
		return
	} else if service.GetUserByID(id).ID != "" && service.GetUserByID(id).IsAlumni() {
		// special case where user was an alumni, and left the server, but is trying to re-verify
		go service.SendDisappearingMessage(m.ChannelID, "Welcome back, we're happy to see you again!", 5*time.Second)
		// assign alumni role
		err = s.GuildMemberRoleAdd(m.GuildID, id, config.AlumniRoleID)
		if err != nil {
			utils.SugarLogger.Errorln(err)
		}
		return
	} else if service.GetUserByEmail(email).ID != "" && service.GetUserByEmail(email).IsMember() {
		go service.SendDisappearingMessage(m.ChannelID, "This email is already registered!", 5*time.Second)
		return
	}

	// verify name and email
	if strings.Contains(firstName, "<") || strings.Contains(lastName, "<") || strings.Contains(email, "<") {
		go service.SendDisappearingMessage(m.ChannelID, "Don't include the < > in your name and email!", 5*time.Second)
		return
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
	}, false)

	// rename user
	err = s.GuildMemberNickname(m.GuildID, id, firstName)
	if err != nil {
		utils.SugarLogger.Errorln(err)
	}

	// sync roles
	service.SyncDiscordRolesForUser(id, member.Roles)

	// google drive access
	_ = service.RemoveMemberFromDrive(config.SharedDriveID, email)

	// assign member role (if alumni, discord handler will remove)
	err = s.GuildMemberRoleAdd(m.GuildID, id, config.MemberRoleID)
	if err != nil {
		utils.SugarLogger.Errorln(err)
	}

	go service.SendDisappearingMessage(m.ChannelID, "You have been verified! Welcome to the server <@"+id+">!", 5*time.Second)
	go service.SendUserWelcomeMessage(id)
}
