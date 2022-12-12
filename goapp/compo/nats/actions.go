package nats

import (
	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/nats-io/nats.go"
)

func Action(ctx app.Context) *Actions {
	return &Actions{ctx}
}

type Actions struct {
	app.Context
}

func (ac *Actions) handle(webSocket *Nats) {
	ac.Handle("nats.callback", webSocket.callbackHandler)
}

func (ac *Actions) WithConn(cb ConnReceiver) {
	ac.NewActionWithValue("nats.callback", cb)
}

type ConnReceiver interface {
	WithConn(ctx app.Context, conn *nats.Conn)
}
