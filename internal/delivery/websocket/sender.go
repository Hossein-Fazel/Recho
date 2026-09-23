package websocket

import (
	"context"

	"github.com/Hossein-Fazel/Recho/internal/application"
	"github.com/Hossein-Fazel/Recho/pkg"
)

type Sender struct {
	ctx     context.Context
	deliver chan application.SendItem
}

var _ application.Sender = (*Sender)(nil)

func NewSender(ctx context.Context) *Sender {
	return &Sender{
		ctx:     ctx,
		deliver: make(chan application.SendItem),
	}
}

func (s *Sender) Broadcast(item application.SendItem) {
	select {
	case s.deliver <- item:
	case <-s.ctx.Done():
		pkg.Logger.Warn().
			Str("event", string(item.Event)).
			Msg("websocket sender: dropping broadcast, shutting down")
	}
}

func (s *Sender) Channel() <-chan application.SendItem {
	return s.deliver
}
