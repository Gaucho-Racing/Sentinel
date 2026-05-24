package commands

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gaucho-racing/sentinel/discord/config"
	"github.com/gaucho-racing/sentinel/discord/pkg/logger"
	"github.com/gaucho-racing/sentinel/discord/service"
)

const verifyReplyTTL = 5 * time.Second

func Verify(args []string, s *discordgo.Session, m *discordgo.MessageCreate) {
	defer s.ChannelMessageDelete(m.ChannelID, m.ID)

	if entityID := service.GetEntityIDForDiscordUser(m.Author.ID); entityID != "" {
		logger.SugarLogger.Infof("Discord user %s is already onboarded as %s", m.Author.ID, entityID)
		dm, err := service.SendDirectMessage(m.Author.ID, fmt.Sprintf("You're already onboarded! Sign in at %s/auth/login", config.WebBaseURL))
		if err != nil {
			logger.SugarLogger.Errorf("Failed to DM onboarded user %s: %v", m.Author.ID, err)
			service.SendDisappearingMessage(m.ChannelID, fmt.Sprintf("<@%s> I couldn't DM you — enable DMs from server members and try again.", m.Author.ID), verifyReplyTTL)
			return
		}
		service.SendDisappearingMessage(m.ChannelID, fmt.Sprintf("<@%s> you're already onboarded — [check your DM](%s) for the login link.", m.Author.ID, service.DMJumpURL(dm.ChannelID, dm.ID)), verifyReplyTTL)
		return
	}

	token, err := service.CreateOnboardingTokenForDiscordUser(
		m.Author.ID,
		m.Author.Username,
		m.Author.GlobalName,
		m.Author.AvatarURL(""),
	)
	if err != nil {
		logger.SugarLogger.Errorf("Failed to mint onboarding token for %s: %v", m.Author.ID, err)
		service.SendDisappearingMessage(m.ChannelID, fmt.Sprintf("<@%s> something went wrong, try again in a minute.", m.Author.ID), verifyReplyTTL)
		return
	}

	link := fmt.Sprintf("%s/onboard?token=%s", config.WebBaseURL, token.ID)
	body := fmt.Sprintf("Welcome to Gaucho Racing! Click here to set up your Sentinel account:\n%s\n\nThis link expires in %s.", link, config.OnboardingTokenTTL)
	dm, err := service.SendDirectMessage(m.Author.ID, body)
	if err != nil {
		logger.SugarLogger.Errorf("Failed to DM onboarding link to %s: %v", m.Author.ID, err)
		service.SendDisappearingMessage(m.ChannelID, fmt.Sprintf("<@%s> I couldn't DM you — enable DMs from server members and run `%sverify` again.", m.Author.ID, config.DiscordPrefix), verifyReplyTTL)
		return
	}

	logger.SugarLogger.Infof("Issued onboarding token %s to Discord user %s", token.ID, m.Author.ID)
	service.SendDisappearingMessage(m.ChannelID, fmt.Sprintf("<@%s> 📬 [Check your DM](%s) for the verification link.", m.Author.ID, service.DMJumpURL(dm.ChannelID, dm.ID)), verifyReplyTTL)
}
