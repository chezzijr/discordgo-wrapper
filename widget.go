package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

type WidgetHandler func(w *Widget, i *discordgo.Interaction)

type Widget struct {
	sync.Mutex
	// The embed that is displaying
	View 			*discordgo.MessageSend
	Message			*discordgo.Message
	Session 		*discordgo.Session
	ChannelID 		string
	Timeout 		time.Duration
	Close			chan bool

	TotalPages		int

	// Bind the label to function that handle button
	Handlers		map[string]WidgetHandler

	// Buttons interface for interacting with pagination
	Controller		[]discordgo.MessageComponent

	// Check if using the default controller
	DefaultCtrl		bool

	// Users that have access to buttons
	UserWhitelist	[]string

	Running	bool
}

func NewWidget(s *discordgo.Session, channelID string, msg *discordgo.MessageSend) *Widget {
	return &Widget{
		ChannelID:      channelID,
		Session:        s,
		Handlers:       map[string]WidgetHandler{},
		Close:          make(chan bool),
		View:   		msg,
	}
}

func (w *Widget) IsUserAllowed(userID string) bool {
	if w.UserWhitelist == nil || len(w.UserWhitelist) == 0 {
		return true
	}
	for _, user := range w.UserWhitelist {
		if user == userID {
			return true
		}
	}
	return false
}

func (w *Widget) DefaultController(page int) []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Style: discordgo.DangerButton,
					CustomID: "First",
					Label: "First",
				},
				discordgo.Button{
					Style: discordgo.SuccessButton,
					CustomID: "Prev",
					Label: "Prev",
				},
				discordgo.Button{
					Style: discordgo.SecondaryButton,
					CustomID: "Page",
					Label: fmt.Sprintf("Page %d/%d", page, w.TotalPages),
					Disabled: true,
				},
				discordgo.Button{
					Style: discordgo.SuccessButton,
					CustomID: "Next",
					Label: "Next",
				},
				discordgo.Button{
					Style: discordgo.DangerButton,
					CustomID: "Last",
					Label: "Last",
				},
			},
		},
	}
}

func (w *Widget) InitializeDefaultController() {
	w.DefaultCtrl = true
	w.Controller = w.DefaultController(1)
}

// Deploy the widget in channel w.ChannelID
func (w *Widget) Deploy() error {
	if w.IsRunning() {
		return ErrAlreadyRunning
	}
	w.Running = true
	defer func() {
		w.Running = false
	}()

	if w.View == nil {
		return ErrNilPage
	}

	// startTime := time.Now()

	// Create initial message.
	msg, err := w.Session.ChannelMessageSendComplex(w.ChannelID, w.View)
	if err != nil {
		return err
	}
	
	w.Message = msg

	var interaction *discordgo.Interaction

	for {
		if w.Timeout != 0 {
			select {
			case i := <-nextInteractionCreateChannel(w.Session):
				interaction = i.Interaction
			case <-time.After(w.Timeout):
				return nil
			case <-w.Close:
				return nil
			}
		} else /* Navigation timeout not enabled */ {
			select {
			case i := <-nextInteractionCreateChannel(w.Session):
				interaction = i.Interaction
			case <-w.Close:
				return nil
			}
		}

		// Check valid Interaction
		if interaction.Type != discordgo.InteractionMessageComponent {
			continue
		}
		if interaction.MessageComponentData().ComponentType != discordgo.ButtonComponent {
			continue
		}
		if interaction.Message.ID != w.Message.ID {
			continue
		}

		if h, ok := w.Handlers[interaction.MessageComponentData().CustomID]; ok {
			if w.IsUserAllowed(interaction.Member.User.ID) {
				go h(w, interaction)
			}
		}
	}
}

// Handle adds a handler for the given emoji name
//    action: The CustomID of button
//    handler  : handler function to call when the button is clicked
//               func(*Widget, *discordgo.InteractionCreate)
func (w *Widget) AddHandler(action string, handler WidgetHandler) {
	if _, ok := w.Handlers[action]; !ok {
		w.Handlers[action] = handler
	}
}

// QueryInput querys the user with ID `id` for input
//    prompt : Question prompt
//    userID : UserID to get message from
//    timeout: How long to wait for the user's response
func (w *Widget) QueryInput(prompt string, userID string, timeout time.Duration) (*discordgo.Message, error) {
	msg, err := w.Session.ChannelMessageSend(w.ChannelID, "<@"+userID+">,  "+prompt)
	if err != nil {
		return nil, err
	}
	defer func() {
		w.Session.ChannelMessageDelete(msg.ChannelID, msg.ID)
	}()

	timeoutChan := make(chan int)
	go func() {
		time.Sleep(timeout)
		timeoutChan <- 0
	}()

	for {
		select {
		case usermsg := <-nextMessageCreateChannel(w.Session):
			if usermsg.Author.ID != userID {
				continue
			}
			w.Session.ChannelMessageDelete(usermsg.ChannelID, usermsg.ID)
			return usermsg.Message, nil
		case <-timeoutChan:
			return nil, ErrTimeout
		}
	}
}

func (w *Widget) IsRunning() bool {
	w.Lock()
	defer w.Unlock()
	return w.Running
}

func (w *Widget) UpdateEmbed(msg *discordgo.MessageSend, i *discordgo.Interaction, index int) error {
	if w.Message == nil {
		return ErrNilMessage
	}

	resp := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: toInteractionResponseData(msg),
	}

	if w.DefaultCtrl {
		resp.Data.Components = w.DefaultController(index + 1)
	}

	return w.Session.InteractionRespond(i, resp)
}