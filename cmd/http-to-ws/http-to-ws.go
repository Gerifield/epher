package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/gerifield/epher/epher"
	"github.com/go-chi/chi"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	listen := flag.String("listen", ":9090", "HTTP and WS server listen address")
	flag.Parse()
	r := chi.NewRouter()
	e := epher.New()
	e.RegisterMetrics()

	r.Get("/subscribe/{room:[a-zA-Z0-9]+}", e.WebsocketHandler) // Websocket
	r.Post("/publish/{room:[a-zA-Z0-9]+}", e.PushHandler)       // Http

	r.Handle("/metrics", promhttp.Handler()) // Prometheus metrics

	log.Printf("Started on %s\n", *listen)
	err := http.ListenAndServe(*listen, r)
	if err != nil {
		log.Fatalln(err)
	}
}
