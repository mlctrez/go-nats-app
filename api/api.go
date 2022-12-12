package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
	natsws "github.com/mlctrez/goapp-natsws"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/o1egl/govatar"
	"image"
	"image/jpeg"
)

type Api struct {
	ns         *server.Server
	serverConn *nats.Conn
}

func New(ns *server.Server) *Api {
	return &Api{ns: ns}
}

func (a *Api) Register(engine *gin.Engine) {

	proxy := &natsws.Proxy{
		Context: context.TODO(),
		Manager: natsws.StaticManager(true, "ws://127.0.0.1:8242"),
	}

	engine.GET("/natsws/:clientId", gin.WrapH(proxy))

	engine.GET("/echo/:who", a.echo)

	a.setupNatsApi()

}

func (a *Api) setupNatsApi() {

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

func (a *Api) echo(c *gin.Context) {
	if a.serverConn != nil {
		who := c.Param("who")

		err := a.serverConn.Publish(fmt.Sprintf("echo.%s", who), []byte("{{ Random 10 100}}"))
		if err != nil {
			app.Logf("error publishing message %w", err)
		}
	}
}

type NatsInfoMsg struct {
	ServerId    string   `json:"server_id"`
	ServerName  string   `json:"server_name"`
	Version     string   `json:"version"`
	Proto       int      `json:"proto"`
	Go          string   `json:"go"`
	Host        string   `json:"host"`
	Port        int      `json:"port"`
	Headers     bool     `json:"headers"`
	MaxPayload  int      `json:"max_payload"`
	ClientId    int      `json:"client_id"`
	ClientIp    string   `json:"client_ip"`
	Cluster     string   `json:"cluster"`
	ConnectUrls []string `json:"connect_urls"`
}
