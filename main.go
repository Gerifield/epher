package main

import (
	"log"
	"net/http"

	"github.com/gerifield/epher/epher"
	"github.com/gorilla/mux"
)

func main() {

	r := mux.NewRouter()
	e := epher.NewEpher()

	r.HandleFunc("/subscribe/{room:[a-zA-Z0-9]+}", e.WebsocketHandler).Methods("GET") // Websocket
	r.HandleFunc("/publish/{room:[a-zA-Z0-9]+}", e.PushHandler).Methods("POST")       // Http

	log.Println("Started")
	http.ListenAndServe("127.0.0.1:9090", r)

}
