package front

import (
	"fmt"
	"github.com/goombaio/namegenerator"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/mlctrez/go-nats-app/goapp"
	nws "github.com/mlctrez/go-nats-app/goapp/compo/nats"
	"github.com/nats-io/nats.go"
	"io"
	"net/http"
	"strconv"
	"time"
)

type Front struct {
	app.Compo
	whoAmi               string     // keeps our name
	avatar               string     // data base64 jpeg image
	nc                   *nats.Conn // is the nats server connection
	input                string     // the current input line
	messages             []string   // the last 10 messages we received
	echoCount            int        // we count how many echos we sent on the echo service
	echoSubscription     *nats.Subscription
	chatRoomSubscription *nats.Subscription
}

var _ app.Initializer = (*Front)(nil)  // Verify the implementation
var _ app.Mounter = (*Front)(nil)      // Verify the implementation
var _ nws.ConnReceiver = (*Front)(nil) // Verify the implementation

// OnInit is called before the component gets mounted
// This is before Render was called for the first time
func (uc *Front) OnInit() {
	app.Log("OnInit")
	uc.whoAmi = namegenerator.NewNameGenerator(time.Now().UTC().UnixNano()).Generate()
	// a blank image
	uc.avatar = "data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7"
}

// OnMount gets called when the component is mounted
// This is after Render was called for the first time
func (uc *Front) OnMount(ctx app.Context) {
	app.Log("OnMount")

	go uc.reconnect(ctx)

	// Notice: We do not care about OnDismount which would be needed
	// when working with a more complex app.
}

func (uc *Front) WithConn(_ app.Context, conn *nats.Conn) {
	uc.nc = conn
}

func (uc *Front) disconnect(ctx app.Context) {
	if uc.chatRoomSubscription != nil {
		_ = uc.chatRoomSubscription.Unsubscribe()
	}
	if uc.echoSubscription != nil {
		_ = uc.echoSubscription.Unsubscribe()
	}
	uc.nc = nil
}

func (uc *Front) reconnect(ctx app.Context) {

	if uc.nc != nil {
		app.Log("uc.nc != nil, not reconnecting")
		return
	}

	start := time.Now()
	for uc.nc == nil {
		nws.Action(ctx).WithConn(uc)
		time.Sleep(20 * time.Millisecond)
		if time.Since(start) > time.Second*5 {
			app.Log("unable to get nats connection")
			return
		}
	}
	app.Logf("got nats conn after %s", time.Since(start))

	var err error
	// now we add a subscription to the chat.room
	// Subscribe to the subject
	uc.chatRoomSubscription, err = uc.nc.Subscribe("chat.room", func(msg *nats.Msg) {
		// Print message data
		ctx.Dispatch(func(ctx app.Context) {
			if len(uc.messages) < 10 {
				uc.messages = append(uc.messages, string(msg.Data))
			} else {
				uc.messages = append(uc.messages[1:], string(msg.Data))
			}
		})
	})

	// grab an avatar
	msg, err := uc.nc.Request("govatar.female", []byte(""), 200*time.Millisecond)
	if err != nil {
		app.Logf("govatar request error %s", err)
	} else {
		ctx.Dispatch(func(ctx app.Context) {
			// set it
			uc.avatar = string(msg.Data)
		})
	}

	// tell them we are here
	err = uc.nc.Publish("chat.say", []byte(uc.whoAmi+" entered the room"))
	if err != nil {
		app.Logf("Publish entry message error %s", err)
	}

	// create an echo service in this browser :)
	uc.echoSubscription, err = uc.nc.Subscribe("echo."+uc.whoAmi, func(msg *nats.Msg) {
		_ = msg.Respond(msg.Data)
		ctx.Dispatch(func(ctx app.Context) {
			uc.echoCount++
		})
	})

	uc.Update()

	app.Window().GetElementByID("inp").Call("focus")

}

func (uc *Front) Render() app.UI {

	return app.Div().Body(
		app.H1().Text("Go-Nats-App"),
		app.Div().Text(func() string {
			if uc.nc == nil {
				return "Not connected to the nats server"
			} else {
				return "Connected to: " + uc.nc.ConnectedServerName()
			}
		}()),
		app.Div().Body(app.Img().Src(uc.avatar).Width(250).Height(250)),
		app.H4().Text("Chat:"),
		app.Form().Body(
			app.Div().Body(
				app.Span().Text(uc.whoAmi+": "),
				app.Input().Value(uc.input).ID("inp").OnInput(uc.ValueTo(&uc.input)),
			),
		).OnSubmit(func(ctx app.Context, e app.Event) {
			e.PreventDefault()
			if uc.nc != nil {
				err := uc.nc.Publish("chat.say", []byte(uc.whoAmi+": "+uc.input))
				if err != nil {
					app.Logf("Publish error %s", err)
				}
				uc.input = "" // clear the message entry
			}
		}),
		app.Range(uc.messages).Slice(func(i int) app.UI {
			return app.Div().Text(uc.messages[len(uc.messages)-1-i])
		}),
		app.H4().Text("Trigger echo with http request"),
		app.A().Href("").Text("Echos sent: "+strconv.Itoa(uc.echoCount)).OnClick(uc.echoClick),

		app.H4().Text("Connection tests"),
		app.Button().Text("disconnect").OnClick(func(ctx app.Context, e app.Event) {
			uc.disconnect(ctx)
		}),
		app.Button().Text("reconnect").OnClick(func(ctx app.Context, e app.Event) {
			go uc.reconnect(ctx)
		}),
	)
}

func (uc *Front) echoClick(ctx app.Context, e app.Event) {
	echoUrl := fmt.Sprintf("%s/echo/%s", goapp.Base(), uc.whoAmi)
	app.Log("get to", echoUrl)
	resp, _ := http.Get(echoUrl)
	_, _ = io.ReadAll(resp.Body)
}
