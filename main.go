package main

import (
	"log"
	"net/http"

	"github.com/gerifield/epher/epher"
	"github.com/gorilla/mux"
)

func main() {

	liveMux := mux.NewRouter()
	privMux := mux.NewRouter()

	e := epher.NewEpher()

	liveMux.HandleFunc("/subscribe/{room:[a-zA-Z0-9]+}", e.WebsocketHandler).Methods("GET")
	privMux.HandleFunc("/publish/{room:[a-zA-Z0-9]+}", e.PushHandler).Methods("POST")

	log.Println("Started")
	go http.ListenAndServe("127.0.0.1:8080", liveMux) // Websocket
	http.ListenAndServe("127.0.0.1:9090", privMux)    // Http

}
