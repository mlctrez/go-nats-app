package goapp

import (
	"fmt"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"strings"
)

// Base returns the base url of the application without any trailing slash.
//
//	Note: returns an empty string when app.IsServer == true.
func Base() string {
	if app.IsServer {
		return ""
	}

	location := app.Window().Get("location")

	protocol := location.Get("protocol").String()
	host := location.Get("host").String()

	return fmt.Sprintf("%s//%s", protocol, host)
}

// BaseWs is the same as Base but transforms http -> ws and https -> wss.
func BaseWs() string {
	return "ws" + strings.TrimPrefix(Base(), "http")
}
