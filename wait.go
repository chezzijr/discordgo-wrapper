package main

import (
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

type EventType string
type EventHandler interface {
	Type() EventType
	Handle(*discordgo.Session, interface{})
}

const (
	ChannelCreate            	EventType 	= "CHANNEL_CREATE"
	ChannelDelete            	EventType 	= "CHANNEL_DELETE"
	ChannelPinsUpdate        	EventType 	= "CHANNEL_PINS_UPDATE"
	ChannelUpdate            	EventType 	= "CHANNEL_UPDATE"
	GuildBanAdd              	EventType 	= "GUILD_BAN_ADD"
	GuildBanRemove           	EventType 	= "GUILD_BAN_REMOVE"
	GuildCreate              	EventType 	= "GUILD_CREATE"
	GuildDelete              	EventType 	= "GUILD_DELETE"
	GuildEmojisUpdate        	EventType 	= "GUILD_EMOJIS_UPDATE"
	GuildIntegrationsUpdate  	EventType 	= "GUILD_INTEGRATIONS_UPDATE"
	GuildMemberAdd           	EventType 	= "GUILD_MEMBER_ADD"
	GuildMemberRemove        	EventType 	= "GUILD_MEMBER_REMOVE"
	GuildMemberUpdate        	EventType 	= "GUILD_MEMBER_UPDATE"
	GuildMembersChunk        	EventType 	= "GUILD_MEMBERS_CHUNK"
	GuildRoleCreate          	EventType 	= "GUILD_ROLE_CREATE"
	GuildRoleDelete          	EventType 	= "GUILD_ROLE_DELETE"
	GuildRoleUpdate          	EventType 	= "GUILD_ROLE_UPDATE"
	GuildUpdate              	EventType 	= "GUILD_UPDATE"
	InteractionCreate        	EventType 	= "INTERACTION_CREATE"
	MessageAck               	EventType 	= "MESSAGE_ACK"
	MessageCreate            	EventType 	= "MESSAGE_CREATE"
	MessageDelete            	EventType 	= "MESSAGE_DELETE"
	MessageDeleteBulk        	EventType 	= "MESSAGE_DELETE_BULK"
	MessageReactionAdd       	EventType 	= "MESSAGE_REACTION_ADD"
	MessageReactionRemove    	EventType 	= "MESSAGE_REACTION_REMOVE"
	MessageReactionRemoveAll 	EventType 	= "MESSAGE_REACTION_REMOVE_ALL"
	MessageUpdate            	EventType 	= "MESSAGE_UPDATE"
	PresenceUpdate           	EventType 	= "PRESENCE_UPDATE"
	PresencesReplace         	EventType 	= "PRESENCES_REPLACE"
	Ready                    	EventType 	= "READY"
	RelationshipAdd          	EventType 	= "RELATIONSHIP_ADD"
	RelationshipRemove       	EventType 	= "RELATIONSHIP_REMOVE"
	Resumed                  	EventType 	= "RESUMED"
	TypingStart              	EventType 	= "TYPING_START"
	UserGuildSettingsUpdate  	EventType 	= "USER_GUILD_SETTINGS_UPDATE"
	UserNoteUpdate           	EventType 	= "USER_NOTE_UPDATE"
	UserSettingsUpdate       	EventType 	= "USER_SETTINGS_UPDATE"
	UserUpdate               	EventType 	= "USER_UPDATE"
	VoiceServerUpdate        	EventType 	= "VOICE_SERVER_UPDATE"
	VoiceStateUpdate         	EventType 	= "VOICE_STATE_UPDATE"
	WebhooksUpdate           	EventType 	= "WEBHOOKS_UPDATE"

	Interface 					EventType 	= "__INTERFACE__"
	Connect    					EventType   = "__CONNECT__"
	Disconnect			  		EventType   = "__DISCONNECT__"
	Event  						EventType	= "__EVENT__"
	RateLimit		 			EventType   = "__RATE_LIMIT__"
)


type WaitingNode struct {
	Prev		*WaitingNode
	Next 		*WaitingNode

	Validate	func(event interface{}) bool
	Channel		chan interface{}
	Closed      bool
}

func NewNode(channel chan interface{}, validate func(event interface{}) bool) *WaitingNode {
	return &WaitingNode{
		Validate: validate,
		Channel: channel,
		Closed: false,
	}
}

type WaitingList struct {
	sync.Mutex

	Event 	EventHandler

	Begin 	*WaitingNode
	End 	*WaitingNode
	Len 	int
}

func (li *WaitingList) Add(node *WaitingNode) {
	li.Lock()
	defer li.Unlock()

	if li.End == nil || li.Begin == nil || li.Len == 0 {
		li.Begin = node
		node.Prev = nil
		node.Next = nil
	} else {
		li.End.Next = node
		node.Prev = li.End
		node.Next = nil
	}
	li.End = node
	li.Len++
}

func (li *WaitingList) Traverse(event interface{}) {
	if li.Begin == nil || li.End == nil || li.Len == 0 {
		return
	}

	li.Lock()
	defer li.Unlock()

	current := li.Begin

	var prev *WaitingNode
	var begin, end *WaitingNode
	count := 0

	for current != nil {
		if current.Closed {
			if current.Prev != nil {
				current.Prev.Next = current.Next
			}
			if current.Next != nil {
				current.Next.Prev = current.Prev
			}		
		} else if current.Validate(event) {
			current.Channel <- event
		} else {
			if prev == nil {
				begin = current
				prev = current
				current.Prev = nil
			} else {
				prev.Next = current
				current.Prev = prev
				prev = current
			}
			count++
		}
		current = current.Next
	}
	end = prev
	if prev != nil {
		end.Next = nil
	}
	li.Begin, li.End, li.Len = begin, end, count
}

func (li *WaitingList) AddHandler(s *discordgo.Session, event EventType) {
	switch event {
	case ChannelCreate:
		s.AddHandler(func(s *discordgo.Session, e *discordgo.ChannelCreate) {
			go func() {
				li.Traverse(e)
			}()
		})
	case ChannelDelete:
		s.AddHandler(func(s *discordgo.Session, e *discordgo.ChannelDelete) {
			go func() {
				li.Traverse(e)
			}()
		})
	case ChannelPinsUpdate:
		s.AddHandler(func(s *discordgo.Session, e *discordgo.ChannelPinsUpdate) {
			go func() {
				li.Traverse(e)
			}()
		})
	case ChannelUpdate:
		s.AddHandler(func(s *discordgo.Session, e *discordgo.ChannelUpdate) {
			go func() {
				li.Traverse(e)
			}()
		})
	case GuildBanAdd:
		s.AddHandler(func(s *discordgo.Session, e *discordgo.GuildBanAdd) {
			go func() {
				li.Traverse(e)
			}()
		})
	case GuildBanRemove:
		s.AddHandler(func(s *discordgo.Session, e *discordgo.GuildBanRemove) {
			go func() {
				li.Traverse(e)
			}()
		})
	case GuildCreate:
		s.AddHandler(func(s *discordgo.Session, e *discordgo.GuildCreate) {
			go func() {
				li.Traverse(e)
			}()
		})
	case GuildDelete:
		s.AddHandler(func(s *discordgo.Session, e *discordgo.GuildDelete) {
			go func() {
				li.Traverse(e)
			}()
		})
	case GuildEmojisUpdate:
		s.AddHandler(func(s *discordgo.Session, e *discordgo.GuildEmojisUpdate) {
			go func() {
				li.Traverse(e)
			}()
		})
	case GuildIntegrationsUpdate:
		s.AddHandler(func(s *discordgo.Session, e *discordgo.GuildIntegrationsUpdate) {
			go func() {
				li.Traverse(e)
			}()
		})
	case GuildMemberAdd:
		s.AddHandler(func(s *discordgo.Session, e *discordgo.GuildMemberAdd) {
			go func() {
				li.Traverse(e)
			}()
		})
	case GuildMemberRemove:
		s.AddHandler(func(s *discordgo.Session, e *discordgo.GuildMemberRemove) {
			go func() {
				li.Traverse(e)
			}()
		})
	case GuildMemberUpdate:
		s.AddHandler(func(s *discordgo.Session, e *discordgo.GuildMemberUpdate) {
			go func() {
				li.Traverse(e)
			}()
		})
	case GuildMembersChunk:
		s.AddHandler(func(s *discordgo.Session, e *discordgo.GuildMembersChunk) {
			go func() {
				li.Traverse(e)
			}()
		})
	case GuildRoleCreate:
		s.AddHandler(func(s *discordgo.Session, e *discordgo.GuildRoleCreate) {
			go func() {
				li.Traverse(e)
			}()
		})
	case GuildRoleDelete:
		s.AddHandler(func(s *discordgo.Session, e *discordgo.GuildRoleDelete) {
			go func() {
				li.Traverse(e)
			}()
		})
	case GuildRoleUpdate:
		s.AddHandler(func(s *discordgo.Session, e *discordgo.GuildRoleUpdate) {
			go func() {
				li.Traverse(e)
			}()
		})
	case GuildUpdate:
		s.AddHandler(func(s *discordgo.Session, e *discordgo.GuildUpdate) {
			go func() {
				li.Traverse(e)
			}()
		})
	case InteractionCreate:
		s.AddHandler(func(s *discordgo.Session, e *discordgo.InteractionCreate) {
			go func() {
				li.Traverse(e)
			}()
		})
	case MessageAck:
		s.AddHandler(func(s *discordgo.Session, e *discordgo.MessageAck) {
			go func() {
				li.Traverse(e)
			}()
		})
	case MessageCreate:
		s.AddHandler(func(s *discordgo.Session, e *discordgo.MessageCreate) {
			go func() {
				li.Traverse(e)
			}()
		})
	case MessageDelete:
		s.AddHandler(func(s *discordgo.Session, e *discordgo.MessageDelete) {
			go func() {
				li.Traverse(e)
			}()
		})
	case MessageDeleteBulk:
		s.AddHandler(func(s *discordgo.Session, e *discordgo.MessageDeleteBulk) {
			go func() {
				li.Traverse(e)
			}()
		})
	case MessageReactionAdd:
		s.AddHandler(func(s *discordgo.Session, e *discordgo.MessageReactionAdd) {
			go func() {
				li.Traverse(e)
			}()
		})
	case MessageReactionRemove:
		s.AddHandler(func(s *discordgo.Session, e *discordgo.MessageReactionRemove) {
			go func() {
				li.Traverse(e)
			}()
		})
	case MessageReactionRemoveAll:
		s.AddHandler(func(s *discordgo.Session, e *discordgo.MessageReactionRemoveAll) {
			go func() {
				li.Traverse(e)
			}()
		})
	case MessageUpdate:
		s.AddHandler(func(s *discordgo.Session, e *discordgo.MessageUpdate) {
			go func() {
				li.Traverse(e)
			}()
		})
	case PresenceUpdate:
		s.AddHandler(func(s *discordgo.Session, e *discordgo.PresenceUpdate) {
			go func() {
				li.Traverse(e)
			}()
		})
	case PresencesReplace:
		s.AddHandler(func(s *discordgo.Session, e *discordgo.PresencesReplace) {
			go func() {
				li.Traverse(e)
			}()
		})
	case Ready:
		s.AddHandler(func(s *discordgo.Session, e *discordgo.Ready) {
			go func() {
				li.Traverse(e)
			}()
		})
	case RelationshipAdd:
		s.AddHandler(func(s *discordgo.Session, e *discordgo.RelationshipAdd) {
			go func() {
				li.Traverse(e)
			}()
		})
	case RelationshipRemove:
		s.AddHandler(func(s *discordgo.Session, e *discordgo.RelationshipRemove) {
			go func() {
				li.Traverse(e)
			}()
		})
	case Resumed:
		s.AddHandler(func(s *discordgo.Session, e *discordgo.Resumed) {
			go func() {
				li.Traverse(e)
			}()
		})
	case TypingStart:
		s.AddHandler(func(s *discordgo.Session, e *discordgo.TypingStart) {
			go func() {
				li.Traverse(e)
			}()
		})
	case UserGuildSettingsUpdate:
		s.AddHandler(func(s *discordgo.Session, e *discordgo.UserGuildSettingsUpdate) {
			go func() {
				li.Traverse(e)
			}()
		})
	case UserNoteUpdate:
		s.AddHandler(func(s *discordgo.Session, e *discordgo.UserNoteUpdate) {
			go func() {
				li.Traverse(e)
			}()
		})
	case UserSettingsUpdate:
		s.AddHandler(func(s *discordgo.Session, e *discordgo.UserSettingsUpdate) {
			go func() {
				li.Traverse(e)
			}()
		})
	case UserUpdate:
		s.AddHandler(func(s *discordgo.Session, e *discordgo.UserUpdate) {
			go func() {
				li.Traverse(e)
			}()
		})
	case VoiceServerUpdate:
		s.AddHandler(func(s *discordgo.Session, e *discordgo.VoiceServerUpdate) {
			go func() {
				li.Traverse(e)
			}()
		})
	case VoiceStateUpdate:
		s.AddHandler(func(s *discordgo.Session, e *discordgo.VoiceStateUpdate) {
			go func() {
				li.Traverse(e)
			}()
		})
	case WebhooksUpdate:
		s.AddHandler(func(s *discordgo.Session, e *discordgo.WebhooksUpdate) {
			go func() {
				li.Traverse(e)
			}()
		})
	}
}

func NewWaitingList() *WaitingList {
	return new(WaitingList)
}

type EventWaiter struct {
	sync.RWMutex

	waiterMapping map[EventType]*WaitingList
}

func (ew *EventWaiter) Get(event EventType) *WaitingList {
	ew.RLock()
	defer ew.RUnlock()

	return ew.waiterMapping[event]
}

func (ew *EventWaiter) Set(event EventType, wt *WaitingList) {
	ew.RLock()
	defer ew.RUnlock()

	ew.waiterMapping[event] = wt
}

func (ew *EventWaiter) WaitFor(event EventType, timeout time.Duration, check func(interface{}) bool) interface{} {
	channel := make(chan interface{})
	node := NewNode(channel, check)
	ew.Get(event).Add(node)
	select {
	case val := <- channel:
		return val
	case <- time.After(timeout):
		node.Closed = true
		close(channel)
		return nil
	}
}

func NewEventWaiter(s *discordgo.Session) *EventWaiter {
	m := &EventWaiter{
		waiterMapping: make(map[EventType]*WaitingList),
	}
	for _, event := range []EventType{
		ChannelCreate,
		ChannelDelete,
		ChannelPinsUpdate,
		ChannelUpdate,
		GuildBanAdd,
		GuildBanRemove,
		GuildCreate,
		GuildDelete,
		GuildEmojisUpdate,
		GuildIntegrationsUpdate,
		GuildMemberAdd,
		GuildMemberRemove,
		GuildMemberUpdate,
		GuildMembersChunk,
		GuildRoleCreate,
		GuildRoleDelete,
		GuildRoleUpdate,
		GuildUpdate,
		InteractionCreate,
		MessageAck,
		MessageCreate,
		MessageDelete,
		MessageDeleteBulk,
		MessageReactionAdd,
		MessageReactionRemove,
		MessageReactionRemoveAll,
		MessageUpdate,
		PresenceUpdate,
		PresencesReplace,
		RateLimit,
		Ready,
		RelationshipAdd,
		RelationshipRemove,
		Resumed,
		TypingStart,
		UserGuildSettingsUpdate,
		UserNoteUpdate,
		UserSettingsUpdate,
		UserUpdate,
		VoiceServerUpdate,
		VoiceStateUpdate,
		WebhooksUpdate,
	} {
		w := NewWaitingList()
		w.AddHandler(s, event)
		m.Set(event, w)
	}	
	return m
}