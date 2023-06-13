package main

import "github.com/bwmarrin/discordgo"

type CheckBox struct {
	Button 	*discordgo.Button
}

type View struct {
	FirstRow 	[]discordgo.MessageComponent
	SecondRow 	[]discordgo.MessageComponent
	ThirdRow 	[]discordgo.MessageComponent
	FourthRow 	[]discordgo.MessageComponent
	FifthRow 	[]discordgo.MessageComponent
}