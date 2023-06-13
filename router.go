package main

import (
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func NewMessageCommandRouter(prefixes []string) *MessageCommandRouter {
	return &MessageCommandRouter{
		prefixes,
		new(MessageCommandMap),
		nil,
		nil,
	}
}

type MessageCommandMiddleware func(h MessageCommandHandler) MessageCommandHandler

type MessageCommandRouter struct {

	// To store all prefixes of the bot
	// Can change it to map[string]string or map[string][]string for server's custom prefix(es)
	Prefixes 			[]string

	// Map command name to the command
	// Key is the command name depends on IgnoreCase
	CommandsMapping 	*MessageCommandMap

	// Function invoked before the command
	Before func(*MessageCommandContext)

	// Function invoked after the command
	After func()
}

func (r *MessageCommandRouter) GetCommand(name string) *MessageCommand {
	lower := strings.ToLower(name)
	
	if c := r.CommandsMapping.Get(lower); c != nil {
		return c
	} 

	if c := r.CommandsMapping.Get(name); c != nil {
		return c
	}

	return nil
}

func (r *MessageCommandRouter) AddCommand(cmd *MessageCommand) {
	r.CommandsMapping.Set(cmd)
}

func (r *MessageCommandRouter) Handler() func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.Bot {
			return
		}

		// Get prefixes
		_, prefix, exists := slicePrefixesString(m.Content, r.Prefixes)

		if !exists {
			return
		}

		commandName, arguments := parseContent(m.Content, prefix)

		var cmd *MessageCommand
		if cmd = r.GetCommand(commandName); cmd == nil {
			return
		}

		// Validate arguments
		if !cmd.ValidateArguments(arguments) {
			return
		}

		// Get converted arguments
		var conv map[string]interface{}
		if converted, err := cmd.ConvertArguments(s, m.Message, arguments); err != nil {
			msg, er := s.ChannelMessageSendReply(
				m.ChannelID, 
				"Not valid arguments. Use help <command> for more info", 
				&discordgo.MessageReference{
					ChannelID: m.ChannelID,
					MessageID: m.ID,
					GuildID: m.GuildID,
			})
			if er != nil {
				return
			}
			go func() {
				time.Sleep(time.Second * 5)
				s.ChannelMessageDelete(msg.ChannelID, msg.ID)
			}()
			return
		} else {
			conv = converted
		}

		ctx := &MessageCommandContext{
			s,
			m.Message,
			r,
			commandName,
			arguments,
			conv,
			cmd,
		}

		handler := cmd.Handler

		go func ()  {
			if r.Before != nil {
				r.Before(ctx)
			}
			handler(ctx)

			if r.After != nil {
				r.After()
			}
		}()
	}
}