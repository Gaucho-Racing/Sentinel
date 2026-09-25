package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/gaucho-racing/sentinel/discord/config"
	"github.com/gaucho-racing/sentinel/discord/pkg/logger"
	"github.com/gaucho-racing/sentinel/discord/pkg/sentinel"
)

var unlinkedMemberSweep syncJob

func TriggerUnlinkedMemberSweep() {
	unlinkedMemberSweep.Start(func(ctx context.Context) {
		if err := sweepUnlinkedGuildMembers(ctx); err != nil && !errors.Is(err, context.Canceled) {
			logger.SugarLogger.Errorf("member sweep: %v", err)
		}
	})
}

func sweepUnlinkedGuildMembers(ctx context.Context) error {
	var auths []externalAuthRow
	if err := sentinel.Get("/api/core/entity/external/DISCORD", &auths); err != nil {
		return fmt.Errorf("list Discord identities: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	members, err := fetchAllGuildMembers(ctx)
	if err != nil {
		return fmt.Errorf("list guild members: %w", err)
	}
	linked := make(map[string]struct{}, len(auths))
	for _, auth := range auths {
		linked[auth.ExternalID] = struct{}{}
	}

	removed := 0
	for _, member := range unlinkedMembersWithRoles(members, linked) {
		if err := ctx.Err(); err != nil {
			return err
		}

		current, err := Discord.GuildMember(config.DiscordGuild, member.User.ID)
		if err != nil {
			if !isDiscordNotFound(err) {
				logger.SugarLogger.Errorf("member sweep: failed to refresh guild member %s: %v", member.User.ID, err)
			}
			continue
		}
		if current.User == nil || current.User.Bot || len(current.Roles) == 0 {
			continue
		}

		var entity entityResponse
		err = sentinel.Get("/api/core/entity/external/DISCORD/"+member.User.ID, &entity)
		if err == nil {
			if entity.ID == "" {
				logger.SugarLogger.Errorf("member sweep: identity lookup for %s returned no entity ID", member.User.ID)
			}
			continue
		}
		var apiError *sentinel.APIError
		if !errors.As(err, &apiError) || apiError.Status != http.StatusNotFound {
			logger.SugarLogger.Errorf("member sweep: failed to check Sentinel identity for %s: %v", member.User.ID, err)
			continue
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		emptyRoles := []string{}
		if _, err := Discord.GuildMemberEdit(config.DiscordGuild, member.User.ID, &discordgo.GuildMemberParams{Roles: &emptyRoles}); err != nil {
			logger.SugarLogger.Errorf("member sweep: failed to remove roles from %s: %v", member.User.ID, err)
			continue
		}
		removed++
		logger.SugarLogger.Infof("member sweep: removed roles from unlinked Discord user %s", member.User.ID)
	}
	logger.SugarLogger.Infof("member sweep: checked %d guild members, removed roles from %d unlinked members", len(members), removed)
	return nil
}

func unlinkedMembersWithRoles(members []*discordgo.Member, linked map[string]struct{}) []*discordgo.Member {
	var candidates []*discordgo.Member
	for _, member := range members {
		if member == nil || member.User == nil || member.User.Bot || len(member.Roles) == 0 {
			continue
		}
		if _, ok := linked[member.User.ID]; !ok {
			candidates = append(candidates, member)
		}
	}
	return candidates
}
