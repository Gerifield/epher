# Event Pusher

This stuff allows users to subscribe to a given channel and let's services to send notifications to those channels with an HTTP request.


# Usage

Start:

`go run main.go`

Connect with websocket to: `http://127.0.0.1:9090/subscribe/test1`

Send HTTP post requests to: `http://127.0.0.1:9090/publish/test1` 