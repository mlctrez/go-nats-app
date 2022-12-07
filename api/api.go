package api

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/o1egl/govatar"
	"image"
	"image/jpeg"
	"net"
	"net/http"
	"net/url"
	"nhooyr.io/websocket"
	"strings"
)

type Api struct {
	ns         *server.Server
	serverConn *nats.Conn
}

func New(ns *server.Server) *Api {
	return &Api{ns: ns}
}

func (a *Api) Register(engine *gin.Engine) {

	a.setupNatsApi()

	// web socket endpoint for websocket api calls
	engine.GET("/ws/:clientId", a.webSocketHandler)

	engine.GET("/echo/:who", a.echo)

}

func (a *Api) setupNatsApi() {

	// setupNatsApiDocLink

	var err error
	a.serverConn, err = nats.Connect("", nats.InProcessServer(a.ns))
	if err != nil {
		app.Logf("error creating serverConn %w", err)
		return
	}

	nc := a.serverConn

	_, err = nc.Subscribe("chat.say", func(msg *nats.Msg) {
		//fmt.Printf("Got: %q\n", msg.Data)
		err = nc.Publish("chat.room", msg.Data)
		if err != nil {
			app.Logf("error publishing chat.room message %w", err)
		}
	})
	if err != nil {
		app.Logf("error subscribing to chat.say %w", err)
	}

	_, err = nc.Subscribe("govatar.female", func(msg *nats.Msg) {

		var img image.Image
		// always female and random
		img, err = govatar.Generate(govatar.FEMALE)
		if err != nil {
			app.Logf("error generating govatar %w", err)
			return
		}
		var buf bytes.Buffer
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80})
		if err != nil {
			app.Logf("error generating jpeg %w", err)
			return
		}
		err = msg.Respond([]byte("data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())))
		if err != nil {
			app.Logf("error responding to message %w", err)
			return
		}
	})
	if err != nil {
		app.Logf("error subscribing to govatar.female %w", err)
		return
	}

}

func (a *Api) webSocketHandler(ginCtx *gin.Context) {

	var err error
	var conn *websocket.Conn

	clientId := ginCtx.Param("clientId")
	fmt.Println("websocket connect", clientId, ginCtx.Request.RemoteAddr)

	var options *websocket.AcceptOptions
	// https://github.com/gorilla/websocket/issues/731
	// Compression in certain Safari browsers is broken, turn it off
	if strings.Contains(ginCtx.Request.UserAgent(), "Safari") {
		options = &websocket.AcceptOptions{CompressionMode: websocket.CompressionDisabled}
	}

	if conn, err = websocket.Accept(ginCtx.Writer, ginCtx.Request, options); err != nil {
		_ = ginCtx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx, cancelFunc := context.WithCancel(context.TODO())
	defer cancelFunc()

	// bridge websocket connection to nats net connection

	var natsUrl *url.URL
	if natsUrl, err = url.Parse(a.ns.ClientURL()); err != nil {
		_ = ginCtx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if natsUrl.Scheme != "nats" {
		err = fmt.Errorf("unsupported nats connection type %q", natsUrl.Scheme)
		_ = ginCtx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	var natsRawTcp net.Conn
	if natsRawTcp, err = net.Dial("tcp4", natsUrl.Host); err != nil {
		_ = ginCtx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	connClosed := false
	natsClosed := false

	go func() {
		var bytes []byte
		var readErr error
		for {
			if _, bytes, readErr = conn.Read(ctx); readErr != nil {
				// when user closes window, websocket sends this status
				if strings.Contains(readErr.Error(), "status = StatusGoingAway") {
					break
				}
				fmt.Println("conn.Read error", readErr)
				break
			}
			fmt.Printf("client -> %q\n", string(bytes))
			if _, writeErr := natsRawTcp.Write(bytes); writeErr != nil {
				fmt.Println("natsRawTcp.Write error", writeErr)
				break
			}
		}

		connClosed = true
		conn.Close(websocket.StatusNormalClosure, "nats close")
		natsClosed = true
		natsRawTcp.Close()
	}()

	reader := bufio.NewReader(natsRawTcp)
	truncateAt := 300
	for !natsClosed {
		var readBytes []byte
		var readErr error
		if readBytes, readErr = reader.ReadBytes('\n'); err != nil && !natsClosed {
			fmt.Println("natsRawTcp read error", readErr)
			break
		}

		if len(readBytes) > truncateAt {
			fmt.Printf("  <- nats %q ... truncated\n", strings.TrimSuffix(string(readBytes[0:truncateAt-1]), "\r\n"))
		} else {
			fmt.Printf("  <- nats %q\n", strings.TrimSuffix(string(readBytes), "\r\n"))
		}

		writeErr := conn.Write(ctx, websocket.MessageBinary, readBytes)
		if writeErr != nil && !natsClosed {
			fmt.Println("conn.Write error", writeErr)
			break
		}
	}

	fmt.Println("websocket disconnect", clientId, ginCtx.Request.RemoteAddr)

	if !natsClosed {
		_ = natsRawTcp.Close()
	}
	if !connClosed {
		_ = conn.Close(websocket.StatusAbnormalClosure, "should already be closed")
	}

}

func (a *Api) echo(c *gin.Context) {
	if a.serverConn != nil {
		who := c.Param("who")

		err := a.serverConn.Publish(fmt.Sprintf("echo.%s", who), []byte("{{ Random 10 100}}"))
		if err != nil {
			app.Logf("error publishing message %w", err)
		}
	}
}
