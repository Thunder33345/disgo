package handlers

import (
	"github.com/DiscoOrg/disgo/api"
	"github.com/DiscoOrg/disgo/api/events"
)

type GuildCreateHandler struct{}

func (h GuildCreateHandler) New() interface{} {
	return &api.Guild{}
}

func (h GuildCreateHandler) Handle(disgo api.Disgo, eventManager api.EventManager, i interface{}) {
	guild, ok := i.(*api.Guild)
	if !ok {
		return
	}
	guild.Disgo = disgo
	oldGuild := disgo.Cache().Guild(guild.ID)
	var wasUnavailable bool
	if oldGuild == nil {
		wasUnavailable = true
	} else {
		wasUnavailable = oldGuild.Unavailable
	}

	disgo.Cache().CacheGuild(guild)
	for i := range guild.Channels {
		channel := guild.Channels[i]
		channel.Disgo = disgo
		channel.GuildID = guild.ID
		switch channel.Type {
		case api.ChannelTypeText, api.ChannelTypeNews:
			disgo.Cache().CacheTextChannel(&api.TextChannel{
				GuildChannel: *channel,
				MessageChannel: api.MessageChannel{
					Channel: channel.Channel,
				},
			})
		case api.ChannelTypeVoice:
			disgo.Cache().CacheVoiceChannel(&api.VoiceChannel{
				GuildChannel: *channel,
			})
		case api.ChannelTypeCategory:
			disgo.Cache().CacheCategory(&api.CategoryChannel{
				GuildChannel: *channel,
			})
		case api.ChannelTypeStore:
			disgo.Cache().CacheStoreChannel(&api.StoreChannel{
				GuildChannel: *channel,
			})
		}
	}

	for i := range guild.Roles {
		role := guild.Roles[i]
		role.Disgo = disgo
		role.GuildID = guild.ID
		disgo.Cache().CacheRole(role)
	}

	genericGuildEvent := events.GenericGuildEvent{
		Event:   api.Event{
			Disgo: disgo,
		},
		GuildID: guild.ID,
	}

	eventManager.Dispatch(genericGuildEvent)

	if wasUnavailable {
		eventManager.Dispatch(events.GuildAvailableEvent{
			GenericGuildEvent: genericGuildEvent,
			Guild: guild,
		})
	} else {
		// guild join
		eventManager.Dispatch(events.GuildJoinEvent{
			GenericGuildEvent: genericGuildEvent,
			Guild: guild,
		})
	}
}