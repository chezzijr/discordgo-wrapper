package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// The context contains infos about invoked command
type MessageCommandContext struct {
	Session 		*discordgo.Session
	Message 		*discordgo.Message
	Router  		*MessageCommandRouter

	// The name that trigger the command
	Trigger			string

	// Raw arguments that user passed in as strings
	RawArgs 		[]string

	// Map name of params to the value
	// Remember to assert type because default type is interface{}
	ConvertedArgs	map[string]interface{}

	Command 		*MessageCommand
}

func (ctx *MessageCommandContext) RespondText(s... interface{}) (*discordgo.Message, error) {
	msg := &discordgo.MessageSend{
		Content: fmt.Sprint(s...),
		Reference: &discordgo.MessageReference{
			MessageID: ctx.Message.ID,
			ChannelID: ctx.Message.ChannelID,
			GuildID: ctx.Message.GuildID,
		},
	}
	return ctx.Session.ChannelMessageSendComplex(ctx.Message.ChannelID, msg)
}

// Reply to the message that invoked the command
func (ctx *MessageCommandContext) Respond(d *discordgo.MessageSend) (*discordgo.Message, error) {
	d.Reference = &discordgo.MessageReference{
		MessageID: ctx.Message.ID,
		ChannelID: ctx.Message.ChannelID,
		GuildID: ctx.Message.GuildID,
	}
	m, err := ctx.Session.ChannelMessageSendComplex(ctx.Message.ChannelID, d)
	return m, err
}

// Send the message to the channel invoked the command
func (ctx *MessageCommandContext) Send(d *discordgo.MessageSend) (*discordgo.Message, error) {
	m, err := ctx.Session.ChannelMessageSendComplex(ctx.Message.ChannelID, d)
	return m, err
}

func (ctx *MessageCommandContext) SendText(s string) (*discordgo.Message, error) {
	m, err := ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, s)
	return m, err
}