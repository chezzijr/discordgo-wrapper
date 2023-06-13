package main

import (
	"sort"

	"github.com/bwmarrin/discordgo"
)

type MessageCommandCategory struct {
	Name        string
	Description string
	Emoji       string

	Item []*MessageCommand
}

func (c *MessageCommandCategory) SortCommand(key func(i, j int) bool) {
	sort.Slice(c.Item, key)
}

func (c *MessageCommandCategory) Embed() *discordgo.MessageEmbed {
	f := make([]*discordgo.MessageEmbedField, len(c.Item))
	for i, cmd := range c.Item {
		f[i] = &discordgo.MessageEmbedField{
			Name: cmd.Name,
			Value: cmd.Description,
			Inline: true,
		}
	}
	return &discordgo.MessageEmbed{
		Title: c.Name + " CATEGORY",
		Fields: f,
	}
}