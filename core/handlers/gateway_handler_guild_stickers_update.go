package handlers

import (
	"github.com/DisgoOrg/disgo/core"
	"github.com/DisgoOrg/disgo/core/events"
	"github.com/DisgoOrg/disgo/discord"
	"github.com/DisgoOrg/snowflake"
)

// gatewayHandlerGuildStickersUpdate handles discord.GatewayEventTypeGuildStickersUpdate
type gatewayHandlerGuildStickersUpdate struct{}

// EventType returns the discord.GatewayEventType
func (h *gatewayHandlerGuildStickersUpdate) EventType() discord.GatewayEventType {
	return discord.GatewayEventTypeGuildStickersUpdate
}

// New constructs a new payload receiver for the raw gateway event
func (h *gatewayHandlerGuildStickersUpdate) New() interface{} {
	return &discord.GuildStickersUpdateGatewayEvent{}
}

// HandleGatewayEvent handles the specific raw gateway event
func (h *gatewayHandlerGuildStickersUpdate) HandleGatewayEvent(bot *core.Bot, sequenceNumber int, v interface{}) {
	payload := *v.(*discord.GuildStickersUpdateGatewayEvent)

	if bot.Caches.Config().CacheFlags.Missing(core.CacheFlagStickers) {
		return
	}

	var (
		stickerCache    = bot.Caches.Stickers().GuildCache(payload.GuildID)
		oldStickers     = map[snowflake.Snowflake]*core.Sticker{}
		newStickers     = map[snowflake.Snowflake]*core.Sticker{}
		updatedStickers = map[snowflake.Snowflake]*core.Sticker{}
	)

	oldStickers = make(map[snowflake.Snowflake]*core.Sticker, len(stickerCache))
	for key, value := range stickerCache {
		va := *value
		oldStickers[key] = &va
	}

	for _, current := range payload.Stickers {
		sticker, ok := stickerCache[current.ID]
		if ok {
			delete(oldStickers, current.ID)
			if !compareSticker(sticker.Sticker, current) {
				updatedStickers[current.ID] = bot.EntityBuilder.CreateSticker(current, core.CacheStrategyYes)
			}
		} else {
			newStickers[current.ID] = bot.EntityBuilder.CreateSticker(current, core.CacheStrategyYes)
		}
	}

	for stickerID := range oldStickers {
		bot.Caches.Stickers().Remove(payload.GuildID, stickerID)
	}

	for _, sticker := range newStickers {
		bot.EventManager.Dispatch(&events.StickerCreateEvent{
			GenericStickerEvent: &events.GenericStickerEvent{
				GenericEvent: events.NewGenericEvent(bot, sequenceNumber),
				GuildID:      payload.GuildID,
				Sticker:      sticker,
			},
		})
	}

	for _, sticker := range updatedStickers {
		bot.EventManager.Dispatch(&events.StickerUpdateEvent{
			GenericStickerEvent: &events.GenericStickerEvent{
				GenericEvent: events.NewGenericEvent(bot, sequenceNumber),
				GuildID:      payload.GuildID,
				Sticker:      sticker,
			},
		})
	}

	for _, sticker := range oldStickers {
		bot.EventManager.Dispatch(&events.StickerDeleteEvent{
			GenericStickerEvent: &events.GenericStickerEvent{
				GenericEvent: events.NewGenericEvent(bot, sequenceNumber),
				GuildID:      payload.GuildID,
				Sticker:      sticker,
			},
		})
	}

}

func compareSticker(a discord.Sticker, b discord.Sticker) bool {
	return a.ID == b.ID &&
		compareSnowflake(a.PackID, b.PackID) &&
		a.Name == b.Name &&
		a.Description == b.Description &&
		a.Tags == b.Tags &&
		a.Type == b.Type &&
		a.FormatType == b.FormatType &&
		compareBool(a.Available, b.Available) &&
		compareSnowflake(a.GuildID, b.GuildID) &&
		compareUser(a.User, b.User) &&
		compareInt(a.SortValue, b.SortValue)
}

func compareSnowflake(s1 *snowflake.Snowflake, s2 *snowflake.Snowflake) bool {
	if s1 == nil && s2 == nil {
		return true
	}

	if s1 == nil || s2 == nil {
		return false
	}

	return *s1 == *s2
}

func compareInt(i1 *int, i2 *int) bool {
	if i1 == nil && i2 == nil {
		return true
	}

	if i1 == nil || i2 == nil {
		return false
	}

	return *i1 == *i2
}

func compareBool(b1 *bool, b2 *bool) bool {
	if b1 == nil && b2 == nil {
		return true
	}

	if b1 == nil || b2 == nil {
		return false
	}

	return *b1 == *b2
}

func compareString(s1 *string, s2 *string) bool {
	if s1 == nil && s2 == nil {
		return true
	}

	if s1 == nil || s2 == nil {
		return false
	}

	return *s1 == *s2
}
