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
	"slices"
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
		} else if strings.Contains(strings.ToLower(r.Name), "bot") {
			utils.SugarLogger.Infof("Found Bot Role: %s", r.ID)
			config.BotRoleID = r.ID
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

func SetDiscordRolesForUser(userID string, roleIds []string) {
	guildMember, err := Discord.GuildMember(config.DiscordGuild, userID)
	if err != nil {
		utils.SugarLogger.Errorln("Error getting guild member, ", err)
		return
	}
	existingRoles := guildMember.Roles
	rolesToAdd := []string{}
	rolesToRemove := []string{}
	for _, id := range roleIds {
		if !contains(existingRoles, id) {
			rolesToAdd = append(rolesToAdd, id)
		}
	}
	for _, id := range existingRoles {
		if !contains(roleIds, id) {
			rolesToRemove = append(rolesToRemove, id)
		}
	}
	utils.SugarLogger.Infof("Adding roles %v, removing roles %v to user %s", rolesToAdd, rolesToRemove, userID)
	for _, id := range rolesToAdd {
		err := Discord.GuildMemberRoleAdd(config.DiscordGuild, userID, id)
		if err != nil {
			utils.SugarLogger.Errorln("Error adding role, ", err)
		}
	}
	for _, id := range rolesToRemove {
		err := Discord.GuildMemberRoleRemove(config.DiscordGuild, userID, id)
		if err != nil {
			utils.SugarLogger.Errorln("Error removing role, ", err)
		}
	}
}

func SetDiscordNicknameForAllUsers() {
	users := GetAllUsers()
	for _, user := range users {
		SetDiscordNicknameForUser(user.ID)
	}
}

func SetDiscordNicknameForUser(userID string) {
	user := GetUserByID(userID)
	if user.ID == "" {
		utils.SugarLogger.Errorln("User not found")
		return
	}
	nickname := user.FirstName
	err := Discord.GuildMemberNickname(config.DiscordGuild, userID, nickname)
	if err != nil {
		utils.SugarLogger.Errorln("Error setting nickname, ", err)
	}
	utils.SugarLogger.Infof("Set nickname for user %s to %s", userID, nickname)
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

func ExchangeCodeForToken(code string) (*model.DiscordAccessTokenResponse, error) {
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

	var accessToken model.DiscordAccessTokenResponse
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
	message := fmt.Sprintf("Welcome to Gaucho Racing, %s! We're super excited to have you on board.\n\nPlease take a moment to complete your Sentinel profile at https://sso.gauchoracing.com/users/%s/edit. This is where you will be able to access all our internal tools and resources. The first time you login to Sentinel you will need to use your Discord account. Once you're in you can then set a password to be able to login with email/password in the future. You should have been added to our shared drive already and you can login to the wiki with your Sentinel account.\n\nHere are some important links:\n**Website:** <https://gauchoracing.com>\n**Wiki:** <https://wiki.gauchoracing.com>\n**GitHub:** <https://github.com/gaucho-racing>\n**Google Drive:** <https://drive.gauchoracing.com>\n\nIf you have any questions, feel free to ask in <#756738476887638111> or DM an officer or lead.", user.FirstName, user.ID)
	SendDirectMessage(userID, message)
}

func FindAllNonVerifiedUsers() {
	members, err := Discord.GuildMembers(config.DiscordGuild, "", 1000)
	if err != nil {
		utils.SugarLogger.Errorln(err.Error())
	}
	sendIds := []string{}
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
			sendIds = append(sendIds, member.User.ID)
		}
		for _, role := range member.Roles {
			if role == config.MemberRoleID {
				memberMembers++
			}
		}
		guildMembers++
	}
	for _, id := range sendIds {
		SendDirectMessage(id, "Hey there Gaucho Racer! It look's like you haven't verified your account yet. Please use the `!verify` command to verify your account before September 7th to avoid any disruption to your server access.  You can run this command in any channel in the Gaucho Racing discord server!\n\nHere's the command usage: `!verify <first name> <last name> <email>`\nAnd here's an example: `!verify Bharat Kathi bkathi@ucsb.edu`")
	}
	utils.SugarLogger.Infof("Total Members: %d", guildMembers)
	utils.SugarLogger.Infof("Members Role: %d", memberMembers)
	utils.SugarLogger.Infof("Verified Members: %d", verifiedMembers)
}

