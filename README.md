# go-nats-app

A spin-off of [go-nats-app](https://github.com/oderwat/go-nats-app/). See the readme there for full details.

Credit to [Hans](https://github.com/oderwat/) for the original and sending me on this journey.

Most of the original code is in [setupNatsApi](api/api.go#L41) for the server
and [front.go](goapp/compo/front/front.go) for the client. Nats server setup is
in [service.go](goapp/service/service.go#L222).

### Changes in this spin-off

In the original code, the nats client in the wasm connects to the nats server
using a [Nats Websocket](https://docs.nats.io/running-a-nats-service/configuration/websocket) via
a [CustomDialer](https://github.com/nats-io/nats.go/blob/6c6add8d63597f84bee75d37bb1520e01552a02d/nats.go#L252).

`Go-App -> Nats Client -> <Nats WebSocket> -> WebSocket Endpoint`

In the (now updated) spin-off, the nats client in the wasm connects using [goapp-natsws](https://github.com/mlctrez/goapp-natsws).

`Go-App -> Nats Client -> <natsws.Connection> -> <natsws.Proxy> -> WebSocket Endpoint`

This allows the nats communication to be on the same host:port where the go-app is served.

Incrementing the echo counter in the original was replaced with a client http get to `/echo/username`.

Interestingly, this led to the discovery of the nats client connection option `nats.InProcessServer`.

A disconnect test was added to simulate the backend disconnecting the websocket.

### Disclaimers

* See readme in [goapp-natsws](https://github.com/mlctrez/goapp-natsws)
* As pointed out, this seems like a bit like re-inventing the wheel.
    * This is not intended to replace the built-in nats websocket client with all of its fail over / cluster
      capabilities. 
    * Use the default nats client if that's needed.

[![Go Report Card](https://goreportcard.com/badge/github.com/mlctrez/go-nats-app)](https://goreportcard.com/report/github.com/mlctrez/go-nats-app)

created by [tigwen](https://github.com/mlctrez/tigwen)
