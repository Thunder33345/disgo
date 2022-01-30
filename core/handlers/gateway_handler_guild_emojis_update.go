package handlers

import (
	"github.com/DisgoOrg/disgo/core"
	"github.com/DisgoOrg/disgo/core/events"
	"github.com/DisgoOrg/disgo/discord"
	"github.com/DisgoOrg/snowflake"
)

// gatewayHandlerGuildEmojisUpdate handles discord.GatewayEventTypeGuildEmojisUpdate
type gatewayHandlerGuildEmojisUpdate struct{}

// EventType returns the discord.GatewayEventType
func (h *gatewayHandlerGuildEmojisUpdate) EventType() discord.GatewayEventType {
	return discord.GatewayEventTypeGuildEmojisUpdate
}

// New constructs a new payload receiver for the raw gateway event
func (h *gatewayHandlerGuildEmojisUpdate) New() interface{} {
	return &discord.GuildEmojisUpdateGatewayEvent{}
}

// HandleGatewayEvent handles the specific raw gateway event
func (h *gatewayHandlerGuildEmojisUpdate) HandleGatewayEvent(bot *core.Bot, sequenceNumber int, v interface{}) {
	payload := *v.(*discord.GuildEmojisUpdateGatewayEvent)

	if bot.Caches.Config().CacheFlags.Missing(core.CacheFlagEmojis) {
		return
	}

	var (
		emojiCache    = bot.Caches.Emojis().GuildCache(payload.GuildID)
		oldEmojis     = map[snowflake.Snowflake]*core.Emoji{}
		newEmojis     = map[snowflake.Snowflake]*core.Emoji{}
		updatedEmojis = map[snowflake.Snowflake]*core.Emoji{}
	)

	oldEmojis = make(map[snowflake.Snowflake]*core.Emoji, len(emojiCache))
	for key, value := range emojiCache {
		va := *value
		oldEmojis[key] = &va
	}

	for _, current := range payload.Emojis {
		emoji, ok := emojiCache[current.ID]
		if ok {
			delete(oldEmojis, current.ID)
			if !compareEmoji(emoji.Emoji, current) {
				updatedEmojis[current.ID] = bot.EntityBuilder.CreateEmoji(payload.GuildID, current, core.CacheStrategyYes)
			}
		} else {
			newEmojis[current.ID] = bot.EntityBuilder.CreateEmoji(payload.GuildID, current, core.CacheStrategyYes)
		}
	}

	for emojiID := range oldEmojis {
		bot.Caches.Emojis().Remove(payload.GuildID, emojiID)
	}

	for _, emoji := range newEmojis {
		bot.EventManager.Dispatch(&events.EmojiCreateEvent{
			GenericEmojiEvent: &events.GenericEmojiEvent{
				GenericEvent: events.NewGenericEvent(bot, sequenceNumber),
				GuildID:      payload.GuildID,
				Emoji:        emoji,
			},
		})
	}

	for _, emoji := range updatedEmojis {
		bot.EventManager.Dispatch(&events.EmojiUpdateEvent{
			GenericEmojiEvent: &events.GenericEmojiEvent{
				GenericEvent: events.NewGenericEvent(bot, sequenceNumber),
				GuildID:      payload.GuildID,
				Emoji:        emoji,
			},
		})
	}

	for _, emoji := range oldEmojis {
		bot.EventManager.Dispatch(&events.EmojiDeleteEvent{
			GenericEmojiEvent: &events.GenericEmojiEvent{
				GenericEvent: events.NewGenericEvent(bot, sequenceNumber),
				GuildID:      payload.GuildID,
				Emoji:        emoji,
			},
		})
	}

}

func compareEmoji(emoji discord.Emoji, emoji2 discord.Emoji) bool {
	return emoji.ID == emoji2.ID &&
		emoji.Name == emoji2.Name &&
		compareSnowflakeSlice(emoji.Roles, emoji2.Roles) &&
		compareUser(emoji.Creator, emoji2.Creator) &&
		emoji.RequireColons == emoji2.RequireColons &&
		emoji.Managed == emoji2.Managed &&
		emoji.Animated == emoji2.Animated &&
		emoji.Available == emoji2.Available &&
		emoji.GuildID == emoji2.GuildID
}

func compareUser(creator *discord.User, creator2 *discord.User) bool {
	if creator == nil && creator2 == nil {
		return true
	}
	if creator == nil || creator2 == nil {
		return false
	}
	return creator.ID == creator2.ID
}

func compareSnowflakeSlice(slice1 []snowflake.Snowflake, slice2 []snowflake.Snowflake) bool {
	if len(slice1) != len(slice2) {
		return false
	}
	for i := range slice1 {
		if slice1[i] != slice2[i] {
			return false
		}
	}
	return true
}