// PopulateDiscordMembers populates the discord roles for all users in the sentinel database
// Can be used for disaster recovery if all user roles are removed from the discord server
func PopulateDiscordMembers() {
	users := GetAllUsers()
	for _, user := range users {
		utils.SugarLogger.Infof("Populating discord member for user %s %s (%s)", user.FirstName, user.LastName, user.Email)
		member, err := Discord.GuildMember(config.DiscordGuild, user.ID)
		if err != nil {
			utils.SugarLogger.Errorf("Error getting discord member for user %s: %s", user.ID, err.Error())
		}
		if member != nil {
			utils.SugarLogger.Infof("Found user in discord")
			utils.SugarLogger.Infof("User has roles: %s", user.Roles)
			err := Discord.GuildMemberRoleAdd(config.DiscordGuild, user.ID, config.MemberRoleID)
			if err != nil {
				utils.SugarLogger.Errorf("Error adding role to user %s: %s", user.Email, err.Error())
			}
			if user.HasRole("d_alumni") {
				err := Discord.GuildMemberRoleAdd(config.DiscordGuild, user.ID, config.AlumniRoleID)
				if err != nil {
					utils.SugarLogger.Errorf("Error adding role to user %s: %s", user.Email, err.Error())
				}
			} else if user.HasRole("d_officer") {
				err := Discord.GuildMemberRoleAdd(config.DiscordGuild, user.ID, config.OfficerRoleID)
				if err != nil {
					utils.SugarLogger.Errorf("Error adding role to user %s: %s", user.Email, err.Error())
				}
			} else if user.HasRole("d_lead") {
				err := Discord.GuildMemberRoleAdd(config.DiscordGuild, user.ID, config.LeadRoleID)
				if err != nil {
					utils.SugarLogger.Errorf("Error adding role to user %s: %s", user.Email, err.Error())
				}
			} else if user.HasRole("d_admin") {
				err := Discord.GuildMemberRoleAdd(config.DiscordGuild, user.ID, config.AdminRoleID)
				if err != nil {
					utils.SugarLogger.Errorf("Error adding role to user %s: %s", user.Email, err.Error())
				}
			}
			utils.SugarLogger.Infof("Added main roles to user %s", user.Email)
			for _, subteam := range user.Subteams {
				err := Discord.GuildMemberRoleAdd(config.DiscordGuild, user.ID, subteam.ID)
				if err != nil {
					utils.SugarLogger.Errorf("Error adding role to user %s: %s", user.Email, err.Error())
				}
			}
			utils.SugarLogger.Infof("Added subteam roles to user %s", user.Email)
		} else {
			utils.SugarLogger.Infof("User not found in discord: %s", user.ID)
		}
	}
}

