package compo

import (
	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/mlctrez/go-nats-app/goapp/compo/front"
	"github.com/mlctrez/go-nats-app/goapp/compo/nats"
)

var _ app.AppUpdater = (*Root)(nil)

type Root struct {
	app.Compo
}

func (r *Root) Render() app.UI {
	return app.Div().Body(
		//app.Div().Text(app.Getenv("GOAPP_VERSION")),
		&nats.Nats{},
		&front.Front{},
	)
}

func (r *Root) OnAppUpdate(ctx app.Context) {
	if app.Getenv("DEV") != "" && ctx.AppUpdateAvailable() {
		ctx.Reload()
	}
}
