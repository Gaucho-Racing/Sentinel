package commands

import (
	"net/http"
	"sentinel/config"
	"sentinel/service"
	"sentinel/utils"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func Github(args []string, s *discordgo.Session, m *discordgo.MessageCreate) {
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

	user := service.GetUserByID(m.Author.ID)
	if user.ID == "" {
		// User not found
		go service.SendDisappearingMessage(m.ChannelID, "You must verify your account first! (`!verify <first name> <last name> <email>`)", 5*time.Second)
		return
	}
	if len(args) < 1 {
		go service.SendDisappearingMessage(m.ChannelID, "Command usage: `!github <username>`", 5*time.Second)
		return
	}
	reqBody := `{"role": "member"}`
	if utils.IsInnerCircle(guildMember.Roles) {
		reqBody = `{"role": "admin"}`
	}
	req, err := http.NewRequest("PUT", "https://api.github.com/orgs/gaucho-racing/memberships/"+args[0], strings.NewReader(reqBody))
	if err != nil {
		utils.SugarLogger.Errorln(err)
		go service.SendDisappearingMessage(m.ChannelID, "Unexpected error occurred, please try again later!", 5*time.Second)
		return
	}
	req.Header.Set("Authorization", "Bearer "+config.GithubToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		utils.SugarLogger.Errorln(err)
		go service.SendDisappearingMessage(m.ChannelID, "Unexpected error occurred, please try again later!", 5*time.Second)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		utils.SugarLogger.Errorln("Failed to add user to GitHub organization")
		go service.SendDisappearingMessage(m.ChannelID, "Failed to add user to GitHub organization", 5*time.Second)
		return
	}
	AddGithubUsernameToRoles(args[0], m.Author.ID)
	go service.SendDisappearingMessage(m.ChannelID, "Successfully invited user to GitHub organization!", 5*time.Second)
}

func AddGithubUsernameToRoles(ghUsername string, userID string) {
	roles := service.GetRolesForUser(userID)
	roles = append(roles, "github_"+ghUsername)
	service.SetRolesForUser(userID, roles)
}
