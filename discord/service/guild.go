package service

import (
	"sort"

	"github.com/bwmarrin/discordgo"
	"github.com/gaucho-racing/sentinel/discord/config"
)

func GetGuildRoles() ([]*discordgo.Role, error) {
	roles, err := Discord.GuildRoles(config.DiscordGuild)
	if err != nil {
		return nil, err
	}
	sort.SliceStable(roles, func(i, j int) bool {
		return roles[i].Position < roles[j].Position
	})
	return roles, nil
}

func GetGuildChannels() ([]*discordgo.Channel, error) {
	channels, err := Discord.GuildChannels(config.DiscordGuild)
	if err != nil {
		return nil, err
	}
	sort.SliceStable(channels, func(i, j int) bool {
		return channels[i].Position < channels[j].Position
	})
	return channels, nil
}