// CleanDiscordMembers does the following:
//  1. Remove all roles from users who are in the discord server but not in the sentinel database
//  2. Remove all roles from users who no longer have the member or alumni role in the sentinel database
//  3. Remove all sentinel roles from users who are no longer a member of the discord server
//
// NOTE: This will NOT kick anyone from the discord server nor DELETE any users from the sentinel database
func CleanDiscordMembers() {
	members, err := Discord.GuildMembers(config.DiscordGuild, "", 1000)
	if err != nil {
		utils.SugarLogger.Errorln(err.Error())
	}
	for _, member := range members {
		// Check if member is a bot
		if member.User.Bot {
			utils.SugarLogger.Infof("Discord user %s (%s) is a bot, skipping", member.User.ID, member.Nick)
			SendMessage(config.DiscordLogChannel, fmt.Sprintf("Discord user %s (%s) is a bot, skipping", member.User.ID, member.Nick))
			// make sure they have the bot role
			err := Discord.GuildMemberRoleAdd(config.DiscordGuild, member.User.ID, config.BotRoleID)
			if err != nil {
				utils.SugarLogger.Errorf("Error adding bot role to user %s: %s", member.User.ID, err.Error())
				SendMessage(config.DiscordLogChannel, fmt.Sprintf("Error adding bot role to user %s: %s", member.User.ID, err.Error()))
			}
			continue
		}
		user := GetUserByID(member.User.ID)
		if user.ID == "" {
			// User is in the discord server but not in the sentinel database
			// Remove all roles from user
			if len(member.Roles) > 0 {
				utils.SugarLogger.Infof("Discord user not found in Sentinel: %s (%s)", member.User.ID, member.Nick)
				SendMessage(config.DiscordLogChannel, fmt.Sprintf("Discord user not found in Sentinel: %s (%s)", member.User.ID, member.Nick))
				for _, role := range member.Roles {
					err := Discord.GuildMemberRoleRemove(config.DiscordGuild, member.User.ID, role)
					if err != nil {
						utils.SugarLogger.Errorf("Error removing role %s from user %s (%s): %s", role, member.User.ID, member.Nick, err.Error())
						SendMessage(config.DiscordLogChannel, fmt.Sprintf("Error removing role %s from user %s (%s): %s", role, member.User.ID, member.Nick, err.Error()))
					}
				}
				utils.SugarLogger.Infof("Removed all roles from user %s (%s) as they are not in the sentinel database", member.User.ID, member.Nick)
				SendMessage(config.DiscordLogChannel, fmt.Sprintf("Removed all roles from user %s (%s) as they are not in the sentinel database", member.User.ID, member.Nick))
			}
		} else if !(user.IsMember() || user.IsAlumni()) {
			// User is in the sentinel database but not a member or alumni
			// Remove all roles from user
			if len(member.Roles) > 0 {
				if slices.Contains(member.Roles, config.MemberRoleID) || slices.Contains(member.Roles, config.AdminRoleID) {
					// User is actually a member or alumni, looks like we hit an inconsistency between discord and sentinel roles (bruh edge case)
					utils.SugarLogger.Infof("Discord user has roles that are not in Sentinel: %s (%s), Discord roles: %v, Sentinel roles: %v", member.User.ID, member.Nick, member.Roles, user.Roles)
					SendMessage(config.DiscordLogChannel, fmt.Sprintf("Discord user has roles that are not in Sentinel: %s (%s), Discord roles: %v, Sentinel roles: %v", member.User.ID, member.Nick, member.Roles, user.Roles))
					// trigger a sync of roles for this user
					SyncDiscordRolesForUser(user.ID, member.Roles)
					continue
				}
				utils.SugarLogger.Infof("Discord user not a member or alumni: %s (%s)", member.User.ID, member.Nick)
				SendMessage(config.DiscordLogChannel, fmt.Sprintf("Discord user not a member or alumni: %s (%s)", member.User.ID, member.Nick))
				for _, role := range member.Roles {
					err := Discord.GuildMemberRoleRemove(config.DiscordGuild, member.User.ID, role)
					if err != nil {
						utils.SugarLogger.Errorf("Error removing role %s from user %s (%s): %s", role, member.User.ID, member.Nick, err.Error())
						SendMessage(config.DiscordLogChannel, fmt.Sprintf("Error removing role %s from user %s (%s): %s", role, member.User.ID, member.Nick, err.Error()))
					}
				}
				utils.SugarLogger.Infof("Removed all roles from user %s (%s) as they are not a member or alumni", member.User.ID, member.Nick)
				SendMessage(config.DiscordLogChannel, fmt.Sprintf("Removed all roles from user %s (%s) as they are not a member or alumni", member.User.ID, member.Nick))
			}
		}
	}
	for _, user := range GetAllUsers() {
		member, err := Discord.GuildMember(config.DiscordGuild, user.ID)
		if err != nil {
			utils.SugarLogger.Errorf("Error getting discord member for user %s: %s", user.ID, err.Error())
			continue
		}
		if member == nil {
			// User is in the sentinel database but no longer in the discord server
			// Delete user roles from sentinel (except if alumni), other jobs will take care of the rest
			if len(user.Roles) == 1 && user.IsAlumni() {
				utils.SugarLogger.Infof("User %s is an alumni and has only the alumni role in Sentinel, skipping", user.ID)
				SendMessage(config.DiscordLogChannel, fmt.Sprintf("User %s is an alumni and has only the alumni role in Sentinel, skipping", user.ID))
				continue
			}
			if len(user.Roles) > 0 {
				utils.SugarLogger.Infof("Removing sentinel roles from user %s as they are no longer in the discord server", user.ID)
				SendMessage(config.DiscordLogChannel, fmt.Sprintf("Removing sentinel roles from user %s as they are no longer in the discord server", user.ID))
				roles := []string{}
				if user.IsAlumni() {
					roles = append(roles, "d_alumni")
				}
				SetRolesForUser(user.ID, roles)
				SendMessage(config.DiscordLogChannel, fmt.Sprintf("Updated sentinel roles for user %s (%s) as they are no longer in the discord server: %v", user.ID, fmt.Sprintf("%s %s", user.FirstName, user.LastName), roles))
			}
		}
	}
}
