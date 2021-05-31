package api

// CommandOptionType specifies the type of the arguments used in Command.Options
type CommandOptionType int

// Constants for each slash command option type
const (
	CommandOptionTypeSubCommand CommandOptionType = iota + 1
	CommandOptionTypeSubCommandGroup
	CommandOptionTypeString
	CommandOptionTypeInteger
	CommandOptionTypeBoolean
	CommandOptionTypeUser
	CommandOptionTypeChannel
	CommandOptionTypeRole
	CommandOptionTypeMentionable
)

func NewCommandOption(optionType CommandOptionType, name string, description string, options ...*CommandOption) *CommandOption {
	return &CommandOption{
		Type:        optionType,
		Name:        name,
		Description: description,
		Options:     options,
	}
}

func NewSubCommand(name string, description string, options ...*CommandOption) *CommandOption {
	return NewCommandOption(CommandOptionTypeSubCommand, name, description, options...)
}

func NewSubCommandGroup(name string, description string, options ...*CommandOption) *CommandOption {
	return NewCommandOption(CommandOptionTypeSubCommandGroup, name, description, options...)
}

func NewStringOption(name string, description string, options ...*CommandOption) *CommandOption {
	return NewCommandOption(CommandOptionTypeString, name, description, options...)
}

func NewIntegerOption(name string, description string, options ...*CommandOption) *CommandOption {
	return NewCommandOption(CommandOptionTypeInteger, name, description, options...)
}

func NewBooleanOption(name string, description string, options ...*CommandOption) *CommandOption {
	return NewCommandOption(CommandOptionTypeBoolean, name, description, options...)
}

func NewUserOption(name string, description string, options ...*CommandOption) *CommandOption {
	return NewCommandOption(CommandOptionTypeUser, name, description, options...)
}

func NewChannelOption(name string, description string, options ...*CommandOption) *CommandOption {
	return NewCommandOption(CommandOptionTypeChannel, name, description, options...)
}

func NewMentionableOption(name string, description string, options ...*CommandOption) *CommandOption {
	return NewCommandOption(CommandOptionTypeMentionable, name, description, options...)
}

// CommandOption are the arguments used in Command.Options
type CommandOption struct {
	Type        CommandOptionType `json:"type"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Required    bool              `json:"required,omitempty"`
	Choices     []*OptionChoice   `json:"choices,omitempty"`
	Options     []*CommandOption  `json:"options,omitempty"`
}

func (o *CommandOption) AddChoice(name string, value interface{}) *CommandOption {
	o.Choices = append(o.Choices, &OptionChoice{
		Name:  name,
		Value: value,
	})
	return o
}

func (o *CommandOption) AddOptions(options ...*CommandOption) *CommandOption {
	o.Options = append(o.Options, options...)
	return o
}

func (o *CommandOption) SetRequired(required bool) *CommandOption {
	o.Required = required
	return o
}

// OptionChoice contains the data for a user using your command
type OptionChoice struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}