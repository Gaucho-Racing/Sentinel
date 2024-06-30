package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

func InitializeRoles() {
	g, err := Discord.Guild(config.DiscordGuild)
	if err != nil {
		utils.SugarLogger.Errorln("Error getting guild,", err)
		return
	}
	for _, r := range g.Roles {
		if strings.Contains(strings.ToLower(r.Name), "member") {
			utils.SugarLogger.Infof("Found Member Role: %s", r.ID)
			config.MemberRoleID = r.ID
		} else if strings.Contains(strings.ToLower(r.Name), "alumnus") {
			utils.SugarLogger.Infof("Found Alumni Role: %s", r.ID)
			config.AlumniRoleID = r.ID
		} else if strings.Contains(strings.ToLower(r.Name), "admin") {
			utils.SugarLogger.Infof("Found Admin Role: %s", r.ID)
			config.AdminRoleID = r.ID
		} else if strings.Contains(strings.ToLower(r.Name), "officer") {
			utils.SugarLogger.Infof("Found Officer Role: %s", r.ID)
			config.OfficerRoleID = r.ID
		} else if strings.Contains(strings.ToLower(r.Name), "lead") {
			utils.SugarLogger.Infof("Found Lead Role: %s", r.ID)
			config.LeadRoleID = r.ID
		}
	}
}

func SyncRolesForAllUsers() {
	members, err := Discord.GuildMembers(config.DiscordGuild, "", 1000)
	if err != nil {
		utils.SugarLogger.Errorln(err.Error())
	}
	count := 0
	for _, member := range members {
		user := GetUserByID(member.User.ID)
		if user.ID != "" {
			SyncDiscordRolesForUser(user.ID, member.Roles)
			count++
		}
	}
	utils.SugarLogger.Infof("Synced roles for %d users", count)
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

func SendDirectMessage(userID string, message string) {
	channel, err := Discord.UserChannelCreate(userID)
	if err != nil {
		utils.SugarLogger.Errorln(err.Error())
	}
	_, err = Discord.ChannelMessageSend(channel.ID, message)
	if err != nil {
		utils.SugarLogger.Errorln(err.Error())
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
	lastActivity := GetLastActivityForUser(user.ID)
	if lastActivity.ID != "" {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Last Activity",
			Value:  lastActivity.Action + " on " + lastActivity.CreatedAt.Format("Jan 2, 2006 3:04 PM"),
			Inline: false,
		})
	}
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

func ExchangeCodeForToken(code string) (*model.AccessTokenResponse, error) {
	tokenURL := "https://discord.com/api/oauth2/token"

	data := url.Values{}
	data.Set("client_id", config.DiscordClientID)
	data.Set("client_secret", config.DiscordClientSecret)
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", config.DiscordRedirectURI)

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		utils.SugarLogger.Errorln("error exchanging code for token: ", string(body))
		return nil, fmt.Errorf("error exchanging code for token")
	}

	var accessToken model.AccessTokenResponse
	err = json.Unmarshal(body, &accessToken)
	if err != nil {
		return nil, err
	}

	return &accessToken, nil
}

func GetDiscordUserFromToken(accessToken string) (*model.DiscordUser, error) {
	userURL := "https://discord.com/api/users/@me"

	req, err := http.NewRequest("GET", userURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		utils.SugarLogger.Errorln("error getting user from token: ", string(body))
		return nil, fmt.Errorf("error getting user from token: %s", string(body))
	}

	var user model.DiscordUser
	err = json.Unmarshal(body, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func SendUserWelcomeMessage(userID string) {
	user := GetUserByID(userID)
	if user.ID == "" {
		utils.SugarLogger.Errorln("User not found")
		return
	}
	message := fmt.Sprintf("Welcome to Gaucho Racing, %s! We're super excited to have you on board.\n\nPlease take a moment to complete your Sentinel profile at https://sso.gauchoracing.com. This is where you will be able to access all our internal tools and resources.\n\nHere are some important links:\n**Website:** <https://gauchoracing.com>\n**GitHub:** <https://github.com/gaucho-racing>\n**Google Drive:** <https://drive.gauchoracing.com>\n**Wiki:** <https://wiki.gauchoracing.com>\n\nIf you have any questions, feel free to ask in <#756738476887638111> or DM an officer or lead.", user.FirstName)
	SendDirectMessage(userID, message)
}

func FindAllNonVerifiedUsers() {
	members, err := Discord.GuildMembers(config.DiscordGuild, "", 1000)
	if err != nil {
		utils.SugarLogger.Errorln(err.Error())
	}
	guildMembers := 0
	memberMembers := 0
	verifiedMembers := 0
	for _, member := range members {
		user := GetUserByID(member.User.ID)
		if user.ID != "" {
			utils.SugarLogger.Infof("User found: %s", user.ID)
			verifiedMembers++
		} else {
			utils.SugarLogger.Infof("User not found: %s", member.User.ID)
		}
		for _, role := range member.Roles {
			if role == config.MemberRoleID {
				memberMembers++
			}
		}
		guildMembers++
	}
	utils.SugarLogger.Infof("Total Members: %d", guildMembers)
	utils.SugarLogger.Infof("Members Role: %d", memberMembers)
	utils.SugarLogger.Infof("Verified Members: %d", verifiedMembers)
	SendDirectMessage("348220961155448833", "Hey there Gaucho Racer! It look's like you haven't verified your account yet. Please use the `!verify` command to verify your account before June 8th to avoid any disruption to your server access. You can run this command in any channel in the Gaucho Racing discord server!\n\nHere's the command usage: `!verify <first name> <last name> <email>`\nAnd here's an example: `!verify Bharat Kathi bkathi@ucsb.edu`")
}
