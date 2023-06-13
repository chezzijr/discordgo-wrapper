package main

import (
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

type Paginator struct {
	sync.Mutex
	Pages		[]*discordgo.MessageSend
	Index 		int

	// Loop back to the beginning or end when on the first or last page.
	Loop   		bool
	Widget 		*Widget

	Session 	*discordgo.Session

	Running 	bool
}

// NewPaginator returns a new Paginator
//    s        : discordgo session
//    channelID: channelID to spawn the paginator on
func NewPaginator(s *discordgo.Session, channelID string) *Paginator {
	p := &Paginator{
		Session:   	s,
		Pages: 		[]*discordgo.MessageSend{},
		Index: 		0,
		Loop:  		true,
		Widget:     NewWidget(s, channelID, nil),
	}
	p.AddHandlers()

	return p
}

func (p *Paginator) Deploy() error {
	if len(p.Pages) == 0 {
		return ErrPagesEmpty
	}
	if p.IsRunning() {
		return ErrAlreadyRunning
	}
	p.Lock()
	p.Running = true
	p.Unlock()

	defer func() {
		p.Lock()
		p.Running = false
		p.Unlock()

		if p.Widget.Message == nil {
			return
		}

		p.Session.ChannelMessageEditComplex(&discordgo.MessageEdit{
			Components: []discordgo.MessageComponent{},
			Channel: p.Widget.ChannelID,
			ID: p.Widget.Message.ID,
		})
	}()

	page, err := p.Page()
	if err != nil {
		return err
	}
	p.Widget.TotalPages = len(p.Pages)
	p.Widget.InitializeDefaultController()
	p.Widget.View = page
	p.Widget.View.Components = p.Widget.Controller
	return p.Widget.Deploy()
}

func (p *Paginator) SetTimeout(duration time.Duration) {
	p.Widget.Timeout = duration
}

func (p *Paginator) AddHandlers() {
	p.Widget.AddHandler("First", func(w *Widget, i *discordgo.Interaction) {
		if err := p.Goto(0); err == nil {
			p.Update(i)
		}
	})
	p.Widget.AddHandler("Prev", func(w *Widget, i *discordgo.Interaction) {
		if err := p.PreviousPage(); err == nil {
			p.Update(i)
		}
	})
	p.Widget.AddHandler("Next", func(w *Widget, i *discordgo.Interaction) {
		if err := p.NextPage(); err == nil {
			p.Update(i)
		}
	})
	p.Widget.AddHandler("Last", func(w *Widget, i *discordgo.Interaction) {
		if err := p.Goto(len(p.Pages) - 1); err == nil {
			p.Update(i)
		}
	})
}

func (p *Paginator) Add(msg ...*discordgo.MessageSend) {
	p.Pages = append(p.Pages, msg...)
}

func (p *Paginator) Page() (*discordgo.MessageSend, error) {
	p.Lock()
	defer p.Unlock()

	if p.Index < 0 || p.Index >= len(p.Pages) {
		return nil, ErrIndexOutOfBounds
	}

	return p.Pages[p.Index], nil
}

func (p *Paginator) NextPage() error {
	p.Lock()
	defer p.Unlock()

	if p.Index+1 >= 0 && p.Index+1 < len(p.Pages) {
		p.Index++
		return nil
	}

	// Set the queue back to the beginning if Loop is enabled.
	if p.Loop {
		p.Index = 0
		return nil
	}

	return ErrIndexOutOfBounds
}

func (p *Paginator) PreviousPage() error {
	p.Lock()
	defer p.Unlock()

	if p.Index-1 >= 0 && p.Index-1 < len(p.Pages) {
		p.Index--
		return nil
	}

	// Set the queue back to the beginning if Loop is enabled.
	if p.Loop {
		p.Index = len(p.Pages) - 1
		return nil
	}

	return ErrIndexOutOfBounds
}

// Goto jumps to the requested page index
//    index: The index of the page to go to
func (p *Paginator) Goto(index int) error {
	p.Lock()
	defer p.Unlock()
	if index < 0 || index >= len(p.Pages) {
		return ErrIndexOutOfBounds
	}
	p.Index = index
	return nil
}

func (p *Paginator) Update(i *discordgo.Interaction) error {
	if p.Widget.Message == nil {
		return ErrNilMessage
	}

	page, err := p.Page()
	if err != nil {
		return err
	}

	err = p.Widget.UpdateEmbed(page, i, p.Index)
	return err
}

// Running returns the running status of the paginator
func (p *Paginator) IsRunning() bool {
	p.Lock()
	defer p.Unlock()
	return p.Running
}

// SetPageFooters sets the footer of each embed to
// Be its page number out of the total length of the embeds.
// func (p *PaginatorEmbed) SetPageFooters() {
// 	for index, embed := range p.Pages {
// 		embed.Footer = &discordgo.MessageEmbedFooter{
// 			Text: fmt.Sprintf("#[%d / %d]", index+1, len(p.Pages)),
// 		}
// 	}
// }