package main

import (
	"context"
	"flag"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"

	"github.com/gerifield/epher/epher"
)

func main() {
	listen := flag.String("listen", ":9090", "HTTP and WS server listen address")
	redisAddr := flag.String("redisAddr", "", "Redis server address to distribute messages (optional)")
	redisPass := flag.String("redisPass", "", "Redis server password (optional)")
	flag.Parse()
	r := chi.NewRouter()

	var redisClient *redis.Client
	if redisAddr != nil && *redisAddr != "" {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     *redisAddr,
			Password: *redisPass,
			DB:       0, // use default DB
		})

		status := redisClient.Ping(context.Background())
		if status.Err() != nil {
			log.Println("redis connection failed:", status.Err())

			return
		}
	}

	e := epher.New(redisClient)

	r.Get("/subscribe/{room:[a-zA-Z0-9]+}", e.WebsocketHandler) // Websocket
	r.Post("/publish/{room:[a-zA-Z0-9]+}", e.PushHandler)       // Http

	r.Handle("/metrics", promhttp.Handler()) // Prometheus metrics

	log.Printf("Started on %s\n", *listen)
	err := http.ListenAndServe(*listen, r)
	if err != nil {
		log.Fatalln(err)
	}
}
