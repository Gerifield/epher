# Event Pusher ![Go](https://github.com/Gerifield/epher/workflows/Go/badge.svg?branch=master) [![Coverage Status](https://coveralls.io/repos/github/Gerifield/epher/badge.svg?branch=master)](https://coveralls.io/github/Gerifield/epher?branch=master)

This stuff allows users to subscribe to a given channel and let the services to send notifications to those channels with an HTTP request.
All the messages sent to the websocket endpoint will be dropped.

# Usage

```
  -listen string
    	HTTP and WS server listen address (default ":9090")
```

# Example

Start the server:

```
$ go run main.go
```
It'll start on port `9090` by default.

Connect with websocket to the subscribe endpoint (for example using websocat):
```
$ websocat ws://127.0.0.1:9090/subscribe/test1
```

Send HTTP post requests to the publish endpoint:
```
$ curl 127.0.0.1:9090/publish/test1 -d 'test message'
```
You should see your message in the websocket client.
