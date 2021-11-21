package events

import (
	"github.com/DisgoOrg/disgo/core"
	"github.com/DisgoOrg/disgo/discord"
)

type WebhooksUpdateEvent struct {
	*GenericEvent
	GuildId   discord.Snowflake
	ChannelID discord.Snowflake
}

func (e *WebhooksUpdateEvent) Guild() *core.Guild {
	return e.Bot().Caches.Guilds().Get(e.GuildId)
}

func (e *WebhooksUpdateEvent) Channel() core.GuildMessageChannel {
	if ch := e.Bot().Caches.Channels().Get(e.ChannelID); ch != nil {
		return ch.(core.GuildMessageChannel)
	}
	return nil
}
