package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	r *MessageCommandRouter
	s *discordgo.Session
)

func init() {
	var err error
	if s, err = discordgo.New("Bot " + ""); err != nil {
		panic(err)
	}
	r = NewMessageCommandRouter([]string{"t."})
}

func init() {
	r.AddCommand(NewMessageCommand(
		"random",
		"randomly pick from a sequence",
		[]string{"random ligma sawcon"},
		true,
		[]*MessageCommandParam{{"strings", MessageCommandParamTypeUser, MessageCommandParamOptionList}},
		[]*MessageCommand{},
		func(ctx *MessageCommandContext) {
			tmp := ctx.ConvertedArgs["strings"].([]interface{})
			li := make([]*discordgo.User, len(tmp))
			for i, v := range tmp {
				li[i] = v.(*discordgo.User)
			}
			ctx.Respond(&discordgo.MessageSend{
				Content: fmt.Sprintf("%+v", li[rand.Intn(len(li))].AvatarURL("")),
			})
		},
	))

	r.AddCommand(NewMessageCommand(
		"randrange",
		"random a range",
		[]string{"randrange 1 100"},
		true,
		[]*MessageCommandParam{
			{"first", MessageCommandParamTypeInteger, MessageCommandParamOptionRequired},
			{"second", MessageCommandParamTypeInteger, MessageCommandParamOptionRequired},
		},
		[]*MessageCommand{},
		func(ctx *MessageCommandContext) {
			first, second := ctx.ConvertedArgs["first"].(int), ctx.ConvertedArgs["second"].(int)
			rand.Seed(time.Now().UnixNano())
			val := rand.Intn(second-first+1) + first
			ctx.RespondText(val)
		},
	))

	r.AddCommand(NewMessageCommand(
		"help",
		"use help",
		[]string{"help games"},
		true,
		[]*MessageCommandParam{{"command", MessageCommandParamTypeString, MessageCommandParamOptionRequired}},
		[]*MessageCommand{},
		func(ctx *MessageCommandContext) {
			e := r.GetCommand(ctx.ConvertedArgs["command"].(string)).Embed()

			ctx.Respond(&discordgo.MessageSend{
				Embed: e,
			})
		},
	))

	s.AddHandler(r.Handler())
}

func main() {

	if err := s.Open(); err != nil {
		panic(err)
	}
	defer s.Close()

	log.Println("Connected")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	log.Println("Disconnected")
}
