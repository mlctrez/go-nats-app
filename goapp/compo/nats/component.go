package nats

import (
	"context"
	"github.com/google/uuid"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/nats-io/nats.go"
	"nhooyr.io/websocket"
	"time"
)

var _ app.Mounter = (*Nats)(nil)

type Nats struct {
	app.Compo
	clientId   string
	conn       *websocket.Conn
	connCtx    context.Context
	connCancel context.CancelFunc
	natsConn   *nats.Conn
}

func (w *Nats) Render() app.UI {
	return app.Div().Style("display", "none").DataSet("name", "nats-component")
}

func (w *Nats) OnMount(ctx app.Context) {
	w.establishClientId(ctx)
	Action(ctx).handle(w)
	ctx.Async(func() { w.connectWebSocket(ctx) })
}

func (w *Nats) establishClientId(ctx app.Context) {
	var err error
	err = ctx.LocalStorage().Get("nats.client.name", &w.clientId)
	if err != nil {
		app.Log("error getting client id from local storage : %s", err)
	}
	if w.clientId == "" {
		w.clientId = uuid.NewString()
		err = ctx.LocalStorage().Set("nats.client.name", w.clientId)
		if err != nil {
			app.Log("error setting client id to local storage : %s", err)
		}
	}
}

func (w *Nats) callbackHandler(ctx app.Context, action app.Action) {
	if cb, ok := action.Value.(ConnReceiver); ok {
		cb.WithConn(ctx, w.natsConn)
	}
}

func (w *Nats) cancelAndReconnect(ctx app.Context) {
	w.connCancel()
	ctx.After(time.Second*5, func(c app.Context) {
		c.Async(func() { w.connectWebSocket(c) })
	})
}

func (w *Nats) connectWebSocket(ctx app.Context) {
	w.connCtx, w.connCancel = context.WithCancel(ctx)

	var err error
	w.natsConn, err = nats.Connect("not_used", nats.Name(w.clientId), nats.SetCustomDialer(w))

	if err != nil {
		w.cancelAndReconnect(ctx)
	}
}
