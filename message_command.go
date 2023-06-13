package main

import (
	"fmt"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
)

// A function for constructing MessageCommand
// Also validate params and generate usage
func NewMessageCommand(
	name 		string, 
	description string, 
	examples 	[]string,
	ignoreCase	bool,
	params 		[]*MessageCommandParam, 
	subcommands []*MessageCommand,
	handler 	MessageCommandHandler,
) *MessageCommand {
	cmd := new(MessageCommand)

	// Initialize fields
	cmd.Name = name
	cmd.Description = description
	cmd.Examples = examples
	cmd.SubCommands = subcommands
	cmd.Handler = handler
	cmd.IgnoreCase = ignoreCase
	
	// Validate params before initializing
	// params options Optional must be the lasts
	// params options List is unique and must be the last 
	opt := 0
	for i, p := range params {
		if p.Option == MessageCommandParamOptionList && i != len(params) - 1 {
			panic("Param option List must be unique and the last param")
		}
		if p.Option == MessageCommandParamOptionOptional {
			if i - opt > 1 {
				panic("Param options Optional must be the lasts")
			} else {
				opt = 1
			}
		}
	}
	cmd.Params = params

	// Initialize Usage field
	s := "**" + name + "**"
	for _, p := range params {
		s += " `" + p.Name + "`"
	}
	cmd.Usage = s

	return cmd
}

// Event type for message
type MessageCommandEvent func(s *discordgo.Session, m *discordgo.MessageCreate)

type MessageCommandHandler func(ctx *MessageCommandContext)

type MessageCommandParamType uint8

type MessageCommandParamOption uint8

// Enum for Param Type
const (
	MessageCommandParamTypeString 			MessageCommandParamType = 1
	MessageCommandParamTypeInteger			MessageCommandParamType = 2
	MessageCommandParamTypeBoolean			MessageCommandParamType = 3
	MessageCommandParamTypeUser				MessageCommandParamType = 4
	MessageCommandParamTypeChannel			MessageCommandParamType = 5
	MessageCommandParamTypeRole				MessageCommandParamType = 6
	// MessageCommandParamTypeMentionable		MessageCommandParamType = 7
	// MessageCommandParamTypeSubCommand		MessageCommandParamType = 8
	// MessageCommandParamTypeSubCommandGroup 	MessageCommandParamType = 9
)

// Enum for Param Option
const (
	MessageCommandParamOptionRequired	MessageCommandParamOption = 1

	// Always be the last parameter
	MessageCommandParamOptionOptional	MessageCommandParamOption = 2
	MessageCommandParamOptionList		MessageCommandParamOption = 3
)

type MessageCommandParam struct {
	Name 	string
	Type 	MessageCommandParamType
	Option	MessageCommandParamOption
}

func (p *MessageCommandParam) OptionType() string {
	var option string
	switch p.Option {
	case MessageCommandParamOptionRequired:
		option = "Required[%s]"
	case MessageCommandParamOptionOptional:
		option = "Optional[%s]"
	case MessageCommandParamOptionList:
		option = "List[%s]"
	default:
		panic("There is no such Option")
	}

	var t string

	switch p.Type {
	case MessageCommandParamTypeString:
		t = "`string`"
	case MessageCommandParamTypeInteger:
		t = "`int`"
	case MessageCommandParamTypeBoolean:
		t = "`boolean`"
	case MessageCommandParamTypeUser:
		t = "`user`"
	case MessageCommandParamTypeRole:
		t = "`role`"
	case MessageCommandParamTypeChannel:
		t = "`channel`"
	default:
		panic("There is no such Type")
	}

	return fmt.Sprintf(option, t)
}

type MessageCommand struct {
	Name 		string
	Params 		[]*MessageCommandParam

	Aliases 	[]string
	IgnoreCase 	bool
	Description string
	Usage		string
	Examples	[]string

	SubCommands []*MessageCommand

	// Command Handler
	// 
	Handler 	MessageCommandHandler
}

func (cmd *MessageCommand) Embed() *discordgo.MessageEmbed {
	param := ""
	for _, p := range cmd.Params {
		param += fmt.Sprintf("`%s`: ", p.Name) + p.OptionType() + "\n"
	}

	f := []*discordgo.MessageEmbedField{
		{
			Name: "Description",
			Value: cmd.Description,
			Inline: false,
		},
		{
			Name: "Usage",
			Value: cmd.Usage,
			Inline: false,
		},
		{
			Name: "Argument(s)",
			Value: param,
			Inline: false,
		},
		{
			Name: "Example(s)",
			Value: strings.Join(cmd.Examples, "\n"),
		},
	}

	return &discordgo.MessageEmbed{
		Title: cmd.Name + " COMMAND",
		Fields: f,
	}
}

func (cmd *MessageCommand) ValidateArguments(arguments []string) bool {
	l := len(arguments)
	
	var minArg, maxArg int

	// Get min arguments
	for i, p := range cmd.Params {
		if p.Option == MessageCommandParamOptionOptional {
			break
		}
		minArg = i + 1
	}

	// Get max arguments
	pl := len(cmd.Params)
	if pl == 0 {
		maxArg = pl
	} else if cmd.Params[pl-1].Option == MessageCommandParamOptionList {
		maxArg = 2000
	} else {
		maxArg = pl
	}

	if l >= minArg && l <= maxArg {
		return true
	}

	return false
}

func (cmd *MessageCommand) ConvertArguments(
	s *discordgo.Session, 
	m *discordgo.Message, 
	arguments []string,
) (convertedArgs map[string]interface{}, err error) {
	paramMap := make(map[string]interface{})

	for i, p := range cmd.Params {
		if p.Option == MessageCommandParamOptionOptional || p.Option == MessageCommandParamOptionRequired {
			if res, err := argumentsConverter(s, m, arguments[i], p.Type); err != nil {
				return nil, err
			} else {
				paramMap[p.Name] = res
			}
		} else if p.Option == MessageCommandParamOptionList {
			li := make([]interface{}, len(arguments) - i)
			for j := 0; j < len(arguments) - i; j++ {
				if res, err := argumentsConverter(s, m, arguments[i + j], p.Type); err != nil {
					return nil, err
				} else {
					li[j] = res
				}
			}
			paramMap[p.Name] = li
			break
		}
	}

	return paramMap, nil
}

type MessageCommandMap struct {
	// To provide thread safe access endpoints
	sync.RWMutex
	Map 	map[string]*MessageCommand
}

func (m *MessageCommandMap) Init() {
	if m.Map == nil {
		m.Map = make(map[string]*MessageCommand)
	}
}

func (m *MessageCommandMap) Set(cmd *MessageCommand) {
	m.Lock()
	defer m.Unlock()

	m.Init()
	if cmd.IgnoreCase {
		m.Map[strings.ToLower(cmd.Name)] = cmd
		for _, alias := range cmd.Aliases {
			m.Map[strings.ToLower(alias)] = cmd
		}
	} else {
		m.Map[cmd.Name] = cmd
		for _, alias := range cmd.Aliases {
			m.Map[alias] = cmd
		}
	}
}

func (m *MessageCommandMap) Get(name string) *MessageCommand {
	m.RLock()
	defer m.RUnlock()

	m.Init()
	if c, ok := m.Map[name]; ok {
		return c
	}

	return nil
}

func (m *MessageCommandMap) Del(name string) {
	m.Lock()
	defer m.Unlock()

	m.Init()
	delete(m.Map, name)
}