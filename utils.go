package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// Check for every element in array is a prefix for a string
func slicePrefixesString(s string, arr []string) (index int, prefix string, exists bool) {
	for i, val := range arr {
		if strings.HasPrefix(s, val) {
			return i, val, true
		}
	}

	return -1, "", false
}

func parseContent(s string, prefix string) (commandName string, arguments []string) {
	s = strings.TrimPrefix(s, prefix)
	tokens := strings.Split(s, " ")
	return tokens[0], tokens[1:]
}

func userConverter(s *discordgo.Session, str string) (*discordgo.User, error) {
	if strings.HasPrefix(str, "<@") && strings.HasSuffix(str, ">") {		
		re, _ := regexp.Compile("[0-9]+")
		m := string(re.Find([]byte(str)))
		
		if u, err := s.User(m); err != nil {
			return nil, err
		} else {
			return u, nil
		}
		
	}
	if _, err := strconv.Atoi(str); err == nil {
		if u, er := s.User(str); er != nil {
			return nil, er
		} else{
			return u, nil
		}
	}
	if strings.Contains(str, "#") {
		for _, g := range s.State.Guilds {
			for _, m := range g.Members {
				if m.User.String() == str {
					return m.User, nil
				}
			}
		}
		return nil, fmt.Errorf("cannot find that user")
	}
	return nil, fmt.Errorf("%s does not match any format for finding user (mention, id, name#tag)", str)
}

func roleConverter(s *discordgo.Session, guildID string, str string) (*discordgo.Role, error) {
	if strings.HasPrefix(str, "<") && strings.HasSuffix(str, ">") {		
		rID := str[4:len(str)-1]
		if r, err := s.State.Role(guildID, rID); err != nil {
			return nil, err
		} else{
			return r, nil
		}
	}
	if _, err := strconv.Atoi(str); err == nil {
		if r, er := s.State.Role(guildID, str); er != nil {
			return nil, er
		} else {
			return r, nil
		}
	}
	return nil, fmt.Errorf("%s does not match any format for finding role (mention, id)", str)
}

func channelConverter(s *discordgo.Session, guildID string, str string) (*discordgo.Channel, error) {
	if strings.HasPrefix(str, "<") && strings.HasSuffix(str, ">") {		
		cID := str[3:len(str)-1]
		if c, err := s.Channel(cID); err != nil {
			return nil, err
		} else{
			return c, nil
		}
	}
	if _, err := strconv.Atoi(str); err == nil {
		if c, er := s.Channel(str); er != nil {
			return nil, er
		} else if c.GuildID != guildID {
			return nil, fmt.Errorf("that channel does not belong to this server")
		} else {
			return c, nil
		}
	}
	return nil, fmt.Errorf("%s does not match any format for finding channel (mention, id)", str)
}

func argumentsConverter(
	s 		*discordgo.Session,
	m 		*discordgo.Message,
	arg 	string, 
	Type 	MessageCommandParamType,
) (res interface{}, err error) {

	switch Type {
	case MessageCommandParamTypeString:
		return arg, nil
	case MessageCommandParamTypeInteger:
		if res, err := strconv.Atoi(arg); err != nil {
			return nil, fmt.Errorf("%s is not integer", arg)
		} else {
			return res, nil
		}
	case MessageCommandParamTypeBoolean:
		if res, err := strconv.ParseBool(arg); err != nil {
			return nil, fmt.Errorf("cannot convert %s to boolean", arg)
		} else {
			return res, nil
		}
	case MessageCommandParamTypeUser:
		if u, err := userConverter(s, arg); err != nil {
			return nil, err
		} else {
			return u, nil
		}
	case MessageCommandParamTypeRole:
		if r, err := roleConverter(s, m.GuildID, arg); err != nil {
			return nil, err
		} else {
			return r, nil
		}
	case MessageCommandParamTypeChannel:
		if c, err := channelConverter(s, m.GuildID, arg); err != nil {
			return nil, err
		} else {
			return c, err
		}
	}
	return nil, fmt.Errorf("cannot recognize the parameter type")
}

func nextMessageCreateChannel(s *discordgo.Session) chan *discordgo.MessageCreate {
	out := make(chan *discordgo.MessageCreate)
	s.AddHandlerOnce(func(_ *discordgo.Session, e *discordgo.MessageCreate) {
		out <- e
	})
	return out
}

func nextInteractionCreateChannel(s *discordgo.Session) chan *discordgo.InteractionCreate {
	out := make(chan *discordgo.InteractionCreate)
	s.AddHandlerOnce(func(_ *discordgo.Session, e *discordgo.InteractionCreate) {
		out <- e
	})
	return out
}

func toInteractionResponseData(msg *discordgo.MessageSend) *discordgo.InteractionResponseData {
	return &discordgo.InteractionResponseData{
		TTS: msg.TTS,
		Content: msg.Content,
		Embeds: msg.Embeds,
		AllowedMentions: msg.AllowedMentions,
		Files: msg.Files,
		Components: msg.Components,
	}
}