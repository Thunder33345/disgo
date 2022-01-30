package handlers

import (
	"github.com/DisgoOrg/disgo/core"
	"github.com/DisgoOrg/disgo/core/events"
	"github.com/DisgoOrg/disgo/discord"
)

// gatewayHandlerGuildDelete handles discord.GatewayEventTypePresenceUpdate
type gatewayHandlerPresenceUpdate struct{}

// EventType returns the discord.GatewayEventType
func (h *gatewayHandlerPresenceUpdate) EventType() discord.GatewayEventType {
	return discord.GatewayEventTypePresenceUpdate
}

// New constructs a new payload receiver for the raw gateway event
func (h *gatewayHandlerPresenceUpdate) New() interface{} {
	return &discord.Presence{}
}

// HandleGatewayEvent handles the specific raw gateway event
func (h *gatewayHandlerPresenceUpdate) HandleGatewayEvent(bot *core.Bot, sequenceNumber int, v interface{}) {
	payload := *v.(*discord.Presence)

	oldPresence := bot.Caches.Presences().GetCopy(payload.GuildID, payload.PresenceUser.ID)

	_ = bot.EntityBuilder.CreatePresence(payload, core.CacheStrategyYes)

	genericEvent := events.NewGenericEvent(bot, sequenceNumber)

	var (
		oldStatus       discord.OnlineStatus
		oldClientStatus *discord.ClientStatus
		oldActivities   []discord.Activity
	)

	if oldPresence != nil {
		oldStatus = oldPresence.Status
		oldClientStatus = &oldPresence.ClientStatus
		oldActivities = oldPresence.Activities
	}

	if oldStatus != payload.Status {
		bot.EventManager.Dispatch(&events.UserStatusUpdateEvent{
			GenericEvent: genericEvent,
			UserID:       payload.PresenceUser.ID,
			OldStatus:    oldStatus,
			Status:       payload.Status,
		})
	}

	if oldClientStatus == nil || oldClientStatus.Desktop != payload.ClientStatus.Desktop || oldClientStatus.Mobile != payload.ClientStatus.Mobile || oldClientStatus.Web != payload.ClientStatus.Web {
		bot.EventManager.Dispatch(&events.UserClientStatusUpdateEvent{
			GenericEvent:    genericEvent,
			UserID:          payload.PresenceUser.ID,
			OldClientStatus: oldClientStatus,
			ClientStatus:    payload.ClientStatus,
		})
	}

	genericUserActivityEvent := events.GenericUserActivityEvent{
		GenericEvent: genericEvent,
		UserID:       payload.PresenceUser.ID,
		GuildID:      payload.GuildID,
	}

	for _, oldActivity := range oldActivities {
		var found bool
		for _, newActivity := range payload.Activities {
			if oldActivity.ID == newActivity.ID {
				found = true
				break
			}
		}
		if !found {
			genericUserActivityEvent.Activity = oldActivity
			bot.EventManager.Dispatch(&events.UserActivityStopEvent{
				GenericUserActivityEvent: &genericUserActivityEvent,
			})
		}
	}

	for _, newActivity := range payload.Activities {
		var found bool
		for _, oldActivity := range oldActivities {
			if newActivity.ID == oldActivity.ID {
				found = true
				break
			}
		}
		if !found {
			genericUserActivityEvent.Activity = newActivity
			bot.EventManager.Dispatch(&events.UserActivityStartEvent{
				GenericUserActivityEvent: &genericUserActivityEvent,
			})
		}
	}

	for _, newActivity := range payload.Activities {
		var oldActivity *discord.Activity
		for _, activity := range oldActivities {
			if newActivity.ID == activity.ID {
				oldActivity = &activity
				break
			}
		}
		if oldActivity != nil && !compareActivity(*oldActivity, newActivity) {
			genericUserActivityEvent.Activity = newActivity
			bot.EventManager.Dispatch(&events.UserActivityUpdateEvent{
				GenericUserActivityEvent: &genericUserActivityEvent,
				OldActivity:              *oldActivity,
			})
		}
	}
}

func compareActivity(a1 discord.Activity, a2 discord.Activity) bool {
	return a1.ID == a2.ID &&
		a1.Name == a2.Name &&
		a1.Type == a2.Type &&
		compareString(a1.URL, a2.URL) &&
		compareActivityTimestamps(a1.Timestamps, a2.Timestamps) &&
		a1.ApplicationID == a2.ApplicationID &&
		compareString(a1.Details, a2.Details) &&
		compareString(a1.State, a2.State) &&
		compareActivityEmoji(a1.Emoji, a2.Emoji) &&
		compareActivityParty(a1.Party, a2.Party) &&
		compareActivityAssets(a1.Assets, a2.Assets) &&
		compareActivitySecrets(a1.Secrets, a2.Secrets) &&
		compareBool(a1.Instance, a2.Instance) &&
		a1.Flags == a2.Flags &&
		compareStringSlice(a1.Buttons, a2.Buttons)
}

func compareActivityTimestamps(t1 *discord.ActivityTimestamps, t2 *discord.ActivityTimestamps) bool {
	if t1 == nil && t2 == nil {
		return true
	}

	if t1 == nil || t2 == nil {
		return false
	}

	return t1.Start == t2.Start &&
		t1.End == t2.End
}

func compareActivityEmoji(e1 *discord.ActivityEmoji, e2 *discord.ActivityEmoji) bool {
	if e1 == nil && e2 == nil {
		return true
	}

	if e1 == nil || e2 == nil {
		return false
	}

	return e1.Name == e2.Name &&
		compareSnowflake(e1.ID, e2.ID) &&
		compareBool(e1.Animated, e2.Animated)
}

func compareActivityParty(p1 *discord.ActivityParty, p2 *discord.ActivityParty) bool {
	if p1 == nil && p2 == nil {
		return true
	}

	if p1 == nil || p2 == nil {
		return false
	}

	return p1.ID == p2.ID &&
		p1.Size[0] == p2.Size[0] &&
		p1.Size[1] == p2.Size[1]

}

func compareActivityAssets(a1 *discord.ActivityAssets, a2 *discord.ActivityAssets) bool {
	if a1 == nil && a2 == nil {
		return true
	}

	if a1 == nil || a2 == nil {
		return false
	}

	return a1.LargeImage == a2.LargeImage &&
		a1.LargeText == a2.LargeText &&
		a1.SmallImage == a2.SmallImage &&
		a1.SmallText == a2.SmallText
}

func compareActivitySecrets(s1 *discord.ActivitySecrets, s2 *discord.ActivitySecrets) bool {
	if s1 == nil && s2 == nil {
		return true
	}

	if s1 == nil || s2 == nil {
		return false
	}

	return s1.Join == s2.Join &&
		s1.Spectate == s2.Spectate &&
		s1.Match == s2.Match
}

func compareStringSlice(a1 []string, a2 []string) bool {
	if a1 == nil && a2 == nil {
		return true
	}

	if a1 == nil || a2 == nil {
		return false
	}

	if len(a1) != len(a2) {
		return false
	}

	for i := range a1 {
		if a1[i] != a2[i] {
			return false
		}
	}

	return true
}
