package main

import (
	"log"
	"net/http"

	"github.com/gerifield/epher/epher"
	"github.com/go-chi/chi"
)

func main() {
	r := chi.NewRouter()
	e := epher.New()

	r.Get("/subscribe/{room:[a-zA-Z0-9]+}", e.WebsocketHandler) // Websocket
	r.Post("/publish/{room:[a-zA-Z0-9]+}", e.PushHandler)       // Http

	log.Println("Started")
	err := http.ListenAndServe("127.0.0.1:9090", r)
	if err != nil {
		log.Fatalln(err)
	}
}
