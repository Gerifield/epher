# Websocket to HTTP callback

This part will push a message from a websocket receiver to a HTTP endpoint.


With the other tool you can "chain" some calls like this:

Start the websocket server from the root dir:
```
$ go run cmd/http-to-ws/http-to-ws.go
```

Start a websocket listener and connect it to a room:
```
$ websocat ws://127.0.0.1:9090/subscribe/test2
```

Start this proxy between the websocket receiver and the proxy's publish endpoint (where the websocat will listen):
```
$ go run cmd/ws-to-http/ws-to-http.go
```

Config file for this example:
```
{
  "routes":[
    {
    "ws_address": "ws://127.0.0.1:9090/subscribe/test1",
    "http_url": "http://127.0.0.1:9090/publish/test2",
    "forward_response": true
    }
  ]
}

```

Now just send a message to the websocket proxy and you should receive it in the websocat listener:
```
$ curl http://127.0.0.1:9090/publish/test1 -d 'some payload'
```
