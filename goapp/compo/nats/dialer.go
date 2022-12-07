package nats

import (
	"context"
	"fmt"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/mlctrez/go-nats-app/goapp"
	"net"
	"nhooyr.io/websocket"
)

func (w *Nats) Dial(_, _ string) (netCon net.Conn, err error) {

	wsUri := fmt.Sprintf("%s/ws/%s", goapp.BaseWs(), w.clientId)

	if w.conn, _, err = websocket.Dial(w.connCtx, wsUri, nil); err != nil {
		app.Logf("websocket.Dial failed %#v", err)
		return nil, fmt.Errorf("websocket.Dial failed %w", err)
	}

	netConn := websocket.NetConn(context.Background(), w.conn, websocket.MessageBinary)
	return netConn, nil
}
