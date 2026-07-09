package fakestaging

import (
	"context"

	discordruntime "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/runtime"
)

type Gateway struct {
	Opens     int
	Closes    int
	Registers int
	ReadyCh   chan struct{}
}

func NewGateway() *Gateway {
	return &Gateway{ReadyCh: make(chan struct{})}
}

func (g *Gateway) Open() error {
	g.Opens++
	return nil
}

func (g *Gateway) Close() error {
	g.Closes++
	return nil
}

func (g *Gateway) RegisterInteractionHandler(discordruntime.Handler) func() {
	g.Registers++
	return func() {}
}

func (g *Gateway) Ready() <-chan struct{} {
	return g.ReadyCh
}

func (g *Gateway) MarkReady() {
	close(g.ReadyCh)
}

func WaitReady(ctx context.Context, gateway *Gateway) error {
	return discordruntime.WaitReady(ctx, gateway)
}
