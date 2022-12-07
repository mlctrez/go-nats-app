# go-nats-app

A spin-off of [go-nats-app](https://github.com/oderwat/go-nats-app/). See the readme there for full details.

Credit to [Hans](https://github.com/oderwat/) for the original and sending me on this journey.

Most of the original code is in [setupNatsApi](api/api.go#:~:text=setupNatsApiDocLink) for the server
and [front.go](goapp/compo/front/front.go) for the client. Nats server setup is
in [service.go](goapp/service/service.go#L222).

### Changes in this spin-off

In the original code, the nats client in the wasm connects to the nats server
using a [Nats Websocket](https://docs.nats.io/running-a-nats-service/configuration/websocket) via
a [CustomDialer](https://github.com/nats-io/nats.go/blob/6c6add8d63597f84bee75d37bb1520e01552a02d/nats.go#L252).

`Go-App -> Nats Client -> <Nats WebSocket> -> WebSocket Endpoint`

In the spin-off, the nats client in the wasm connects using
a [CustomDialer](https://github.com/nats-io/nats.go/blob/6c6add8d63597f84bee75d37bb1520e01552a02d/nats.go#L252) again,
but the connection is to a [WebSocket Handler](api/api.go#L94).

This handler adapts the websocket binary payloads to the standard Nats tcp
connection [Client Protocol](https://docs.nats.io/reference/reference-protocols/nats-protocol).

`Go-App -> Nats Client -> <Nats WebSocket> -> <WebSocket Handler> -> Nats Endpoint`

This allows publishing the nats client endpoint as a websocket.
The client nats connection and underlying websocket are managed as a go-app [component](goapp/compo/nats).

Incrementing the echo counter in the original was replaced with a client http get to `/echo/username`.
Interestingly, this led to the discovery of the nats client connection option `nats.InProcessServer`. Neat.

### Disclaimers

* Not fully tested, not production ready, no warranty, no support, etc.
* The note [here](https://docs.nats.io/running-a-nats-service/configuration/websocket) for writers of client libraries:
    * "Any data from a frame must be going through a parser that can handle partial protocols".
    * In this small demo, no issues were encountered. Maybe it is not applicable.
* [WebSocket Handler](api/api.go#L94) definitely needs to cover a few more error / connect / reconnect / disconnect
  cases.
 

[![Go Report Card](https://goreportcard.com/badge/github.com/mlctrez/go-nats-app)](https://goreportcard.com/report/github.com/mlctrez/go-nats-app)

created by [tigwen](https://github.com/mlctrez/tigwen)
