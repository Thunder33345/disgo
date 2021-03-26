package api

import (
	"time"
)

// The MessageType indicates the Message type
type MessageType int

// Constants for the MessageType
const (
	MessageTypeDefault MessageType = iota
	MessageTypeRecipientAdd
	MessageTypeRecipientRemove
	MessageTypeCall
	MessageTypeChannelNameChange
	MessageTypeChannelIconChange
	ChannelPinnedMessage
	MessageTypeGuildMemberJoin
	MessageTypeUserPremiumGuildSubscription
	MessageTypeUserPremiumGuildSubscriptionTier1
	MMessageTypeUserPremiumGuildSubscriptionTier2
	MessageTypeUserPremiumGuildSubscriptionTier3
	MessageTypeChannelFollowAdd
	_
	MessageTypeGuildDiscoveryDisqualified
	MessageTypeGuildDiscoveryRequalified
	_
	_
	_
	MessageTypeReply
	MessageTypeApplicationCommand
)

// The MessageFlags of a Message
type MessageFlags int64

// Add allows you to add multiple bits together, producing a new bit
func (f MessageFlags) Add(bits ...Bit) Bit {
	total := MessageFlags(0)
	for _, bit := range bits {
		total |= bit.(MessageFlags)
	}
	f |= total
	return f
}

// Remove allows you to subtract multiple bits from the first, producing a new bit
func (f MessageFlags) Remove(bits ...Bit) Bit {
	total := MessageFlags(0)
	for _, bit := range bits {
		total |= bit.(MessageFlags)
	}
	f &^= total
	return f
}

// HasAll will ensure that the bit includes all of the bits entered
func (f MessageFlags) HasAll(bits ...Bit) bool {
	for _, bit := range bits {
		if !f.Has(bit) {
			return false
		}
	}
	return true
}

// Has will check whether the Bit contains another bit
func (f MessageFlags) Has(bit Bit) bool {
	return (f & bit.(MessageFlags)) == bit
}

// MissingAny will check whether the bit is missing any one of the bits
func (f MessageFlags) MissingAny(bits ...Bit) bool {
	for _, bit := range bits {
		if !f.Has(bit) {
			return true
		}
	}
	return false
}

// Missing will do the inverse of Bit.Has
func (f MessageFlags) Missing(bit Bit) bool {
	return !f.Has(bit)
}

// Constants for MessageFlags
const (
	MessageFlagNone        MessageFlags = 0
	MessageFlagCrossposted MessageFlags = 1 << iota
	MessageFlagIsCrosspost
	MessageFlagSuppressEmbeds
	MessageFlagSourceMessageDeleted
	MessageFlagUrgent
	_
	MessageFlagEphemeral
)

// Message is a struct for messages sent in discord text-based channels
type Message struct {
	Disgo           Disgo
	ID              Snowflake     `json:"id"`
	GuildID         *Snowflake    `json:"guild_id"`
	Reactions       []Reactions   `json:"reactions"`
	Attachments     []interface{} `json:"attachments"`
	Tts             bool          `json:"tts"`
	Embeds          []*Embed      `json:"embeds,omitempty"`
	CreatedAt       time.Time     `json:"timestamp"`
	MentionEveryone bool          `json:"mention_everyone"`
	Pinned          bool          `json:"pinned"`
	EditedTimestamp interface{}   `json:"edited_timestamp"`
	Author          User          `json:"author"`
	MentionRoles    []interface{} `json:"mention_roles"`
	Content         *string       `json:"content,omitempty"`
	ChannelID       Snowflake     `json:"channel_id"`
	Mentions        []interface{} `json:"mentions"`
	MessageType     MessageType   `json:"type"`
	LastUpdated     *time.Time
}

// MessageInteraction is sent on the Message object when the message_events is a response to an interaction
type MessageInteraction struct {
	ID   Snowflake       `json:"id"`
	Type InteractionType `json:"type"`
	Name string          `json:"name"`
	User User            `json:"user"`
}

// missing Member, mention channels, nonce, webhook id, type, activity, application, message_reference, flags, stickers
// referenced_message, interaction
// https://discord.com/developers/docs/resources/channel#message-object

// Guild gets the guild_events the message_events was sent in
func (m Message) Guild() *Guild {
	if m.GuildID == nil {
		return nil
	}
	return m.Disgo.Cache().Guild(*m.GuildID)
}

// Channel gets the channel the message_events was sent in
func (m Message) Channel() *MessageChannel {
	return nil //m.Disgo.Cache().MessageChannel(m.ChannelID)
}

// AddReactionByEmote allows you to add an Emote to a message_events via reaction
func (m Message) AddReactionByEmote(emote Emote) error {
	return m.AddReaction(emote.Reaction())
}

// AddReaction allows you to add a reaction to a message_events from a string, for example a custom emoji ID, or a native
// emoji
func (m Message) AddReaction(emoji string) error {
	return m.Disgo.RestClient().AddReaction(m.ChannelID, m.ID, emoji)
}

// Reactions contains information about the reactions of a message_events
type Reactions struct {
	Count int   `json:"count"`
	Me    bool  `json:"me"`
	Emoji Emote `json:"emoji"`
}